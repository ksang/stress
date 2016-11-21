# stress
Network performance test tool, powered by [fasthttp](https://github.com/valyala/fasthttp) and [etcd](https://github.com/coreos/etcd).

[![Build Status](https://travis-ci.org/ksang/stress.svg?branch=master)](https://travis-ci.org/ksang/stress) [![Go Report Card](https://goreportcard.com/badge/github.com/ksang/stress)](https://goreportcard.com/report/github.com/ksang/stress)

### build

	make

Above command will create a single binary in `build` folder, the binary is used for both *Target* (server) and *Archer* (client) functionality.

### usage

	./stress archer -h
	./stress target -h

### example

`$./stress archer -v -u stress -t http://127.0.0.1:8080`

Above command will launch archer client connecting to localhost sending data read from stress binary

`$./stress -proc 16 target -bind 0.0.0.0:8080`

Above command will listen on address 0.0.0.0:8080 with 16 GOMAXPROC

	Start first instance:

	$./stress target -bind 127.0.0.1:8080 \
					-name etcd0 \
					-peer http://127.0.0.1:4001 \
					-client http://127.0.0.1:4002 \
					-initial-cluster etcd0=http://127.0.0.1:4001,etcd1=http://127.0.0.1:5001

	Start second instance:

	$./stress target -bind 127.0.0.1:8081 \
					-name etcd1 \
					-peer http://127.0.0.1:5001 \
					-client http://127.0.0.1:5002 \
					-initial-cluster etcd0=http://127.0.0.1:4001,etcd1=http://127.0.0.1:5001

Above commands will run two stress instances with etcd clusering storing stats to etcd KV. To check stats, you can run `etcdctl` with etcd client api v3, below command is for example above:

	$ETCDCTL_API=3 etcdctl --endpoints http://127.0.0.1:4002,http://127.0.0.1:5002 get --prefix stress
	stress/ConnectionNumber/etcd0
	10
	stress/ConnectionNumber/etcd1
	10
	stress/ReceivedBytes/etcd0
	41280
	stress/ReceivedBytes/etcd1
	26880
	stress/RequestCount/etcd0
	430
	stress/RequestCount/etcd1
	280