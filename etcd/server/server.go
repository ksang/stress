/*
package server provides embed etcd server in the process.
In stress we use etcd server to store stats across all target servers
to get over all performance stats. In this package we minimize etcd server
configuration for minimal usage.
*/
package server

import (
	"log"
	"net/url"
	"os"
	"time"

	"github.com/coreos/etcd/embed"
)

// Config is the configuration interface of etcd embed server
// used by stress target
type Config struct {
	ListenPeerURLs      []url.URL
	ListenClientURLs    []url.URL
	AdvertisePeerURLs   []url.URL
	AdvertiseClientURLs []url.URL
	Name                string
	InitialCluster      string
	InitialClusterToken string
}

// Start etcd embed server
func StartEmbedServer(cfg Config) (*embed.Etcd, error) {
	inCfg := embed.NewConfig()
	mergeConfig(inCfg, cfg)
	if err := inCfg.Validate(); err != nil {
		return nil, err
	}
	return embed.StartEtcd(inCfg)
}

// StartAndServe the etcd server permanently
func StartAndServe(cfg Config) {
	etcd, err := StartEmbedServer(cfg)
	if err != nil {
		log.Fatal("Failed to start etcd server: %s", err)
	}
	defer etcd.Close()
	select {
	case <-etcd.Server.ReadyNotify():
		log.Println("etcd server is ready!")
	case <-time.After(60 * time.Second):
		etcd.Server.Stop() // trigger a shutdown
		log.Println("etcd server took too long to start, stopping")
	}
	log.Fatal(<-etcd.Err())
}

func mergeConfig(c *embed.Config, cfg Config) {
	hostname, err := os.Hostname()
	if err != nil {
		log.Println("Failed to get hostname:", err)
	}
	c.Name = hostname
	if len(cfg.ListenPeerURLs) > 0 {
		c.LPUrls = cfg.ListenPeerURLs
	}
	if len(cfg.ListenClientURLs) > 0 {
		c.LCUrls = cfg.ListenClientURLs
	}
	if len(cfg.AdvertisePeerURLs) > 0 {
		c.APUrls = cfg.ListenPeerURLs
	}
	if len(cfg.AdvertiseClientURLs) > 0 {
		c.ACUrls = cfg.AdvertiseClientURLs
	}
	if len(cfg.Name) > 0 {
		c.Name = cfg.Name
	}
	if len(cfg.InitialCluster) > 0 {
		c.InitialCluster = cfg.InitialCluster
	} else {
		c.InitialCluster = c.InitialClusterFromName(c.Name)
	}
	if len(cfg.InitialClusterToken) > 0 {
		c.InitialClusterToken = cfg.InitialClusterToken
	}
	c.Dir = "etcd_data_" + c.Name
}
