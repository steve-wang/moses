package moses

import (
	"fmt"
	"io"
	"net"
)

type hello struct {
	ver     byte
	methods []byte
}

func (p *hello) match(rw io.ReadWriter, method byte) error {
	var head [2]byte
	if err := read(rw, head[:]); err != nil {
		return err
	}
	if head[0] != _VERSION {
		return fmt.Errorf("VER is invalid: %d", head[0])
	}
	count := int(head[1])
	if count <= 0 {
		return fmt.Errorf("NMETHODS is invliad: %d", count)
	}
	methods := make([]byte, count)
	if err := read(rw, methods); err != nil {
		return err
	}
	for _, m := range methods {
		if m == method {
			return write(rw, []byte{_VERSION, m})
		}
	}
	write(rw, []byte{_VERSION, 0xff})
	return fmt.Errorf("METHOD[%d] not found: %v", method, methods)
}

func (p *hello) auth(r io.Reader) (string, string, error) {
	var buff [2]byte
	if err := read(r, buff[:]); err != nil {
		return "", "", err
	}
	if buff[0] != 1 {
		return "", "", fmt.Errorf("invalid VER: %d", buff[0])
	}
	name := make([]byte, int(buff[1]))
	if err := read(r, name); err != nil {
		return "", "", err
	}
	if err := read(r, buff[:1]); err != nil {
		return "", "", err
	}
	pass := make([]byte, int(buff[0]))
	if err := read(r, pass); err != nil {
		return "", "", err
	}
	return string(name), string(pass), nil
}

func (p *hello) findMethod(method byte) bool {
	for _, m := range p.methods {
		if m == method {
			return true
		}
	}
	return false
}

type reqipv4 struct {
	IP   [4]byte
	Port uint16
}

func (p *reqipv4) Address() string {
	return fmt.Sprintf(
		"%d.%d.%d.%d:%d",
		p.IP[0], p.IP[1], p.IP[2], p.IP[3], p.Port)
}

func (p *reqipv4) Read(r io.Reader) error {
	if err := read(r, p.IP[:]); err != nil {
		return err
	}
	var port [2]byte
	if err := read(r, port[:]); err != nil {
		return err
	}
	p.Port = uint16(port[0]) << 8
	p.Port |= uint16(port[1])
	return nil
}

func (p *reqipv4) Write(w io.Writer) error {
	if err := write(w, p.IP[:]); err != nil {
		return err
	}
	var port [2]byte
	port[0] = byte((p.Port >> 8) & 0xff)
	port[1] = byte(p.Port & 0xff)
	return write(w, port[:])
}

type reqname struct {
	reqipv4
}

func (p *reqname) Read(r io.Reader) error {
	var data [1]byte
	if err := read(r, data[:]); err != nil {
		return err
	}
	size := int(data[0])
	if size > 0 {
		name := make([]byte, size)
		if err := read(r, name[:]); err != nil {
			return err
		}
		addr, err := net.ResolveIPAddr("ip4", string(name))
		if err != nil {
			return err
		}
		copy(p.IP[:], addr.IP.To4())
	}
	var port [2]byte
	if err := read(r, port[:]); err != nil {
		return err
	}
	p.Port = uint16(port[0]) << 8
	p.Port |= uint16(port[1])
	return nil
}

type Auth struct {
	User     string
	Password string
	Address  string
}

func (p *Auth) Read(r io.Reader) (err error) {
	p.User, err = readString(r)
	if err != nil {
		return err
	}
	p.Password, err = readString(r)
	if err != nil {
		return err
	}
	p.Address, err = readString(r)
	return err
}

func (p *Auth) Write(w io.Writer) error {
	if err := writeString(w, p.User); err != nil {
		return err
	}
	if err := writeString(w, p.Password); err != nil {
		return err
	}
	return writeString(w, p.Address)
}
