package main

import (
	"log"
	"net"

	"./ping"
)

func main() {
	lis, err := net.Listen("tcp", ":6543")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	pingServer := ping.NewServer()
	pingServer.Run(lis)
}
