package archer

import "os"

// Config is the config settings for stress archer
type Config struct {
	// target string
	Target string
	// interval duration in string
	Interval string
	// connection number
	ConnNum int
	// data
	Data []byte
	// if print log
	PrintLog bool
	// if print client errors
	PrintError bool
	// total number, 0 means non-stop
	Num uint64
	// signal channel for SIGHUP
	Sighup chan os.Signal
}
