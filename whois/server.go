package whois

import (
	"context"
	"fmt"

	"github.com/millken/jwhois"
)

type Server struct {
}

func NewService() *Server {
	return &Server{}
}

// Ip Service
func (s *Server) IP(ctx context.Context, req *Request) (res *Response, err error) {

	return nil, nil
}

// Domain service
func (s *Server) Domain(ctx context.Context, req *Request) (res *Response, err error) {
	w, err := jwhois.Whois(req.Name)

	if err != nil {
		return nil, err
	}
	fmt.Println("Source: " + string(w))

	res = &Response{
		Data: string(w),
	}
	return res, nil
}
