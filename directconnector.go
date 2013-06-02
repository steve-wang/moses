package moses 

import "net"

type DirectConnector struct {
}

func (p *DirectConnector) Connect(address string) (net.Conn, error) {
	return net.Dial("tcp4", address)
} 