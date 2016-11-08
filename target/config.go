package target

import "os"

// Config is the config settings for Target
type Config struct {
	// <addr>:<port> to bind target
	BindAddress string
	// if print log to console
	PrintLog bool
	// signal channel
	Sighup chan os.Signal
}
