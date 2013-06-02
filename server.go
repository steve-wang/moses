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
	c1, c2 net.Conn
}

type Acceptor interface {
	Accept(net.Conn) (*Proxy, error)
}

type Server struct {
	listener        net.Listener
	acceptor Acceptor
}

func NewServer(acceptor Acceptor) *Server {
	return &Server{acceptor: acceptor}
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
	proxy, err := p.acceptor.Accept(src)
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
	go run(proxy.c1, proxy.c2)
	go run(proxy.c2, proxy.c1)
	<-cherr
	proxy.c1.Close()
	proxy.c2.Close()
	wg.Wait()
	return nil
}
