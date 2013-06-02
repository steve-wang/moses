package moses

import (
	"fmt"
	"io"
)

type hello struct {
	ver byte
	methods []byte
}

func (p *hello) read(r io.Reader) error {
	var head [2]byte
	if err := read(r, head[:]); err != nil {
		return err
	}
	if count := int(head[1]); count > 0 {
		methods := make([]byte, count)
		if err := read(r, methods); err != nil {
			return err
		}
		p.methods = methods
	}
	p.ver = head[0]
	return nil
}

func (p *hello) findMethod(method byte) bool {
	for _, m := range p.methods {
		if m == method {
			return true
		}
	}
	return false
}

type reqname struct {
	Name string
	Port uint16
}

func (p *reqname) Address() string {
	return fmt.Sprintf("%s:%d", p.Name, p.Port)
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
		p.Name = string(name)
	}
	var port [2]byte
	if err := read(r, port[:]); err != nil {
		return err
	}
	p.Port = uint16(port[0]) << 8
	p.Port |= uint16(port[1])
	return nil
}

func (p *reqname) Write(w io.Writer) error {
	var buf [258]byte
	buf[0] = byte(len(p.Name))
	index := 1
	for i := byte(0); i < byte(len(p.Name)); i++ {
		buf[index] = p.Name[i]
		index++
	}
	buf[index] = byte((p.Port >> 8) & 0xff)
	index++
	buf[index] = byte(p.Port & 0xff)
	index++
	return write(w, buf[:index])
}

type reqipv4 struct {
	IP [4]byte
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
