package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func handleForcedClose(cli *client) {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Println("\rgoodbye...")
		// relay the message that the user has disconnected
		cli.sendMessage(3, cli.name+" has disconnected")
		os.Exit(0)
	}()
}

func formatTimestamp(timestamp int64) string {
	tt := time.Unix(timestamp, 0)
	return fmt.Sprintf("%d:%d:%d", tt.Hour(), tt.Minute(), tt.Second())
}
