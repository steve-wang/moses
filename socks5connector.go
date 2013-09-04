package moses

import (
	"errors"
	"fmt"
	"net"
)

type SOCKS5Connector struct {
	User     string
	Password string
	Host     string
	Port     uint16
}

func (p *SOCKS5Connector) Connect(address string) (_ Connection, err error) {
	conn, err := net.Dial("tcp4", fmt.Sprintf("%s:%d", p.Host, p.Port))
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			conn.Close()
		}
	}()
	if err := p.auth(conn); err != nil {
		return nil, err
	}
	if err := p.connect(conn, address); err != nil {
		return nil, err
	}
	return conn, nil
}

func (p *SOCKS5Connector) connect(conn Connection, address string) error {
	addr, err := net.ResolveTCPAddr("tcp4", address)
	if err != nil {
		return err
	}
	host := addr.IP.String()
	if len(host) > 255 {
		return errors.New("the lenght of host exceeds limit")
	}
	buff := make([]byte, 7+len(host))
	buff[0] = _VERSION
	buff[1] = 1
	buff[3] = 3
	buff[4] = byte(len(host))
	index := 5
	index += copy(buff[index:], host)
	port := uint16(addr.Port)
	// bigendian
	buff[index] = byte((port & 0xff00) >> 8)
	buff[index+1] = byte(port & 0xff)
	if err := write(conn, buff); err != nil {
		return err
	}
	var head [5]byte
	if err := read(conn, head[:]); err != nil {
		return err
	}
	switch {
	case head[0] != _VERSION:
		return fmt.Errorf("invalid VERSION: %d", head[0])
	case head[1] != 0:
		return fmt.Errorf("failed to connecto to %s: reason(%d)", address, head[1])
	}
	addrlen := 0
	switch head[3] {
	case 1:
		addrlen = 5
	case 3:
		addrlen = int(head[4]) + 2
	default:
		return fmt.Errorf("unsupported ATYP: %d", head[3])
	}
	addrbuf := make([]byte, addrlen)
	return read(conn, addrbuf)
}

func (p *SOCKS5Connector) auth(conn Connection) error {
	head := [3]byte{_VERSION, 1, 0}
	if p.User != "" {
		head[2] = 2
	}
	if err := write(conn, head[:]); err != nil {
		return err
	}
	var resp [2]byte
	if err := read(conn, resp[:]); err != nil {
		return err
	}
	switch {
	case resp[0] != _VERSION:
		return fmt.Errorf("invalid VERSION: %d", resp[0])
	case resp[1] == 0xff:
		return fmt.Errorf("refused METHOD: %d", head[2])
	case resp[1] != head[2]:
		return fmt.Errorf("invalid METHOD: %d", resp[1])
	}
	if head[2] != 2 {
		return nil
	}
	return p.authorize(conn)
}

func (p *SOCKS5Connector) authorize(conn Connection) error {
	if len(p.User) > 255 || len(p.Password) > 255 {
		return errors.New("the length of user or password exceeds limit")
	}
	buff := make([]byte, 3+len(p.User)+len(p.Password))
	index := 0
	buff[index] = 1
	index++
	buff[index] = byte(len(p.User))
	index++
	index += copy(buff[index:], p.User)
	buff[index] = byte(len(p.Password))
	index++
	index += copy(buff[index:], p.Password)
	if err := write(conn, buff); err != nil {
		return err
	}
	var resp [2]byte
	if err := read(conn, resp[:]); err != nil {
		return err
	}
	if resp[0] != 1 {
		return fmt.Errorf("invalid VER: %d", resp[0])
	}
	if resp[1] != 0 {
		return fmt.Errorf("authentication is refused: %d", resp[1])
	}
	return nil
}
