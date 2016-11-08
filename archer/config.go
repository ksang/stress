package archer

// Config is the config settings for Target
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
}
