# Define parameters
BINARY=stress
SHELL := /bin/bash
GOPACKAGES = $(shell go list ./... | grep -v vendor)
ROOTDIR = $(pwd)

.PHONY: build env install test linux get-deps

default: build

build: main.go 
	go build -v -o ./build/${BINARY} main.go

env: 
	export GOPATH=${GOPATH}

install:
	go install  ./...

test:
	go test -race -cover ${GOPACKAGES}

clean:
	rm -rf build

linux: main.go
	GOOS=linux GOARCH=amd64 go build -o ./build/linux/${BINARY} main.go
	
