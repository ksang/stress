package target

import "os"

// Config is the config settings for stress target
type Config struct {
	// <addr>:<port> to bind target
	BindAddress string
	// if print log periodically
	PrintLog bool
	// signal channel for SIGHUP
	Sighup chan os.Signal
}
