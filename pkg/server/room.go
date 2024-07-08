package server

import (
	"fmt"
	"log/slog"
	"net"
)

type Person struct {
	conn  net.Conn
	name  string
	msgCh chan string
	room  string
}

func makePerson(conn net.Conn, currCount *int) *Person {
	*currCount++
	name := fmt.Sprintf("Person%d", *currCount)
	person := &Person{
		conn:  conn,
		name:  name,
		msgCh: make(chan string),
	}
	person.joinRoom(defaultRoom)
	_, err := conn.Write([]byte("Hello " + person.name + "\n" + greeting))
	if err != nil {
		slog.Error("Couldn't write greeting", "err", err)
	}
	return person
}

type Room struct {
	name   string
	chatCh chan string
	people []*Person
}

func makeRoom(name string) *Room {
	room := &Room{
		name:   name,
		chatCh: make(chan string),
	}

	slog.Info("Created new room", "name", name)

	go func(r *Room) {
		for {
			out := <-r.chatCh
			for _, p := range r.people {
				p.msgCh <- out
			}
		}
	}(room)

	return room
}
