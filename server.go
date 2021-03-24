package main

import (
	"net"
	"sync"
)

type server struct {
	lock sync.Mutex
}

func (s *server) listen(addr string) error {
	laddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		return err
	}
	lst, err := net.ListenTCP("tcp", laddr)
	if err != nil {
		return err
	}
	defer lst.Close()

	for {
		conn, err := lst.AcceptTCP()
		if err != nil {
			return err
		}

		go s.handleConnection(conn)
	}
}

func (s *server) CloseConnections() {
	for _, conn := range s.connections {
		conn.Close()
	}
}

func (s *server) handleConnection(conn net.Conn) {
	defer conn.Close()
}
