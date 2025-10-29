// Package server implements a WebSocket server for Socket.IO connections.
// It manages namespaces, rooms, and client sockets with event-based communication.
package server

import (
	"net/http"
	"sync"

	"github.com/givensuman/go-sockets"
	"github.com/givensuman/go-sockets/internal/emitter"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// Server is the main Socket.IO server that handles WebSocket upgrades and manages namespaces.
type Server struct {
	emitter.EventEmitter
	upgrader   websocket.Upgrader
	namespaces sync.Map // map[string]*Namespace
}

// NewServer creates a new Socket.IO server with default WebSocket upgrader settings.
func NewServer() *Server {
	return &Server{
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		},
	}
}

// Of returns the namespace for the given path, creating it if it doesn't exist.
// If path is empty, it defaults to "/".
func (s *Server) Of(path string) *Namespace {
	if path == "" {
		path = "/"
	}

	if ns, ok := s.namespaces.Load(path); ok {
		return ns.(*Namespace)
	}

	ns := &Namespace{
		name: path,
	}

	s.namespaces.Store(path, ns)
	return ns
}

// ServeHTTP handles HTTP requests, upgrading them to WebSocket connections for Socket.IO.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "upgrade failed", http.StatusBadRequest)
		return
	}

	namespace := r.URL.Path
	if namespace == "" {
		namespace = "/"
	}
	ns := s.Of(namespace)

	id := uuid.New().String()
	socket := &Socket{
		EventEmitter: emitter.EventEmitter{},
		Conn:         conn,
		ID:           id,
		writeChan:    make(chan sockets.Packet, 10),
		Namespace:    ns,
	}
	ns.sockets.Store(id, socket)

	// Add default handlers for join/leave
	socket.On("join", func(room string) {
		socket.Join(room)
	})
	socket.On("leave", func(room string) {
		socket.Leave(room)
	})

	go socket.readLoop()
	go socket.writeLoop()

	ns.Emit("connection", socket)
}
