package main

import (
	"context"
	"log"

	"../../whois"
	"google.golang.org/grpc"
)

func main() {
	address := "127.0.0.1:6543"
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	client := whois.NewWhoisClient(conn)

	req := &whois.Request{
		Name: "www.g.com",
	}
	data, err := client.IP(context.Background(), req)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	log.Printf("RECV: %s", data)
	log.Println("END")
}
