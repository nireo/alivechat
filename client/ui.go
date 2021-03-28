package main

import (
	"log"

	"github.com/marcusolsson/tui-go"
)

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
		peerMessage, err := conv.Send([]byte(e.Text()))
		if err != nil {
			message := "[ERROR] cannot send otr message"
			messageArea.Append(tui.NewHBox(tui.NewLabel(message), tui.NewSpacer()))
		}

		for _, msg := range peerMessage {
			c.sendMsg <- string(msg)
		}

		messageArea.Append(tui.NewHBox(tui.NewLabel(e.Text()), tui.NewSpacer()))
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
