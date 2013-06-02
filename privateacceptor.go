package moses

import (
	"errors"
	"net"
)

type AuthChecker interface {
	Check(user, password string) bool
}

type PrivateAcceptor struct {
	checker AuthChecker
}

func NewPrivateAcceptor(checker AuthChecker) *PrivateAcceptor {
	return &PrivateAcceptor{
		checker: checker,
	}
}

func (p *PrivateAcceptor) Accept(conn net.Conn, connector Connector) (*Proxy, error) {
	var key [4]byte
	if err := read(conn, key[:]); err != nil {
		return nil, err
	}
	client := &PrivateConnection{
		Conn: conn,
		rw:   NewPrivateReadWriter(conn, key),
	}
	var auth Auth
	if err := auth.Read(client); err != nil {
		return nil, err
	}
	if p.checker == nil || !p.checker.Check(auth.User, auth.Password) {
		client.Write([]byte{ERRCODE_AUTH})
		return nil, errors.New("not permitted")
	}
	dst, err := connector.Connect(auth.Address)
	if err != nil {
		client.Write([]byte{ERRCODE_CONN})
		return nil, err
	}
	defer func() {
		if err != nil {
			dst.Close()
		}
	}()
	if err := write(client, []byte{0}); err != nil {
		return nil, err
	}
	return &Proxy{
		c1: client,
		c2: dst}, nil
}
