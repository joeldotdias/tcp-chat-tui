package main

import (
	"log"

	"github.com/joeldotdias/tcp-chat-tui/pkg/server"
)

func main() {
	server := server.InitServer()
	if err := server.Run(); err != nil {
		log.Fatal(err)
	}
}
