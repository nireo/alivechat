package main

import (
	"flag"
	"log"
	"net"
)

var port string

func init() {
	flag.StringVar(&port, "port", "1200", "the port to host the server")
	flag.Parse()
}

func main() {
	lAddr, err := net.ResolveUDPAddr("udp4", ":"+port)
	if err != nil {
		log.Fatalf("error resolving udp addr: %s", err)
	}

	var s server
	s.messageChannel = make(chan message)
	s.clients = make(map[string]client, 0)

	s.conn, err = net.ListenUDP("udp", lAddr)
	if err != nil {
		log.Fatalf("error listening for udp connection")
	}

	// run the message sending channel in a independent goroutine
	go s.handleMessageSend()

	for {
		s.handleMessage()
	}
}
