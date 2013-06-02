package moses

import (
	"fmt"
	"io"
	"math/rand"
	"net"
	"time"
)

type PrivateConnection struct {
	net.Conn
	rw io.ReadWriter
}

func (p *PrivateConnection) Read(data []byte) (int, error) {
	return p.rw.Read(data)
}

func (p *PrivateConnection) Write(data []byte) (int, error) {
	return p.rw.Write(data)
}

type PrivateConnector struct {
	User string
	Password string
	ProxyAddress string
}

func (p *PrivateConnector) Connect(address string) (_ net.Conn, err error) {
	conn, err := net.Dial("tcp4", p.ProxyAddress)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			conn.Close()
		}
	}()
	rn := rand.New(rand.NewSource(time.Now().Unix())).Uint32()
	var key [4]byte
	key[0] = byte(rn)
	key[1] = byte((rn >> 8) & 0xff)
	key[2] = byte((rn >> 16) & 0xff)
	key[3] = byte((rn >> 24) & 0xff)
	if err := write(conn, key[:]); err != nil {
		return nil, err
	}
	client := &PrivateConnection{
		Conn: conn,
		rw: NewPrivateReadWriter(conn, key),
	}
	auth := Auth{
		User: p.User,
		Password: p.Password,
		Address: address,
	}
	if err := auth.Write(client); err != nil {
		return nil, err
	}
	var permit [1]byte
	if err := read(client, permit[:]); err != nil {
		return nil, err
	}
	if permit[0] != 0 {
		return nil, fmt.Errorf("failed: %d", permit[0])
	}
	return client, nil
}
