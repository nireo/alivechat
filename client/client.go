package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
)

type client struct {
	conn *net.UDPConn
	ID string
	name string
	alive bool

	sendMsg chan string
	recMsg chan string

	toID string
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

func (c *client) sendMessage(action int, msg string) {
	m := message{
		Content: msg,
		Name: c.name,
		Action: action,
	}

	data, err := json.Marshal(m)
	if err != nil {
		log.Printf("error marshing message")
	}

	if _, err := c.conn.Write(data); err != nil {
		log.Printf("error writing string data: %s", err)
	}
}

func (c *client) receiveMessage() {
	var buf [4096]byte
	for c.alive {
		n, err := c.conn.Read(buf[0:])
		if err != nil {
			log.Printf("error receiving message")
		}

		c.recMsg <- string(buf[0:n])
	}
}

func (c *client) sendMessageChannel() {
	for c.alive {
		msg := <- c.sendMsg
		m := message{
			Action: 1,
			Name: c.name,
			ID: c.ID,
			Content: msg,
			ToID: c.toID,
		}
		data, err := json.Marshal(m)
		if err != nil {
			log.Printf("error marshaling json: %s", err)
		}

		if _, err := c.conn.Write(data); err != nil {
			log.Printf("error writing string data: %s", err)
		}
	}
}

func (c *client) getMessage() {
	scanner := bufio.NewScanner(os.Stdin)
	for c.alive {
		fmt.Print(">>> ")
		scanned := scanner.Scan()
		if !scanned {
			continue
		}

		line := scanner.Text()
		if line == "/quit" {
			c.alive = false
			return
		} else if line == "/list" {
			c.sendMessage(2, "")
		} else {
			c.sendMessage(1, line)
		}
	}
}

func (c *client) print() {
	for c.alive {
		msg := <- c.recMsg
		var m message
		json.Unmarshal([]byte(msg), &m)
		fmt.Printf("[%s]: %s\n", m.Name, m.Content)
	}
}

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("please provide an address to connect to")
	}
	uAddr, err := net.ResolveUDPAddr("udp4", os.Args[1])
	if err != nil {
		log.Fatalf("could not resolve udp addr")
	}

	var c client
	c.alive = true
	c.sendMsg = make(chan string)
	c.recMsg = make(chan string)

	c.conn, err = net.DialUDP("udp", nil, uAddr)
	if err != nil {
		log.Fatalf("error establishing connection to the server")
	}
	defer c.conn.Close()

	r := bufio.NewScanner(os.Stdin)
	fmt.Println("username: ")
	r.Scan()
	name := r.Text()
	c.name = name

	// 0=new user joined
	c.sendMessage(0, name+" has joined")

	go c.print()
	go c.receiveMessage()
	go c.sendMessageChannel()

	c.getMessage()
}

