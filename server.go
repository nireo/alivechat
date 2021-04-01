package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"time"
)

// server represents the worker that takes care of redirecting messages to
// other users and handling keeping state of the chat.
type server struct {
	conn           *net.UDPConn
	messageChannel chan message

	// uuid is the unique identifier
	clients map[string]client
}

// client represents a chatter with a given name and identifer.
type client struct {
	ID   string
	Name string
	Addr *net.UDPAddr
}

// message represents a message exchanged between users and it also holds some other
// metadate related to the message.
type message struct {
	ID        string
	Name      string
	Content   string
	ToID      string
	From      *net.UDPAddr
	To        *net.UDPAddr
	Timestamp int64
	Action    int
}

func (s *server) handleMessage() {
	var buffer [8192]byte
	size, addr, err := s.conn.ReadFromUDP(buffer[0:])
	if err != nil {
		return
	}

	msg := buffer[0:size]

	var m message
	if err := json.Unmarshal(msg, &m); err != nil {
		log.Printf("error unmarshaling request json")
	}

	// set the message information
	m.Timestamp = time.Now().Unix()

	fmt.Println(m.Content)

	// handle different actions
	switch m.Action {
	case 0:
		// create new user
		var c client
		c.Addr = addr
		c.ID = m.Name
		c.Name = m.Name
		m.ToID = ""

		s.clients[c.ID] = c
		m.ID = c.ID

		// send data from the server
		m.Name = "server"
		s.messageChannel <- m
	case 1:
		m.From = addr
		// find the user to send the message to
		to, _ := s.clients[m.ToID]
		m.To = to.Addr
		s.messageChannel <- m
	case 2:
		// we want to only send it back to the user
		m.From = addr
		m.Name = "server"

		var toSend string
		for name := range s.clients {
			toSend += fmt.Sprintf("name: %s\n", name)
		}
		fmt.Println(toSend)
		m.Content = toSend
		s.messageChannel <- m
	case 3:
		// relay the message that someone has disconnected
		delete(s.clients, m.ID)
		s.messageChannel <- m
	default:
		log.Println("unsupported message action")
	}

}

func (s *server) handleMessageSend() {
	for {
		msg := <-s.messageChannel
		data, _ := json.Marshal(msg)
		if msg.To != nil {
			s.conn.WriteToUDP(data, msg.To)
		} else {
			// we don't want to echo back the messages of a user
			if msg.Name != "server" {
				continue
			}
			s.conn.WriteToUDP(data, msg.From)
		}
	}
}

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
