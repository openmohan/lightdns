package lightdns

import (
	"github.com/google/gopacket/layers"
)

//DNSServer is the contains the runtime information
type DNSServer struct {
	port    int
	handler Handler
}

type Handler interface {
	ServeDNS(r *layers.DNS)
}

func NewDNSServer(port int) {
	return
}
