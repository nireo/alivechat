package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"time"

	"github.com/marcusolsson/tui-go"
	"golang.org/x/crypto/otr"
)

type client struct {
	conn  *net.UDPConn
	ID    string
	name  string
	alive bool

	sendMsg chan string
	recMsg  chan string

	toID string

	msgBox *tui.Box
	ui     tui.UI
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
	var buf [8192]byte
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

func (c *client) writeMsgToUI(msg string) {
	c.ui.Update(func() {
		c.msgBox.Append(tui.NewHBox(tui.NewLabel(msg), tui.NewSpacer()))
	})
}

func (c *client) startInterface() {
	messageArea := tui.NewVBox()
	messageAreaScroll := tui.NewScrollArea(messageArea)
	messageAreaScroll.SetAutoscrollToBottom(true)
	messageAreaBox := tui.NewVBox(messageAreaScroll)
	messageAreaBox.SetBorder(true)

	input := tui.NewEntry()
	input.SetFocused(true)
	input.SetSizePolicy(tui.Expanding, tui.Maximum)

	inputBox := tui.NewHBox(input)
	inputBox.SetBorder(true)
	inputBox.SetSizePolicy(tui.Expanding, tui.Maximum)

	chat := tui.NewVBox(messageAreaBox, inputBox)
	chat.SetSizePolicy(tui.Expanding, tui.Expanding)

	input.OnSubmit(func(e *tui.Entry) {
		msg := e.Text()
		if msg == "/list" {
			c.sendMessage(2, "")
			input.SetText("")
			return
		} else if msg == "/quit" {
			os.Exit(0)
		} else if strings.HasPrefix(msg, "/connect") {
			name := strings.Split(msg, " ")
			c.toID = strings.Join(name[1:], " ")
			c.sendMsg <- otr.QueryMessage
		}
		peerMessage, err := conv.Send([]byte(e.Text()))
		if err != nil {
			message := "[ERROR] cannot send otr message"
			messageArea.Append(tui.NewHBox(tui.NewLabel(message), tui.NewSpacer()))
		}

		for _, m := range peerMessage {
			c.sendMsg <- string(m)
		}

		mesg := fmt.Sprintf("[%s] <%s> %s", formatTimestamp(time.Now().Unix()), c.name, e.Text())
		messageArea.Append(tui.NewHBox(tui.NewLabel(mesg), tui.NewSpacer()))
		input.SetText("")
	})

	root := tui.NewHBox(chat)

	ui, err := tui.New(root)
	if err != nil {
		log.Fatal(err)
	}

	ui.SetKeybinding("Esc", func() {
		ui.Quit()
		c.sendMessage(3, c.name+" has disconnected")
	})

	c.msgBox = messageArea
	c.ui = ui
}

func (c *client) print() {
	for c.alive {
		msg := <-c.recMsg
		var m message
		json.Unmarshal([]byte(msg), &m)

		if m.Action == 1 {
			content, peerMessages := parseContent(&m)
			c.writeMsgToUI(content)
			if len(peerMessages) > 0 {
				for _, msg := range peerMessages {
					c.sendMsg <- string(msg)
				}
			}
		} else if m.Action == 2 {
			c.writeMsgToUI(fmt.Sprintf("<%s> %s\n", m.Name, m.Content))
		} else if m.Action == 3 {
			c.writeMsgToUI(fmt.Sprintf("<%s> %s\n", m.Name, m.Content))
		} else if m.Action == 0 {
			c.writeMsgToUI(fmt.Sprintf("<%s> %s\n", m.Name, m.Content))
		}
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

	generatePrivateKey()
	c.sendMsg = make(chan string)
	c.recMsg = make(chan string)

	c.conn, err = net.DialUDP("udp", nil, uAddr)
	if err != nil {
		log.Fatalf("error establishing connection to the server")
	}
	defer c.conn.Close()

	r := bufio.NewScanner(os.Stdin)
	fmt.Print("username: ")
	r.Scan()
	name := r.Text()
	c.name = name

	// this makes sure the user is removed from the user's list even
	// if they close the chat using Ctrl-C
	handleForcedClose(&c)

	// 0=new user joined
	c.sendMessage(0, name+" has joined")

	go c.print()
	go c.receiveMessage()
	go c.sendMessageChannel()

	// display the current users
	c.sendMessage(2, "")

	c.startInterface()
	if err := c.ui.Run(); err != nil {
		log.Fatalf("error staring ui: %s", err)
	}
}
