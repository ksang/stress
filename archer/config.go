package archer

import "os"

// Config is the config settings for stress archer
type Config struct {
	// target url
	Target string
	// interval duration
	Interval string
	// connection number
	ConnNum int
	// data
	Data []byte
	// if print log periodically
	PrintLog bool
	// if print client errors
	PrintError bool
	// total number, 0 means non-stop
	Num uint64
	// signal channel for SIGHUP
	Sighup chan os.Signal
}
