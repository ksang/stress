/*
package target provides the stress server for performance testing
*/
package target

import (
	"log"
	"net"
	"os"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/embed"
	"github.com/valyala/fasthttp"
	"golang.org/x/net/context"

	"github.com/ksang/stress/etcd/server"
	"github.com/ksang/stress/util"
)

type httpStats struct {
	requestCount  uint64
	receivedBytes uint64
}

type httpTarget struct {
	stats  httpStats
	ln     *StatsListener
	sighup chan os.Signal
	etcd   *embed.Etcd
}

// StatsListener records listener related stats including connection number
// it is a wrapper of net.Listener
type StatsListener struct {
	net.Listener
	ConnNumber uint64
}

// StatsListen creats StatsListener internally creates a tcp4 net.Listener
func StatsListen(addr string) (*StatsListener, error) {
	ln, err := net.Listen("tcp4", addr)
	if err != nil {
		return nil, err
	}
	return &StatsListener{ln, 0}, nil
}

// Accept wraps origin Accept method and records connection number
func (l *StatsListener) Accept() (net.Conn, error) {
	c, err := l.Listener.Accept()
	if err != nil {
		return nil, err
	}
	atomic.AddUint64(&l.ConnNumber, 1)
	return &statsListenConn{l, c}, nil
}

type statsListenConn struct {
	*StatsListener
	net.Conn
}

// Accept wraps origin Close method and records connection number
func (c *statsListenConn) Close() error {
	err := c.Conn.Close()
	atomic.AddUint64(&c.StatsListener.ConnNumber, ^uint64(0))
	return err
}

// HandleFastHTTP is request handler for http target, records stats for performance testing.
func (h *httpTarget) HandleFastHTTP(ctx *fasthttp.RequestCtx) {
	size := h.RequestSize(ctx)
	atomic.AddUint64(&h.stats.receivedBytes, size)
	atomic.AddUint64(&h.stats.requestCount, 1)
	ctx.SuccessString("Text", "Stress target OK")
}

// RequestSize return the size of this http request.
func (h *httpTarget) RequestSize(ctx *fasthttp.RequestCtx) uint64 {
	var ret uint64
	head := ctx.Request.Header
	head.VisitAll(func(k, v []byte) {
		ret += uint64(len(k))
		ret += uint64(len(v))
	})
	ret += uint64(head.ContentLength())
	return ret
}

func (h *httpTarget) ConnNumber() uint64 {
	return atomic.LoadUint64(&h.ln.ConnNumber)
}

func (h *httpTarget) ReceivedBytes() uint64 {
	return atomic.LoadUint64(&h.stats.receivedBytes)
}

func (h *httpTarget) RequestCount() uint64 {
	return atomic.LoadUint64(&h.stats.requestCount)
}

func (h *httpTarget) Close() {
	if h.etcd != nil {
		h.etcd.Close()
	}
}

// Start HTTP target by providing target configurations
func StartHTTPTarget(cfg Config) error {
	sLn, err := StatsListen(cfg.BindAddress)
	if err != nil {
		log.Fatal("failed to bind: ", err)
	}
	target := &httpTarget{
		ln:     sLn,
		sighup: cfg.Sighup,
	}
	server := fasthttp.Server{
		Handler:                       target.HandleFastHTTP,
		MaxRequestBodySize:            65536 * 32768,
		Concurrency:                   65536 * 32768,
		DisableHeaderNamesNormalizing: true,
	}
	go target.PrintStats(cfg.PrintLog)

	if cfg.EnableEtcd {
		go target.StartEtcdServer(cfg.Etcd)
	}

	log.Printf("HTTP Target serving at: %s", cfg.BindAddress)
	return server.Serve(target.ln)
}

func (h *httpTarget) PrintStats(periodic bool) {
	c := make(chan struct{}, 1)
	if periodic {
		defer h.PrintStatsOnce()
		go func() {
			for {
				select {
				case <-time.After(5 * time.Second):
					c <- struct{}{}
				}
			}
		}()
	}

	go func() {
		for {
			select {
			case <-c:
				h.PrintStatsOnce()
			case <-h.sighup:
				h.PrintStatsOnce()
			}
		}
	}()
}

func (h *httpTarget) PrintStatsOnce() {
	log.Printf("ConnNum: %v, Received Bytes: %v, Request Count: %v",
		h.ConnNumber(), h.ReceivedBytes(), h.RequestCount())
}

func (h *httpTarget) StartEtcdServer(cfg server.Config) {
	cfg.InitialClusterToken = StressClusterToken
	etcd, err := server.StartEmbedServer(cfg)
	if err != nil {
		log.Printf("failed to start etcd server: %v", err)
		return
	}
	h.etcd = etcd
	defer etcd.Close()
	select {
	case <-etcd.Server.ReadyNotify():
		log.Printf("etcd Server is ready")
		// start updating etcd stats
		go h.UpdateEtcdStats()
	case <-time.After(60 * time.Second):
		etcd.Server.Stop() // trigger a shutdown
		log.Printf("etcd server took too long to start")
	}
	log.Printf("etcd server abnormal terminate: %v", <-etcd.Err())
}

func (h *httpTarget) UpdateEtcdStats() {
	endpoints := util.ParseUrlsToStrings(h.etcd.Config().LCUrls)
	log.Printf("etcd client got endpoints: %v", endpoints)
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		log.Printf("failed to create etcd client: %v", err)
		return
	}
	defer cli.Close()
	kv := clientv3.NewKV(cli)
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	for {
		select {
		case <-time.After(1 * time.Second):
			h.UpdateEtcdStatsOnce(ctx, kv)
		}
	}
}

func (h *httpTarget) UpdateEtcdStatsOnce(ctx context.Context, kv clientv3.KV) {
	rb := strconv.FormatUint(h.ReceivedBytes(), 10)
	rc := strconv.FormatUint(h.RequestCount(), 10)
	cn := strconv.FormatUint(h.ConnNumber(), 10)
	name := h.etcd.Config().Name

	ops := []clientv3.Op{
		clientv3.OpPut(util.ReceivedBytesKey+"/"+name, rb),
		clientv3.OpPut(util.RequestCountKey+"/"+name, rc),
		clientv3.OpPut(util.ConnNumberKey+"/"+name, cn)}

	for _, op := range ops {
		if _, err := kv.Do(context.TODO(), op); err != nil {
			log.Printf("failed to update etcd RequestCount: %v", err)
		}
	}
}
