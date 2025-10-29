// Package server implements a WebSocket server for Socket.IO connections.
// It manages namespaces, rooms, and client sockets with event-based communication.
package server

import (
	"encoding/json"
	"net/http"
	"sync"

	gosock "github.com/givensuman/go-sockets"
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

// Namespace represents a Socket.IO namespace, managing sockets and rooms within it.
type Namespace struct {
	emitter.EventEmitter
	name    string
	sockets sync.Map // map[string]*Socket
	rooms   sync.Map // map[string]sync.Map // roomName -> socketID -> true
}

// BroadcastOperator is used to broadcast events to multiple sockets, optionally filtered by rooms.
type BroadcastOperator struct {
	namespace *Namespace
	targets   []string
}

// To filters the broadcast targets to only sockets in the specified room.
func (bo *BroadcastOperator) To(room string) *BroadcastOperator {
	var newTargets []string
	if roomMap, ok := bo.namespace.rooms.Load(room); ok {
		for _, id := range bo.targets {
			if _, inRoom := roomMap.(*sync.Map).Load(id); inRoom {
				newTargets = append(newTargets, id)
			}
		}
	}
	return &BroadcastOperator{
		namespace: bo.namespace,
		targets:   newTargets,
	}
}

// Emit broadcasts an event to all targets in the BroadcastOperator.
func (bo *BroadcastOperator) Emit(event string, args ...interface{}) {
	eventData := append([]interface{}{event}, args...)
	data, _ := json.Marshal(eventData)
	packet := gosock.Packet{
		Type:      gosock.Event,
		Data:      json.RawMessage(data),
		Namespace: bo.namespace.name,
	}
	for _, id := range bo.targets {
		if sock, ok := bo.namespace.sockets.Load(id); ok {
			select {
			case sock.(*Socket).writeChan <- packet:
			default:
				// channel full, skip
			}
		}
	}
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

// To creates a BroadcastOperator for broadcasting to all sockets in the specified room.
func (ns *Namespace) To(room string) *BroadcastOperator {
	var targets []string
	if roomMap, ok := ns.rooms.Load(room); ok {
		roomMap.(*sync.Map).Range(func(key, value interface{}) bool {
			targets = append(targets, key.(string))
			return true
		})
	}
	return &BroadcastOperator{
		namespace: ns,
		targets:   targets,
	}
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
		writeChan:    make(chan gosock.Packet, 10),
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
