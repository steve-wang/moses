package moses

import (
	"fmt"
)

type SOCKS4Acceptor struct {
}

func (p SOCKS4Acceptor) Accept(src Connection, connector Connector) (_ Connection, _ Connection, err error) {
	var head [8]byte
	if err := read(src, head[:]); err != nil {
		return nil, nil, err
	}
	switch {
	case head[0] != 4:
		return nil, nil, fmt.Errorf("invalid VN: %d", head[0])
	case head[1] != 1:
		return nil, nil, fmt.Errorf("invalid CD: %d", head[1])
	}
	var buff [256]byte
	for {
		n, err := src.Read(buff[:])
		if err != nil {
			return nil, nil, err
		}
		if n > 0 && buff[n-1] == 0 {
			break
		}
	}
	port := uint16(head[3])
	port |= uint16(head[2]) << 8
	address := fmt.Sprintf("%d.%d.%d.%d:%d", head[4], head[5], head[6], head[7], port)
	var resp [8]byte
	copy(resp[2:], head[2:])
	dst, err := connector.Connect(address)
	if err != nil {
		resp[1] = 92
		write(src, resp[:])
		return nil, nil, err
	}
	resp[1] = 90
	if err := write(src, resp[:]); err != nil {
		dst.Close()
		return nil, nil, err
	}
	return src, dst, nil
}
