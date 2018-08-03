package ping

import (
	"log"
	"math"
	"runtime"
	"time"
)

const defaultPacketNum = 100

type Server struct {
}
type milliDuration time.Duration

func (hd milliDuration) Float() float32 {
	milliseconds := time.Duration(hd).Nanoseconds()
	milliseconds = milliseconds / 1000000
	return float32(milliseconds)
}

func NewService() *Server {
	return &Server{}
}

// Hello Service
func (s *Server) Hello(req *Request, stream Ping_HelloServer) error {
	defer func() {
		err := recover()
		if err != nil {
			buf := make([]byte, 1024)
			n := runtime.Stack(buf, true)
			log.Printf("[ERROR] %s", string(buf[:n]))
			runtime.Goexit()
		}
	}()
	pinger, err := NewPinger(req.Host)
	if err != nil {
		return err
	}
	if req.GetCount() == 0 {
		pinger.SetCount(defaultPacketNum)
	} else {
		pinger.SetCount(int(req.GetCount()))
	}
	pinger.OnRecv = func(pkt *Packet) {
		res := &Response{
			Addr: pkt.IPAddr.String(),
			Time: If(pkt.Nbytes == 0, float32(-1), float32(Round(float64(pkt.Rtt.Nanoseconds())/float64(1000000), 2))).(float32),
			Seq:  int32(pkt.Seq),
		}
		if err := stream.Send(res); err != nil {
			log.Printf("ERROR: %s", err)
			pinger.Stop()
			return
		}
	}
	pinger.Run()

	return nil
}

func If(condition bool, trueVal, falseVal interface{}) interface{} {
	if condition {
		return trueVal
	}
	return falseVal
}

func Round(f float64, n int) float64 {
	pow10_n := math.Pow10(n)
	return math.Trunc((f+0.0)*pow10_n) / pow10_n
}
