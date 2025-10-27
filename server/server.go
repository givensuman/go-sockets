// Package server implements a WebSocket server that manages client connections
package server

import (
	"net/http"
	"sync"

	gosock "github.com/givensuman/go-sockets"
	"github.com/givensuman/go-sockets/internal/emitter"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type Server struct {
	emitter.EventEmitter
	upgrader websocket.Upgrader
	sockets  sync.Map // map[string]*Socket
}

func NewServer() *Server {
	return &Server{
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		},
	}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "upgrade failed", http.StatusBadRequest)
		return
	}

	id := uuid.New().String()
	socket := &Socket{
		EventEmitter: emitter.EventEmitter{},
		Conn:         conn,
		ID:           id,
		writeChan:    make(chan gosock.Packet, 10),
	}
	s.sockets.Store(id, socket)

	go socket.readLoop()
	go socket.writeLoop()

	s.Emit("connection", socket)
}
