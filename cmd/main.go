package main

import (
	"log"

	"github.com/joeldotdias/tcp-chat-tui/pkg/server"
)

func main() {
	server := server.NewServer()
	if err := server.Run(); err != nil {
		log.Fatal(err)
	}
}
