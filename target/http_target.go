/*
Package target provides a server to provide stress test result.
*/
package target

import (
	"log"
	"net"
	"sync/atomic"
	"time"

	"github.com/valyala/fasthttp"
)

type httpStats struct {
	requestCount  uint64
	receivedBytes uint64
}

// HttpTarget is the http server target for network performance test
// it records stats when serving http requests
type HttpTarget struct {
	stats httpStats
	ln    *StatsListener
}

// StatsListener records listener related stats including connection number
type StatsListener struct {
	net.Listener
	ConnNumber uint64
}

// StatsListen returns StatsListener which records connection number and other
// stats.
func StatsListen(addr string) (*StatsListener, error) {
	ln, err := net.Listen("tcp4", addr)
	if err != nil {
		return nil, err
	}
	return &StatsListener{ln, 0}, nil
}

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

func (c *statsListenConn) Close() error {
	err := c.Conn.Close()
	atomic.AddUint64(&c.StatsListener.ConnNumber, ^uint64(0))
	return err
}

// HandleFastHTTP is request handler for http target, records stats for performance testing.
func (h *HttpTarget) HandleFastHTTP(ctx *fasthttp.RequestCtx) {
	size := h.RequestSize(ctx)
	atomic.AddUint64(&h.stats.receivedBytes, size)
	atomic.AddUint64(&h.stats.requestCount, 1)
	ctx.SuccessString("Text", "Stress target OK")
}

// RequestSize calculated the size of this http request.
func (h *HttpTarget) RequestSize(ctx *fasthttp.RequestCtx) uint64 {
	var ret uint64
	head := ctx.Request.Header
	head.VisitAll(func(k, v []byte) {
		ret += uint64(len(k))
		ret += uint64(len(v))
	})
	ret += uint64(head.ContentLength())
	return ret
}

// ConnNumber returns the current connection number of http target
func (h *HttpTarget) ConnNumber() uint64 {
	return atomic.LoadUint64(&h.ln.ConnNumber)
}

// ReceivedBytes returns the total bytes received of http target
func (h *HttpTarget) ReceivedBytes() uint64 {
	return atomic.LoadUint64(&h.stats.receivedBytes)
}

// ConnNumber returns the current connection number of http target
func (h *HttpTarget) RequestCount() uint64 {
	return atomic.LoadUint64(&h.stats.requestCount)
}

func StartHTTPTarget(cfg Config) error {
	sLn, err := StatsListen(cfg.BindAddress)
	if err != nil {
		log.Fatal("failed to bind: ", err)
	}
	target := &HttpTarget{
		ln: sLn,
	}
	server := fasthttp.Server{
		Handler:                       target.HandleFastHTTP,
		MaxRequestBodySize:            65536 * 32768,
		Concurrency:                   65536 * 32768,
		DisableHeaderNamesNormalizing: true,
	}
	if cfg.PrintLog {
		go target.PrintStats(5 * time.Second)
		defer target.PrintStatsOnce()
	}
	log.Printf("HTTP Target serving at: %s", cfg.BindAddress)
	return server.Serve(target.ln)
}

func (h *HttpTarget) PrintStats(i time.Duration) {
	for {
		select {
		case <-time.After(i):
			h.PrintStatsOnce()
		}
	}
}

func (h *HttpTarget) PrintStatsOnce() {
	log.Printf("ConnNum: %v, Received Bytes: %v, Request Count: %v",
		h.ConnNumber(), h.ReceivedBytes(), h.RequestCount())
}
