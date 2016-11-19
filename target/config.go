package target

import (
	"os"

	"github.com/ksang/stress/etcd/server"
)

var (
	StressClusterToken = "stress_etcd_cluster"
)

// Config is the config settings for stress target
type Config struct {
	// <addr>:<port> to bind target
	BindAddress string
	// if print log periodically
	PrintLog bool
	// signal channel for SIGHUP
	Sighup chan os.Signal
	// if enable etcd
	EnableEtcd bool
	// etcd server config
	Etcd server.Config
}
