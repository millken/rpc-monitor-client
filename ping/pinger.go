package ping

import (
	"fmt"
	"log"
	"math/rand"
	"net"
	"sync"
	"syscall"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
	"golang.org/x/net/ipv6"
)

const (
	timeSliceLength  = 8
	protocolICMP     = 1
	protocolIPv6ICMP = 58
)

// Pinger represents ICMP packet sender/receiver
type Pinger struct {
	// Interval is the wait time between each packet send. Default is 1s.
	Interval time.Duration
	// Count tells pinger to stop after sending (and receiving) Count echo
	// packets. If this option is not specified, pinger will operate until
	// interrupted.
	Count int
	// Number of packets sent
	PacketsSent int

	// Number of packets received
	PacketsRecv int

	// Number of packets lost
	PacketsLost int
	id          int
	ipaddr      *net.IPAddr
	ipv4        bool
	size        int
	sequence    int
	lastSeq     int
	// stop chan bool
	done         chan bool
	lastSendTime time.Time

	// OnRecv is called when Pinger receives and processes a packet
	OnRecv func(*Packet)
}

// Packet represents a received and processed ICMP echo packet.
type Packet struct {
	// Rtt is the round-trip time it took to ping.
	Rtt time.Duration

	// IPAddr is the address of the host being pinged.
	IPAddr *net.IPAddr

	// NBytes is the number of bytes in the message.
	Nbytes int

	// Seq is the ICMP sequence number.
	Seq int
}

type packet struct {
	bytes  []byte
	nbytes int
}

// NewPinger returns a new Pinger struct pointer
func NewPinger(addr string) (*Pinger, error) {
	ipaddr, err := net.ResolveIPAddr("ip", addr)
	if err != nil {
		return nil, err
	}

	var ipv4 bool
	if isIPv4(ipaddr.IP) {
		ipv4 = true
	} else if isIPv6(ipaddr.IP) {
		ipv4 = false
	}

	return &Pinger{
		ipaddr:   ipaddr,
		Interval: time.Second,
		Count:    -1,

		id:          rand.Intn(0xffff),
		ipv4:        ipv4,
		PacketsSent: 0,
		PacketsRecv: 0,
		PacketsLost: 0,
		size:        timeSliceLength,
		sequence:    0,
		lastSeq:     0,
		done:        make(chan bool),
	}, nil
}

// Count ping
func (p *Pinger) SetCount(n int) {
	p.Count = n
}

// Run runs the pinger.
func (p *Pinger) Run() {
	var conn *icmp.PacketConn
	var err error
	if p.ipv4 {
		if conn, err = icmp.ListenPacket("ip4:icmp", ""); err != nil {
			fmt.Printf("Error listening for ICMP packets: %s\n", err.Error())
			return
		}
	} else {
		if conn, err = icmp.ListenPacket("ip6:ipv6-icmp", ""); err != nil {
			fmt.Printf("Error listening for ICMP packets: %s\n", err.Error())
			return
		}
	}
	defer conn.Close()

	var wg sync.WaitGroup
	recv := make(chan *packet, 5)
	wg.Add(1)
	go p.recvICMP(conn, recv, &wg)

	interval := time.NewTicker(p.Interval)
	defer interval.Stop()
	p.lastSendTime = time.Now()
	for {
		select {
		case <-p.done:
			wg.Wait()
			return
		case <-interval.C:
			//timeout
			if time.Since(p.lastSendTime).Seconds() >= 4 && p.lastSeq < p.sequence {
				p.PacketsLost++
				recv <- &packet{bytes: []byte{}, nbytes: 0}
			}
			if p.lastSeq < p.sequence {
				continue
			}
			err = p.sendICMP(conn)
			p.lastSendTime = time.Now()
			if err != nil {
				fmt.Println("FATAL: ", err.Error())
			}
		case r := <-recv:
			err := p.processPacket(r)
			if err != nil {
				fmt.Println("FATAL: ", err.Error())
			}
			if p.Count > 0 && p.PacketsLost+p.PacketsRecv >= p.Count {
				close(p.done)
				wg.Wait()
				return
			}
		}
	}
}

