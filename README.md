# stress
Network performance test tool based on HTTP, it is based on [fasthttp](https://github.com/valyala/fasthttp).

# build

	make

Above command will create a single binary in `build` folder, the binary is used for both *Target* (server) and *Archer* (client) functionality.

# usage

### archer

	archer [-lev] [-c] <ConnNum> [-n] <Num> [-i] <duration> [-u] <data>
	       -t <url>:
	  run stress in archer mode, acting as http client.
	  -c int
	    	connection number (default 10)
	  -e	print client error
	  -i string
	    	archer mode: remote target url (default "100ms")
	  -l	print stat log to stdout
	  -n uint
	    	total number of requests to send, 0 means non-stop
	  -t string
	    	archer mode: remote target url
	  -u string
	    	data to send, it will try to open file first, if failed will use the string provided.
	  -v	print log + print client error

### target

	target [-l] [-bind] <address:port>:
	  run stress in target mode, acting as http server.
	  -bind string
	    	target mode: local addr to bind (default "0.0.0.0:8080")
	  -l	print stat log to stdout
