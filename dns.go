package lightdns

import (
	"fmt"
	"net"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

type handlerConvert func(*UdpConnection, *layers.DNS)

//DNSServer is the contains the runtime information
type DNSServer struct {
	port    int
	handler Handler
}

type Handler interface {
	serveDNS(*UdpConnection, *layers.DNS)
}

type UdpConnection struct {
	conn net.PacketConn
	addr net.Addr
}

//Create new DNSServer
func NewDNSServer(port int, handler Handler) *DNSServer {
	return &DNSServer{port: port, handler: handler}
}

type serveMux struct {
	handler map[string]Handler
}

func NewServeMux() *serveMux {
	h := make(map[string]Handler)
	return &serveMux{handler: h}
}

// func (f HandlerConvert) serveDNS(u *UdpResponse, r *layers.DNS) {
// 	f(u, r)
// }

func (srv *serveMux) HandleFunc(pattern string, f func(*UdpConnection, *layers.DNS)) {
	srv.handler[pattern] = handlerConvert(f)
}

func (f handlerConvert) serveDNS(w *UdpConnection, r *layers.DNS) {
	f(w, r)
}

func (dns *DNSServer) StartToServe() {
	addr := net.UDPAddr{
		Port: 1234,
		IP:   net.ParseIP("127.0.0.1"),
	}
	l, _ := net.ListenUDP("udp", &addr)
	udpConnection := &UdpConnection{conn: l}
	dns.serve(udpConnection)
}

func (dns *DNSServer) serve(u *UdpConnection) {
	for {
		tmp := make([]byte, 1024)
		_, addr, _ := u.conn.ReadFrom(tmp)
		u.addr = addr

		packet := gopacket.NewPacket(tmp, layers.LayerTypeDNS, gopacket.Default)
		dnsPacket := packet.Layer(layers.LayerTypeDNS)
		tcp, _ := dnsPacket.(*layers.DNS)
		dns.handler.serveDNS(u, tcp)
		fmt.Println(tcp.OpCode)
	}
}

func (srv *serveMux) serveDNS(u *UdpConnection, request *layers.DNS) {
	fmt.Println("tetssss")
	var h Handler
	fmt.Println("sss")
	if len(request.Questions) < 1 { // allow more than one question
		fmt.Println("Nothing exists")
		return
	}
	fmt.Println(string(request.Questions[0].Name))
	if h = srv.match(string(request.Questions[0].Name), request.Questions[0].Type); h == nil {
		fmt.Println("retuernd")
	}
	fmt.Println("retuernd")
	if h == nil {
		fmt.Println("no ter")
	} else {
		h.serveDNS(u, request)
	}
}

func (udp *UdpConnection) Write(b []byte) error {
	udp.conn.WriteTo(b, udp.addr)
	return nil
}

func (mux *serveMux) match(q string, t layers.DNSType) Handler {
	fmt.Println("22222")
	fmt.Println(mux)
	fmt.Println(q)

	var handler Handler
	b := make([]byte, len(q)) // worst case, one label of length q
	off := 0
	end := false
	for {
		l := len(q[off:])
		for i := 0; i < l; i++ {
			b[i] = q[off+i]
			if b[i] >= 'A' && b[i] <= 'Z' {
				b[i] |= 'a' - 'A'
			}
		}
		fmt.Println(string(b[:l]))
		if h, ok := mux.handler[string(b[:l])]; ok { // causes garbage, might want to change the map key
			if uint16(t) != uint16(43) {
				return h
			}
			// Continue for DS to see if we have a parent too, if so delegeate to the parent
			handler = h
		}
		off, end = NextLabel(q, off)
		if end {
			break
		}
	}
	// Wildcard match, if we have found nothing try the root zone as a last resort.
	if h, ok := mux.handler["."]; ok {
		return h
	}
	return handler
}

func NextLabel(s string, offset int) (i int, end bool) {
	quote := false
	for i = offset; i < len(s)-1; i++ {
		switch s[i] {
		case '\\':
			quote = !quote
		default:
			quote = false
		case '.':
			if quote {
				quote = !quote
				continue
			}
			return i + 1, false
		}
	}
	return i + 1, true
}
