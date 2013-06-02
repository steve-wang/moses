package moses

import (
	"errors"
	"fmt"
	"io"
	"net"
)

const (
	_VERSION = 5
)

type SOCKS5Acceptor struct {
	connector Connector
}

func NewSOCKS5Acceptor(connector Connector) *SOCKS5Acceptor {
	return &SOCKS5Acceptor{connector}
}

func (p *SOCKS5Acceptor) welcome(rw io.ReadWriter) (err error) {
	var hi hello
	if err := hi.read(rw); err != nil {
		return err
	}
	defer func() {
		rep := []byte{_VERSION, 0}
		if err != nil {
			rep[1] = 0xff
			write(rw, rep[:])
		} else {
			err = write(rw, rep[:])
		}
	}()
	if !hi.findMethod(0) {
		return errors.New("METHOD(0) not found")
	}
	return nil
}

func (p *SOCKS5Acceptor) Accept(src net.Conn) (_ *Proxy, err error) {
	if err := p.welcome(src); err != nil {
		return nil, err
	}
	var head [4]byte
	if err := read(src, head[:]); err != nil {
		return nil, err
	}
	if head[0] != _VERSION {
		return nil, fmt.Errorf("illegal VER: %d", head[0])
	}
	if head[1] != 1 {
		return nil, fmt.Errorf("unsupported CMD: %d", head[1])
	}
	var req interface {
		Read(io.Reader) error
		Write(io.Writer) error
		Address() string
	}
	switch head[3] {
	case 1:
		req = &reqipv4{}
	case 3:
		req = &reqname{}
	default:
		return nil, fmt.Errorf("unsupported ATYP: %d", head[3])
	}
	if err := req.Read(src); err != nil {
		return nil, err
	}
	dst, err := p.connector.Connect(req.Address())
	if err != nil {
		write(src, []byte{_VERSION, 4, 0, head[3]})
		req.Write(src)
		return nil, err
	}
	defer func() {
		if err != nil {
			dst.Close()
		}
	}()
	if err := write(src, []byte{_VERSION, 0, 0, head[3]}); err != nil {
		return nil, err
	}
	if err := req.Write(src); err != nil {
		return nil, err
	}
	return &Proxy{
		c1: src,
		c2: dst}, nil
}
