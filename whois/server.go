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

func (s *Server) Whois(ctx context.Context, req *Request) (res *Response, err error) {
	tpe, w, err := jwhois.Whois(req.Name)

	if err != nil {
		return nil, err
	}
	fmt.Println("Source: " + string(w))

	res = &Response{
		Type: tpe,
		Data: string(w),
	}
	return res, nil
}
