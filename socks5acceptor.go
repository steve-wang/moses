package moses

import (
	"errors"
	"fmt"
	"io"
)

const (
	_VERSION = 5
)

type AuthChecker interface {
	Check(string, string) bool
}

type SOCKS5Acceptor struct {
	checker AuthChecker
}

func NewSOCK5Acceptor(checker AuthChecker) *SOCKS5Acceptor {
	return &SOCKS5Acceptor{checker: checker}
}

func (p *SOCKS5Acceptor) welcome(rw io.ReadWriter) (err error) {
	var hi hello
	if p.checker == nil {
		return hi.match(rw, 0)
	}
	// need authentication
	if err := hi.match(rw, 2); err != nil {
		return err
	}
	user, pass, err := hi.auth(rw)
	if err != nil {
		return err
	}
	resp := [2]byte{1, 0}
	if !p.checker.Check(user, pass) {
		resp[1] = 0xff
		write(rw, resp[:])
		return errors.New("AUTH failed")
	}
	return write(rw, resp[:])
}

func (p *SOCKS5Acceptor) Accept(src Connection, connector Connector) (_ Connection, _ Connection, err error) {
	if err := p.welcome(src); err != nil {
		return nil, nil, err
	}
	var head [4]byte
	if err := read(src, head[:]); err != nil {
		return nil, nil, err
	}
	if head[0] != _VERSION {
		return nil, nil, fmt.Errorf("illegal VER: %d", head[0])
	}
	if head[1] != 1 {
		return nil, nil, fmt.Errorf("unsupported CMD: %d", head[1])
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
		return nil, nil, fmt.Errorf("unsupported ATYP: %d", head[3])
	}
	if err := req.Read(src); err != nil {
		return nil, nil, err
	}
	dst, err := connector.Connect(req.Address())
	if err != nil {
		write(src, []byte{_VERSION, 4, 0, head[3]})
		req.Write(src)
		return nil, nil, err
	}
	defer func() {
		if err != nil {
			dst.Close()
		}
	}()
	if err := write(src, []byte{_VERSION, 0, 0, 1}); err != nil {
		return nil, nil, err
	}
	if err := req.Write(src); err != nil {
		return nil, nil, err
	}
	return src, dst, nil
}
