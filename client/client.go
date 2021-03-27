package main

import (
	"bufio"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"

	"golang.org/x/crypto/otr"
)

// Instead of implementing the OTR-protocol myself, I just use the one provided by go :)
var (
	privKey   otr.PrivateKey
	conv      otr.Conversation
	secChange otr.SecurityChange
)

type client struct {
	conn  *net.UDPConn
	ID    string
	name  string
	alive bool

	sendMsg chan string
	recMsg  chan string

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
	Action    int
}

func (c *client) sendMessage(action int, msg string) {
	m := message{
		Content: msg,
		Name:    c.name,
		Action:  action,
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
		msg := <-c.sendMsg
		m := message{
			Action:  1,
			Name:    c.name,
			ID:      c.ID,
			Content: msg,
			ToID:    c.toID,
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
		scanned := scanner.Scan()
		if !scanned {
			continue
		}

		line := scanner.Text()
		switch line {
		case "/quit":
			c.alive = false
			return
		case "/list":
			c.sendMessage(2, "")
		case "/connect":
			// connect and chat with another user privately
			fmt.Println("userid ?")
			if _, err := fmt.Scanln(&c.toID); err != nil {
				log.Printf("error reading user id")
				continue
			}
			fmt.Println("making secure tunnel...")
			c.sendMsg <- otr.QueryMessage
		default:
			fmt.Printf("%s: %s\n", c.name, line)
			peerMessage, err := conv.Send([]byte(line))
			if err != nil {
				log.Println("error sending otr message")
				continue
			}

			for _, msg := range peerMessage {
				c.sendMsg <- string(msg)
			}
		}
	}
}

func (c *client) print() {
	for c.alive {
		msg := <-c.recMsg
		var m message
		json.Unmarshal([]byte(msg), &m)

		if m.Action == 1 {
			bytes := []byte(m.Content)
			out, enc, _, mpeer, err := conv.Receive(bytes)
			if err != nil {
				log.Printf("error receiving conversation bytes")
			}

			if len(out) > 0 {
				if !enc {
					log.Println("conversation is not encrypted")
					fmt.Printf("[%s]: %s\n", m.Name, string(out))
				} else {
					fmt.Printf("[%s]: %s\n", m.Name, string(out))
				}
			}

			if len(mpeer) > 0 {
				for _, msg := range mpeer {
					c.sendMsg <- string(msg)
				}
			}
		}

		fmt.Printf("[%s]: %s\n", m.Name, m.Content)
	}
}

func genPrivate() {
	key := new(otr.PrivateKey)
	key.Generate(rand.Reader)
	bytes := key.Serialize(nil)

	parsedBytes, ok := privKey.Parse(bytes)
	if !ok {
		log.Printf("failed parsing private key")
	}

	if len(parsedBytes) > 0 {
		log.Printf("the key buffer is not empty after key ")
	}

	conv.PrivateKey = &privKey
	conv.FragmentSize = 1000
	log.Println("client fingerprint:", conv.PrivateKey.Fingerprint())
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

	genPrivate()
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

	// display the current users
	c.sendMessage(2, "")

	c.getMessage()

	// relay the message that the user has disconnected
	c.sendMessage(3, c.name+" has disconnected")
}