func (p *Pinger) recvICMP(
	conn *icmp.PacketConn,
	recv chan<- *packet,
	wg *sync.WaitGroup,
) {
	defer wg.Done()
	for {
		select {
		case <-p.done:
			return
		default:
			bytes := make([]byte, 512)
			conn.SetReadDeadline(time.Now().Add(time.Millisecond * 100))
			n, _, err := conn.ReadFrom(bytes)
			if err != nil {
				if neterr, ok := err.(*net.OpError); ok {
					if neterr.Timeout() {
						// Read timeout
						continue
					} else {
						close(p.done)
						return
					}
				}
			}

			recv <- &packet{bytes: bytes, nbytes: n}
		}
	}
}

//Stop pinger
func (p *Pinger) Stop() {
	close(p.done)
}

func (p *Pinger) sendICMP(conn *icmp.PacketConn) error {
	var typ icmp.Type
	if p.ipv4 {
		typ = ipv4.ICMPTypeEcho
	} else {
		typ = ipv6.ICMPTypeEchoRequest
	}

	var dst net.Addr = p.ipaddr

	t := timeToBytes(time.Now())
	if p.size-timeSliceLength != 0 {
		t = append(t, byteSliceOfSize(p.size-timeSliceLength)...)
	}
	p.sequence++
	bytes, err := (&icmp.Message{
		Type: typ,
		Code: 0,
		Body: &icmp.Echo{
			ID:   p.id,
			Seq:  p.sequence,
			Data: t,
		},
	}).Marshal(nil)
	if err != nil {
		return err
	}

	for {
		if _, err := conn.WriteTo(bytes, dst); err != nil {
			log.Printf("[ERROR] sendICMP write err: %s", err)
			if neterr, ok := err.(*net.OpError); ok {
				if neterr.Err == syscall.ENOBUFS {
					continue
				}
			}
		}
		p.PacketsSent++
		break
	}
	return nil
}

func (p *Pinger) processPacket(recv *packet) error {
	var bytes []byte
	var proto int
	p.lastSeq++
	if recv.nbytes == 0 {
		outPkt := &Packet{
			Nbytes: recv.nbytes,
			IPAddr: p.ipaddr,
			Seq:    p.lastSeq,
		}

		handler := p.OnRecv
		if handler != nil {
			handler(outPkt)
		}

		return nil
	}
	if p.ipv4 {
		bytes = ipv4Payload(recv.bytes)
		proto = protocolICMP
	} else {
		bytes = recv.bytes
		proto = protocolIPv6ICMP
	}

	var m *icmp.Message
	var err error
	if m, err = icmp.ParseMessage(proto, bytes[:recv.nbytes]); err != nil {
		return fmt.Errorf("Error parsing icmp message")
	}

	if m.Type != ipv4.ICMPTypeEchoReply && m.Type != ipv6.ICMPTypeEchoReply {
		// Not an echo reply, ignore it
		return nil
	}

	outPkt := &Packet{
		Nbytes: recv.nbytes,
		IPAddr: p.ipaddr,
	}

	switch pkt := m.Body.(type) {
	case *icmp.Echo:
		if pkt.ID == p.id && pkt.Seq == p.sequence {
			outPkt.Rtt = time.Since(bytesToTime(pkt.Data[:timeSliceLength]))
			outPkt.Seq = p.lastSeq
			p.PacketsRecv++
			handler := p.OnRecv
			if handler != nil {
				handler(outPkt)
			}
		}
	default:
		return fmt.Errorf("Error, invalid ICMP echo reply. Body type: %T, %s",
			pkt, pkt)
	}

	return nil
}

func byteSliceOfSize(n int) []byte {
	b := make([]byte, n)
	for i := 0; i < len(b); i++ {
		b[i] = 1
	}

	return b
}

func ipv4Payload(b []byte) []byte {
	if len(b) < ipv4.HeaderLen {
		return b
	}
	hdrlen := int(b[0]&0x0f) << 2
	return b[hdrlen:]
}

func bytesToTime(b []byte) time.Time {
	var nsec int64
	for i := uint8(0); i < 8; i++ {
		nsec += int64(b[i]) << ((7 - i) * 8)
	}
	return time.Unix(nsec/1000000000, nsec%1000000000)
}

func isIPv4(ip net.IP) bool {
	return len(ip.To4()) == net.IPv4len
}

func isIPv6(ip net.IP) bool {
	return len(ip) == net.IPv6len
}

func timeToBytes(t time.Time) []byte {
	nsec := t.UnixNano()
	b := make([]byte, 8)
	for i := uint8(0); i < 8; i++ {
		b[i] = byte((nsec >> ((7 - i) * 8)) & 0xff)
	}
	return b
}
