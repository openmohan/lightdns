package lightdns

import (
	"fmt"
	"net"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

type ZoneType uint16

const (
	DNSForwardLookupZone ZoneType = 1
	DNSReverseLookupZone ZoneType = 2 //Todo
)

type Handler interface {
	serveDNS(*udpConnection, *layers.DNS)
}

//DNSServer is the contains the runtime information
type DNSServer struct {
	port    int
	handler Handler
}

//NewDNSServer - Creates new DNSServer
func NewDNSServer(port int) *DNSServer {
	handler := NewServeMux()
	return &DNSServer{port: port, handler: handler}
}

type serveMux struct {
	handler map[string]Handler
}

//NewServeMux - creates a servemux with handler field intialised
func NewServeMux() *serveMux {
	h := make(map[string]Handler)
	return &serveMux{handler: h}
}

//handleFunc - registers a zone data with a handler for it
func (srv *serveMux) handleFunc(pattern string, f func(*udpConnection, *layers.DNS)) {
	srv.handler[pattern] = handlerConvert(f)
}

func (srv *serveMux) serveDNS(u *udpConnection, request *layers.DNS) {
	var h Handler
	if len(request.Questions) < 1 { // allow more than one question
		return
	}
	if h = srv.match(string(request.Questions[0].Name), request.Questions[0].Type); h == nil {
		//todo: log handler not found
		fmt.Println("no handler found for ", request.Questions[0].Name)
	} else {
		h.serveDNS(u, request)
	}
}

//StartToServe - creates a UDP connection and uses the connection to serve DNS
func (dns *DNSServer) StartAndServe() {
	addr := net.UDPAddr{
		Port: dns.port,
		IP:   net.ParseIP("127.0.0.1"),
	}
	l, _ := net.ListenUDP("udp", &addr)
	udpConnection := &udpConnection{conn: l}
	dns.serve(udpConnection)
}

func (dns *DNSServer) serve(u *udpConnection) {
	for {
		tmp := make([]byte, 1024)
		_, addr, _ := u.conn.ReadFrom(tmp)
		u.addr = addr
		packet := gopacket.NewPacket(tmp, layers.LayerTypeDNS, gopacket.Default)
		dnsPacket := packet.Layer(layers.LayerTypeDNS)
		tcp, _ := dnsPacket.(*layers.DNS)
		dns.handler.serveDNS(u, tcp)
	}
}

type handlerConvert func(*udpConnection, *layers.DNS)

func (f handlerConvert) serveDNS(w *udpConnection, r *layers.DNS) {
	f(w, r)
}

type udpConnection struct {
	conn net.PacketConn
	addr net.Addr
}

func (udp *udpConnection) Write(b []byte) error {
	udp.conn.WriteTo(b, udp.addr)
	return nil
}

type customHandler func(string) (string, error)

func generateHandler(records map[string]string, lookupFunc customHandler) func(w *udpConnection, r *layers.DNS) {
	return func(w *udpConnection, r *layers.DNS) {
		switch r.Questions[0].Type {
		case layers.DNSTypeA:
			handleATypeQuery(w, r, records, lookupFunc)
		}
	}
}

//AddZoneData - Depending on the zoneType and recordType  this function generates appropriate handler and registers in the serveMux
func (dns *DNSServer) AddZoneData(zone string, records map[string]string, lookupFunc func(string) (string, error), lookupZone ZoneType) {
	if lookupZone == DNSForwardLookupZone {
		serveMuxCurrent := dns.handler.(*serveMux)
		serveMuxCurrent.handleFunc(zone, generateHandler(records, lookupFunc))
	}
}

func (srv *serveMux) match(q string, t layers.DNSType) Handler {
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
		if h, ok := srv.handler[string(b[:l])]; ok { // causes garbage, might want to change the map key
			if uint16(t) != uint16(43) {
				return h
			}
			// Continue for DS to see if we have a parent too, if so delegeate to the parent
			handler = h
		}
		off, end = nextLabel(q, off)
		if end {
			break
		}
	}
	// Wildcard match, if we have found nothing try the root zone as a last resort.
	if h, ok := srv.handler["."]; ok {
		return h
	}
	return handler
}

func nextLabel(s string, offset int) (i int, end bool) {
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

func handleATypeQuery(w *udpConnection, r *layers.DNS, records map[string]string, lookupFunc customHandler) {
	replyMess := r
	var dnsAnswer layers.DNSResourceRecord
	dnsAnswer.Type = layers.DNSTypeA
	var ip string
	var err error
	var ok bool
	if lookupFunc == nil {
		ip, ok = records[string(r.Questions[0].Name)]
		if !ok {
			//Todo: Log no data present for the IP and handle:todo
		}
	} else {
		ip, err = lookupFunc(string(r.Questions[0].Name))
	}
	a, _, _ := net.ParseCIDR(ip + "/24")
	dnsAnswer.Type = layers.DNSTypeA
	dnsAnswer.IP = a
	dnsAnswer.Name = []byte(r.Questions[0].Name)
	dnsAnswer.Class = layers.DNSClassIN
	replyMess.QR = true
	replyMess.ANCount = 1
	replyMess.OpCode = layers.DNSOpCodeNotify
	replyMess.AA = true
	replyMess.Answers = append(replyMess.Answers, dnsAnswer)
	replyMess.ResponseCode = layers.DNSResponseCodeNoErr
	buf := gopacket.NewSerializeBuffer()
	opts := gopacket.SerializeOptions{} // See SerializeOptions for more details.
	err = replyMess.SerializeTo(buf, opts)
	if err != nil {
		panic(err)
	}
	w.Write(buf.Bytes())
}
