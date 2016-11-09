package target

import "os"

// Config is the config settings for stress target
type Config struct {
	// <addr>:<port> to bind target
	BindAddress string
	// if print log to console
	PrintLog bool
	// signal channel to handle SIGHUP
	Sighup chan os.Signal
}
