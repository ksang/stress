package server

import (
	"testing"
	"time"

	"github.com/coreos/etcd/embed"
)

func TestNewConfig(t *testing.T) {
	cfg := embed.NewConfig()
	t.Logf("%#v", *cfg)
}

func TestStartServer(t *testing.T) {
	cfg := Config{}
	etcd, err := StartEmbedServer(cfg)
	if err != nil {
		t.Errorf("%s", err)
		return
	}
	defer etcd.Close()
	select {
	case <-etcd.Server.ReadyNotify():
		t.Logf("etcd Server is ready!")
	case <-time.After(60 * time.Second):
		etcd.Server.Stop() // trigger a shutdown
		t.Errorf("Server took too long to start!")
	}

	select {
	case err := <-etcd.Err():
		t.Errorf("etcd server error: %s", err)
	case <-time.After(10 * time.Second):
		t.Logf("etcd start server ok")
	}
	etcd.Close()
}
