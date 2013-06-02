package moses

import (
	"fmt"
	"io"
	"net"
	"sync"
)

type Connector interface {
	Connect(string) (net.Conn, error)
}

type Proxy struct {
	Con1, Con2 net.Conn
}

type Acceptor interface {
	Accept(net.Conn, Connector) (*Proxy, error)
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
	proxy, err := p.acceptor.Accept(src, p.connector)
	if err != nil {
		src.Close()
		return err
	}
	var wg sync.WaitGroup
	cherr := make(chan error, 2)
	run := func(src, dst net.Conn) {
		defer wg.Done()
		_, err := io.Copy(src, dst)
		cherr <- err
	}
	wg.Add(2)
	go run(proxy.Con1, proxy.Con2)
	go run(proxy.Con2, proxy.Con1)
	<-cherr
	proxy.Con1.Close()
	proxy.Con2.Close()
	wg.Wait()
	return nil
}
