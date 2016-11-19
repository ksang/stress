/*
stress is network performance test tool using HTTP, it is based on fasthttp.
For usage please see command line help or README.md
*/
package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/google/subcommands"
	"golang.org/x/net/context"

	"github.com/ksang/stress/archer"
	"github.com/ksang/stress/etcd/server"
	"github.com/ksang/stress/target"
	"github.com/ksang/stress/util"
)

var (
	maxproc int
)

func init() {
	flag.IntVar(&maxproc, "proc", 4, "GOMAXPROC setting")
}

func main() {
	subcommands.Register(subcommands.HelpCommand(), "")
	subcommands.Register(subcommands.FlagsCommand(), "")
	subcommands.Register(subcommands.CommandsCommand(), "")
	subcommands.Register(&targetCmd{}, "")
	subcommands.Register(&archerCmd{}, "")

	flag.Parse()

	runtime.GOMAXPROCS(maxproc)

	ctx := context.Background()
	os.Exit(int(subcommands.Execute(ctx)))
}

type targetCmd struct {
	bindaddr string
	printlog bool
	//etcd related configuartions
	peerURLs       string
	clientURLs     string
	name           string
	initialCluster string
}

func (*targetCmd) Name() string     { return "target" }
func (*targetCmd) Synopsis() string { return "run as target (server) mode" }
func (*targetCmd) Usage() string {
	return `target [-l] [-bind] <address:port>:
  run stress in target mode, acting as http server.
`
}

func (t *targetCmd) SetFlags(f *flag.FlagSet) {
	f.StringVar(&t.bindaddr, "bind", "0.0.0.0:8080", "target mode: local addr to bind")
	f.BoolVar(&t.printlog, "l", false,
		"print stat log to stdout periodically")
	f.StringVar(&t.name, "name", "",
		"etcd node name, set this value to enable etcd")
	f.StringVar(&t.peerURLs, "peer", "",
		"etcd peer urls for advertise and listen, default is http://localhost:2380")
	f.StringVar(&t.clientURLs, "client", "",
		"etcd client urls for advertise and listen, default is http://localhost:2379")
	f.StringVar(&t.initialCluster, "initial-cluster", "",
		"etcd initial cluster string")
}

func (t *targetCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	// init signal
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGHUP)

	cfg := target.Config{
		BindAddress: t.bindaddr,
		PrintLog:    t.printlog,
		Sighup:      sig,
	}

	// init etcd configs
	etcdCfg := server.Config{}
	if len(t.name) > 0 {
		cfg.EnableEtcd = true
		etcdCfg.Name = t.name
		pu, err := util.ParseStringToUrl(t.peerURLs)
		if err != nil {
			log.Fatalf("Failed to parse peer url: %s", err)
		}
		cu, err := util.ParseStringToUrl(t.clientURLs)
		if err != nil {
			log.Fatalf("Failed to parse client url: %s", err)
		}
		etcdCfg.ListenClientURLs = cu
		etcdCfg.AdvertiseClientURLs = cu
		etcdCfg.ListenPeerURLs = pu
		etcdCfg.AdvertisePeerURLs = pu
		etcdCfg.InitialCluster = t.initialCluster
		cfg.Etcd = etcdCfg
	}

	log.Fatal(target.StartHTTPTarget(cfg))
	return subcommands.ExitSuccess
}

type archerCmd struct {
	target   string
	interval string
	data     string
	printlog bool
	printerr bool
	verbose  bool
	connnum  int
	num      uint64
}

func (*archerCmd) Name() string     { return "archer" }
func (*archerCmd) Synopsis() string { return "run as archer (client) mode" }
func (*archerCmd) Usage() string {
	return `archer [-lev] [-c] <ConnNum> [-n] <Num> [-i] <duration> [-u] <data> 
       -t <url>:
  run stress in archer mode, acting as http client.
`
}

func (a *archerCmd) SetFlags(f *flag.FlagSet) {
	f.StringVar(&a.target, "t", "", "archer mode: remote target url")
	f.StringVar(&a.interval, "i", "100ms", "archer mode: remote target url")
	f.StringVar(&a.data, "u", "",
		"data to send, it will try to open file first, if failed will use the string provided.")
	f.BoolVar(&a.printlog, "l", false,
		"print stat log to stdout  periodically")
	f.BoolVar(&a.printerr, "e", false, "print client error")
	f.BoolVar(&a.verbose, "v", false, "print log + print client error")
	f.IntVar(&a.connnum, "c", 10, "connection number")
	f.Uint64Var(&a.num, "n", 0, "total number of requests to send, 0 means non-stop")
}

func (a *archerCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	// check target
	if len(a.target) == 0 {
		fmt.Printf("Error: you must specify target url\n")
		f.PrintDefaults()
		return subcommands.ExitFailure
	}
	// verbose logging
	if a.verbose {
		a.printlog = true
		a.printerr = true
	}
	// init signal
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGHUP)
	// get input data
	var data []byte
	file, err := os.Open(a.data)
	if err != nil {
		data = []byte(a.data)
		goto CONFIG
	}
	defer file.Close()
	data, err = ioutil.ReadAll(file)
	if err != nil {
		log.Fatalf("Failed to read file: %s", err)
	}

CONFIG:
	cfg := archer.Config{
		Target:     a.target,
		Interval:   a.interval,
		ConnNum:    a.connnum,
		Data:       data,
		PrintLog:   a.printlog,
		PrintError: a.printerr,
		Num:        a.num,
		Sighup:     sig,
	}
	if err := archer.StartHTTPArcher(cfg); err != nil {
		log.Fatal(err)
	}
	return subcommands.ExitSuccess
}
