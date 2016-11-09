# stress
Network performance test tool using HTTP, it is based on [fasthttp](https://github.com/valyala/fasthttp).

### build

	make

Above command will create a single binary in `build` folder, the binary is used for both *Target* (server) and *Archer* (client) functionality.

### usage

##### archer

	archer [-lev] [-c] <ConnNum> [-n] <Num> [-i] <duration> [-u] <data>
	       -t <url>:
	  run stress in archer mode, acting as http client.
	  -c int
	    	connection number (default 10)
	  -e	print client error
	  -i string
	    	archer mode: remote target url (default "100ms")
	  -l	print stat log to stdout periodically
	  -n uint
	    	total number of requests to send, 0 means non-stop
	  -t string
	    	archer mode: remote target url
	  -u string
	    	data to send, it will try to open file first, if failed will use the string provided.
	  -v	print log + print client error

`./stress archer -v -u stress -t 127.0.0.1:8080`

Above command will launch archer client connecting to localhost sending data read from stress binary

##### target

	target [-l] [-bind] <address:port>:
	  run stress in target mode, acting as http server.
	  -bind string
	    	target mode: local addr to bind (default "0.0.0.0:8080")
	  -l	print stat log to stdout periodically

`./stress -proc 16 target -bind 0.0.0.0:8080`

Above command will listen on address 0.0.0.0:8080 with 16 GOMAXPROC
