package whois

import (
	"context"
	"unicode/utf8"

	"github.com/millken/jwhois"
)

type Server struct {
}

func NewService() *Server {
	return &Server{}
}

func (s *Server) Whois(ctx context.Context, req *Request) (res *Response, err error) {
	w, err := jwhois.Whois(req.Name)

	if err != nil {
		return nil, err
	}
	sw := string(w)
	//fmt.Println("Source: " + sw)
	if !utf8.ValidString(sw) {
		v := make([]rune, 0, len(sw))
		for i, r := range sw {
			if r == utf8.RuneError {
				_, size := utf8.DecodeRuneInString(sw[i:])
				if size == 1 {
					continue
				}
			}
			v = append(v, r)
		}
		sw = string(v)
	}
	res = &Response{
		Data: sw,
	}
	return res, nil
}
