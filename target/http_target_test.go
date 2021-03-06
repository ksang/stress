package target

import (
	"io"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"github.com/valyala/fasthttp"

	"github.com/ksang/stress/etcd/server"
)

func RunHTTPTarget(cfg Config) (*httpTarget, error) {
	sLn, err := StatsListen(cfg.BindAddress)
	if err != nil {
		return nil, err
	}
	target := &httpTarget{
		ln: sLn,
	}
	server := fasthttp.Server{
		Handler:            target.HandleFastHTTP,
		MaxRequestBodySize: 999999999,
	}
	if cfg.PrintLog {
		go target.PrintStats(true)
	}
	if cfg.EnableEtcd {
		go target.StartEtcdServer(cfg.Etcd)
	}
	go server.Serve(target.ln)
	return target, nil
}

func TestConnection(t *testing.T) {
	cfg := Config{
		BindAddress: "0.0.0.0:8888",
		PrintLog:    true,
	}

	target, err := RunHTTPTarget(cfg)
	if err != nil {
		t.Errorf("failed to start target: %s", err)
	}
	time.Sleep(1 * time.Second)

	// use not keep alive request as by default http client will keep conntection
	transport := http.Transport{
		DisableKeepAlives: true,
	}

	client := http.Client{
		Transport: &transport,
	}
	req, err := http.NewRequest("GET", "http://127.0.0.1:8888", nil)
	if err != nil {
		t.Errorf("%s", err)
	}

	res, err := client.Do(req)
	if err != nil {
		t.Errorf("failed to connect, %s", err)
	}
	// check ConnNum
	if target.ConnNumber() != 1 {
		t.Errorf("ConnNumber incorrect, not 1, actual: %v", target.ConnNumber())
	}
	// close connection
	io.Copy(ioutil.Discard, res.Body)
	res.Body.Close()
	time.Sleep(1 * time.Second)

	// check ConnNum again
	if target.ConnNumber() != 0 {
		t.Errorf("ConnNumber incorrect, not 0, actual: %v", target.ConnNumber())
	}
	target.Close()
}

func TestStartEtcd(t *testing.T) {
	etcdCfg := server.Config{
		Name:           "stress0",
		InitialCluster: "stress0=http://localhost:2380",
	}
	cfg := Config{
		BindAddress: "0.0.0.0:8889",
		PrintLog:    true,
		EnableEtcd:  true,
		Etcd:        etcdCfg,
	}
	target, err := RunHTTPTarget(cfg)
	if err != nil {
		t.Errorf("failed to start target: %s", err)
	}
	time.Sleep(10 * time.Second)
	target.Close()
}
