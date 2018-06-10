package main

import (
	"context"
	"io"
	"log"

	"../../ping"
	"google.golang.org/grpc"
)

func main() {
	address := "127.0.0.1:6543"
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	client := ping.NewPingClient(conn)

	req := &ping.Request{
		Host: "www.g.com",
	}
	stream, err := client.Hello(context.Background(), req)
	if err != nil {
		log.Fatalf("Error on get Hello: %v", err)
	}
	for {
		hello, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("%v.Hello(_) = _, %v", client, err)
		}
		log.Printf("Hello: %+v time: %.2f", hello, hello.Time)
	}
	log.Println("END")
}
