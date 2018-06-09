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
	RegisterPingServer(srv, &StreamServer{})
	return srv.Serve(listen)
}

func (s *StreamServer) Hello(req *Request, stream Ping_HelloServer) error {
	fmt.Printf("%+v", req)
	//res := new(Response)
	for {
		res := &Response{Msg: req.Host}
		if err := stream.Send(res); err != nil {
			return err
		}
	}
	return nil
}

func (s *Server) Send(res *Response) error {
	res = new(Response)

	res.Msg = fmt.Sprintf("q=%s", "req.Host")

	return nil
}
