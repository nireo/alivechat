package main

import (
	"encoding/json"
	"github.com/google/uuid"
	"log"
	"net"
	"strings"
	"time"
)

type ConnectionStatus int

const (
	JOINING ConnectionStatus = iota
	LEAVING
)

func genUUID() string {
	uid := uuid.New()
	return strings.Replace(uid.String(), "-", "", -1)
}

type server struct {
	conn           *net.UDPConn
	messageChannel chan message

	// uuid is the unique identifier
	clients        map[string]client
}

type client struct {
	ID   string
	Name string
	Addr *net.UDPAddr
}

type message struct {
	ID        string
	Name      string
	Content   string
	ToID      string
	From      *net.UDPAddr
	To        *net.UDPAddr
	Timestamp int64
	Action int
}

func (s *server) handleMessage() {
	var buffer [4096]byte
	size, addr, err := s.conn.ReadFromUDP(buffer[0:])
	if err != nil {
		return
	}

	msg := buffer[0:size]
	m := s.parse(msg)

	// set the message information
	m.Timestamp = time.Now().Unix()

	// handle different actions
	switch m.Action {
	case 0:
		// create new user
		var c client
		c.Addr = addr
		c.ID = genUUID()
		c.Name = m.Name
		m.ToID = ""

		s.clients[c.ID] = c
		m.ID = c.ID
		s.messageChannel <- m
	case 1:
		m.From = addr
		// find the user to send the message to
		to, _ := s.clients[m.ToID]
		m.To = to.Addr
		s.messageChannel <- m
	default:
		log.Println("unsupported message action")
	}

}

func (s *server) parse(msg []byte) message {
	var m message
	json.Unmarshal(msg, &m)
	return m
}

func (s *server) handleMessageSend() {
	for {
		msg := <- s.messageChannel
		data, _ := json.Marshal(msg)
		if msg.ToID != "" {
			s.conn.WriteToUDP(data, msg.To)
		} else {
			s.conn.WriteToUDP(data, msg.From)
		}
	}
}

func main() {
	lAddr, err := net.ResolveUDPAddr("udp4", ":1200")
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
