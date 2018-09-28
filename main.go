package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"runtime"

	"./ping"
	"./whois"
	"github.com/kardianos/service"
	"google.golang.org/grpc"
)

var listen string

func init() {
	flag.StringVar(&listen, "listen", ":58000", "default listen address")
}

func stack() {
	err := recover()
	if err != nil {
		buf := make([]byte, 1024)
		n := runtime.Stack(buf, true)
		fmt.Printf("[ERROR] %s", string(buf[:n]))
		runtime.Goexit()
	}

}

type program struct{}

func (p *program) Start(s service.Service) error {
	go p.run()
	return nil
}
func (p *program) run() {
	listen, err := net.Listen("tcp", listen)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	srv := grpc.NewServer()
	ping.RegisterPingServer(srv, ping.NewService())
	whois.RegisterWhoisServer(srv, whois.NewService())
	srv.Serve(listen)
}
func (p *program) Stop(s service.Service) error {
	return nil
}

func main() {
	var err error
	defer stack()
	if err != nil {
		panic(err)
	}
	flag.Parse()
	svcConfig := &service.Config{
		Name:        "rpc-monitor-client",
		DisplayName: "Agent for RPC monitor",
		Description: " This is a monitor tool for network.  It is designed to run well.",
	}

	prg := &program{}
	s, err := service.New(prg, svcConfig)
	if err != nil {
		log.Fatal(err)
	}
	if len(os.Args) > 1 {

		err = service.Control(s, os.Args[1])
		if err != nil {
			fmt.Printf("Failed (%s) : %s\n", os.Args[1], err)
			return
		}
		fmt.Printf("Succeeded (%s)\n", os.Args[1])
		return
	}
	s.Run()
}
