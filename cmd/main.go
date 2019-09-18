package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/dappledger/AnnSecChan/handlers/app"
	"github.com/dappledger/AnnSecChan/modules/config"
)

var (
	configFile = flag.String("c", "", "config file path")
)

type IServer interface {
	OnStart(cfg *config.Config) error
	Shutdown()
}

func interceptSignal(s IServer) {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		s.Shutdown()
		os.Exit(0)
	}()
}

func main() {
	fmt.Println("begin start server")
	flag.Parse()
	runtime.GOMAXPROCS(runtime.NumCPU())
	cfg := config.LoadConfigFile(*configFile)
	appServer := &app.App{}
	interceptSignal(appServer)
	if err := appServer.OnStart(cfg); err != nil {
		fmt.Println(err)
	}
	return
}
