package moses

import (
	"fmt"
	"io"
	"net"
	"sync"
)

type Connection interface {
	Read([]byte) (int, error)
	Write([]byte) (int, error)
	Close() error
}

type Connector interface {
	Connect(string) (Connection, error)
}

type Acceptor interface {
	Accept(Connection, Connector) (Connection, Connection, error)
}

type Server struct {
	listener  net.Listener
	connector Connector
	acceptor  Acceptor
}

func NewServer(acceptor Acceptor, connector Connector) *Server {
	return &Server{
		acceptor:  acceptor,
		connector: connector,
	}
}

func (p *Server) Start(port uint16) error {
	listener, err := net.Listen("tcp4", fmt.Sprintf(":%d", port))
	if err != nil {
		return err
	}
	p.listener = listener
	return nil
}

func (p *Server) Close() {
	p.listener.Close()
}

func (p *Server) Serve() {
	for {
		conn, err := p.listener.Accept()
		if err != nil {
			break
		}
		go p.process(conn)
	}
}

func (p *Server) process(src net.Conn) (err error) {
	con1, con2, err := p.acceptor.Accept(src, p.connector)
	if err != nil {
		src.Close()
		return err
	}
	var wg sync.WaitGroup
	cherr := make(chan error, 2)
	run := func(src, dst Connection) {
		defer wg.Done()
		_, err := io.Copy(src, dst)
		cherr <- err
	}
	wg.Add(2)
	go run(con1, con2)
	go run(con2, con1)
	<-cherr
	con1.Close()
	con2.Close()
	wg.Wait()
	return nil
}
