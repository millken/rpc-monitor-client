package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"runtime"

	"./ping"
	"./whois"
	"github.com/hashicorp/logutils"
	"github.com/kardianos/service"
	"google.golang.org/grpc"
)

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
	listen, err := net.Listen("tcp", ":6543")
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
	defer stack()
	filterWriter, err := os.Create("roma.log")
	if err != nil {
		panic(err)
	}
	filter := &logutils.LevelFilter{
		Levels:   []logutils.LogLevel{"DEBUG", "INFO", "WARN", "ERROR"},
		MinLevel: logutils.LogLevel("WARN"),
		Writer:   filterWriter,
	}
	log.SetOutput(filter)

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
