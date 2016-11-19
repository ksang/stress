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
		t.Logf("Server took too long to start!")
	}
	t.Fatal(<-etcd.Err())
}
