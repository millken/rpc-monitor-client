package whois

import (
	"context"
	"fmt"
	"log"

	"github.com/millken/ripego"
)

type Server struct {
}

func NewService() *Server {
	return &Server{}
}

// Ip Service
func (s *Server) IP(ctx context.Context, req *Request) (res *Response, err error) {
	w, err := ripego.IPLookup("8.8.8.8")

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Inetnum: " + w.Inetnum)
	fmt.Println("Desc: " + w.Descr)
	fmt.Println("Source: " + w.Source)

	res = &Response{
		Data: w.Descr,
	}
	return
}

// Domain service
func (s *Server) Domain(ctx context.Context, req *Request) (res *Response, err error) {

	return nil, nil
}
