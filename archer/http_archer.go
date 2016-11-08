/*
package archer provides the stress client for performance testing
*/
package archer

import (
	"log"
	"net/url"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/valyala/fasthttp"
)

type archerStats struct {
	sentBytes     uint64
	receivedBytes uint64
	succeeded     uint64
	failed        uint64
}

type httpArcher struct {
	stats    archerStats
	printErr bool
	target   string
	host     string
	interval time.Duration
	connNum  int
	num      uint64
	data     []byte
	sighup   chan os.Signal
}

func (h *httpArcher) Launch() error {
	var (
		count uint64
		wg    sync.WaitGroup
	)
	client := &fasthttp.Client{
		MaxConnsPerHost:               h.connNum,
		DisableHeaderNamesNormalizing: true,
	}

	for i := 0; i < h.connNum; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			req := &fasthttp.Request{}
			req.SetBody(h.data)
			req.SetHost(h.host)
			req.SetRequestURI(h.target)
			req.Header.SetMethod("PUT")
			req.Header.SetContentLength(len(h.data))
			size := req.Header.ContentLength() + req.Header.Len()

			res := &fasthttp.Response{}
			for {
				time.Sleep(h.interval)
				// total number finished
				if h.num > 0 {
					if atomic.LoadUint64(&count) >= h.num {
						return
					}
					atomic.AddUint64(&count, 1)
				}

				if err := client.Do(req, res); err != nil {
					if h.printErr {
						log.Printf("client DO err: %s", err)
					}
					atomic.AddUint64(&h.stats.failed, 1)
					continue
				}
				atomic.AddUint64(&h.stats.succeeded, 1)
				atomic.AddUint64(&h.stats.sentBytes, uint64(size))
				atomic.AddUint64(&h.stats.receivedBytes, uint64(res.Header.Len()))
				atomic.AddUint64(&h.stats.receivedBytes, uint64(res.Header.ContentLength()))
			}
		}()
	}
	wg.Wait()
	return nil
}

func (h *httpArcher) SentBytes() uint64 {
	return atomic.LoadUint64(&h.stats.sentBytes)
}

func (h *httpArcher) ReceivedBytes() uint64 {
	return atomic.LoadUint64(&h.stats.receivedBytes)
}

func (h *httpArcher) Succeeded() uint64 {
	return atomic.LoadUint64(&h.stats.succeeded)
}

func (h *httpArcher) Failed() uint64 {
	return atomic.LoadUint64(&h.stats.failed)
}

func StartHTTPArcher(cfg Config) error {
	u, err := url.Parse(cfg.Target)
	if err != nil {
		return err
	}
	interval, err := time.ParseDuration(cfg.Interval)
	if err != nil {
		return err
	}
	archer := &httpArcher{
		target:   cfg.Target,
		host:     u.Host,
		interval: interval,
		connNum:  cfg.ConnNum,
		data:     cfg.Data,
		num:      cfg.Num,
		sighup:   cfg.Sighup,
	}
	go archer.PrintStats(cfg.PrintLog)
	return archer.Launch()
}

func (h *httpArcher) PrintStats(periodic bool) {
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

func (h *httpArcher) PrintStatsOnce() {
	log.Printf("Sent Bytes: %v, Received Bytes: %v, Succeeded: %v, Failed: %v",
		h.SentBytes(), h.ReceivedBytes(), h.Succeeded(), h.Failed())
}
