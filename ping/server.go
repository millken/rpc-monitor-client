package ping

import (
	"fmt"
	"net"

	"google.golang.org/grpc"
)

type Server struct {
}

type StreamServer struct {
}

func NewServer() *Server {
	return &Server{}
}

func (s *Server) Run(listen net.Listener) error {
	srv := grpc.NewServer()
	RegisterPingServer(srv, s)
	return srv.Serve(listen)
}

func (s *Server) Hello(req *Request, stream Ping_HelloServer) error {
	fmt.Printf("%+v", req)
	//res := new(Response)
	pinger, err := NewPinger("www.baidu.com")
	if err != nil {
		panic(err)
	}
	pinger.Count = 10
	pinger.Run()
	/* 	for {
		res := &Response{Msg: req.Host}
		if err := stream.Send(res); err != nil {
			return err
		}
	} */
	return nil
}
