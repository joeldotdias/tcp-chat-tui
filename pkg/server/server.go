package server

import (
	"bufio"
	"fmt"
	"log/slog"
	"net"
	"slices"
	"strings"
)

const helpStr = `HELP (All commands must start with a ":")
"quit"      Quit the server
"rooms"     List all rooms
"create"    Create a new room
"join"      Leave current room and join a new one
"people"    List all the people in your room
"rename"    Change your name
"help"      Bring up the help menu
`

const (
	defaultRoom = "zero"
	greeting    = `You are now in the default chat room Zero.
Type ":help" to see the list of commands.
`
)

var rooms = make(map[string]*Room)

type Server struct {
	listener net.Listener
	count    int
}

func NewServer() *Server {
	rooms[defaultRoom] = makeRoom(defaultRoom)
	return &Server{
		count: 0,
	}
}

func (s *Server) Run() error {
	listener, err := net.Listen("tcp", ":6969")
	if err != nil {
		return err
	}
	s.listener = listener
	slog.Info("Listening on", "addr", "6969")
	defer listener.Close()

	for {
		conn, err := s.listener.Accept()
		if err != nil {
			return err
		}

		go s.handleConn(conn)
	}
}

func (s *Server) handleConn(conn net.Conn) {
	slog.Info("Incoming connection", "addr", conn.RemoteAddr())
	person := makePerson(conn, &s.count)

	// send messages
	go func(p *Person) {
		for {
			buffer, err := bufio.NewReader(p.conn).ReadString('\n')
			if err != nil {
				slog.Error("Couldn't read message", "err", err)
				return
			}
			buffer = strings.Trim(buffer, "\r\n")
			if strings.HasPrefix(buffer, ":") {
				shouldQuit, err := p.handleCmd(buffer)
				if err != nil {
					slog.Error(err.Error())
				}
				if shouldQuit {
					break
				}
			} else if strings.HasPrefix(buffer, "@") {
				// private message
				parts := strings.SplitN(buffer, " ", 2)
				receiver := strings.Trim(parts[0], "@")
				room := rooms[p.room]
				idx := slices.IndexFunc(room.people, func(ps *Person) bool {
					return ps.name == receiver
				})
				if idx == -1 {
					_, err := p.conn.Write([]byte(fmt.Sprintf("\"%s\" is not present in this room. Type \":people\" to list all people in this room\n", receiver)))
					slog.Error("Didn't find user", "err", err)
					return
				}
				sendTo := room.people[idx]
				sendTo.msgCh <- "@" + p.name + ": " + parts[1] + "\n"
			} else {
				// normal message
				msg := p.name + ": " + buffer + "\n"
				room, exists := rooms[p.room]
				if !exists {
					slog.Error("How would this even happen", "err", err)
				}
				room.chatCh <- msg

			}
		}
	}(person)

	// get messages
	go func(p *Person) {
		for {
			msg := <-p.msgCh
			if strings.HasPrefix(msg, "*** "+person.name) {
				continue
			}
			if strings.HasPrefix(msg, p.name+":") {
				msg = strings.Replace(msg, p.name, "You", 1)
			}
			_, err := p.conn.Write([]byte(msg))
			if err != nil {
				slog.Error("Couldn't to send message", "err", err)
			}
		}
	}(person)
}

func (p *Person) handleCmd(cmd string) (bool, error) {
	shouldQuit := false // set to true if cmd is ":quit"
	parts := strings.Split(cmd, " ")

	switch parts[0] {
	case ":quit":
		slog.Info(p.name + " has left the chat")
		p.leaveRoom()
		p.conn.Close()
		shouldQuit = true

	case ":help":
		_, err := p.conn.Write([]byte(helpStr))
		if err != nil {
			return shouldQuit, err
		}
	case ":rooms":
		idx := 0
		for _, room := range rooms {
			idx++
			pplLen := len(room.people)
			var nPpl string
			if pplLen == 1 {
				nPpl = "1 person"
			} else {
				nPpl = fmt.Sprintf("%d people", pplLen)
			}
			_, err := p.conn.Write([]byte(fmt.Sprintf("%d. %s (%s)\n", idx, room.name, nPpl)))
			if err != nil {
				return shouldQuit, err
			}
		}
	case ":join":
		p.joinRoom(parts[1])
	case ":create":
		p.createRoom(parts[1])
	case ":rename":
		p.name = parts[1]
		_, err := p.conn.Write([]byte("You are now " + p.name + "\n"))
		if err != nil {
			return shouldQuit, err
		}
	case ":people":
		idx := 0
		for _, ps := range rooms[p.room].people {
			idx++
			_, err := p.conn.Write([]byte(fmt.Sprintf("%d. %s\n", idx, ps.name)))
			if err != nil {
				return shouldQuit, err
			}
		}
	default:
		_, err := p.conn.Write([]byte(fmt.Sprintf("\"%s\" is not a valid command. Type \":help\" to list all commands\n", strings.TrimPrefix(parts[0], ":"))))
		if err != nil {
			return shouldQuit, err
		}
	}

	return shouldQuit, nil
}

func (p *Person) createRoom(roomName string) {
	room := makeRoom(roomName)
	rooms[room.name] = room
	p.joinRoom(roomName)
}

func (p *Person) joinRoom(roomName string) {
	room, exists := rooms[roomName]
	if !exists {
		_, err := p.conn.Write([]byte(fmt.Sprintf("\"%s\" does not exist. Type \":rooms\" to list all rooms\n", roomName)))
		if err != nil {
			slog.Error("Couldn't write message", "err", err)
		}
		return
	}

	room.people = append(room.people, p)

	if p.room != "" {
		p.leaveRoom()
	}

	p.room = roomName
	_, err := p.conn.Write([]byte("Welcome to " + roomName + "\n"))
	if err != nil {
		slog.Error("Couldn't write message", "err", err)
	}
	room.chatCh <- "*** " + p.name + " has joined the room ***\n"
}

func (p *Person) leaveRoom() {
	prevRoom := rooms[p.room]
	idx := slices.Index(prevRoom.people, p)
	prevRoom.people = slices.Delete(prevRoom.people, idx, idx+1)
	prevRoom.chatCh <- "*** " + p.name + " has left the room ***\n"
	slog.Info(p.name + " is leaving " + prevRoom.name + "\n")
}
