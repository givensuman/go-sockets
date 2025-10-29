package server

import (
	"sync"

	"github.com/givensuman/go-sockets/internal/emitter"
)

// Namespace represents a Socket.IO namespace, managing sockets and rooms within it.
type Namespace struct {
	emitter.EventEmitter
	name    string
	sockets sync.Map // map[string]*Socket
	rooms   sync.Map // map[string]sync.Map // roomName -> socketID -> true
}

// To creates a BroadcastOperator for broadcasting to all sockets in the specified room.
func (ns *Namespace) To(room string) *BroadcastOperator {
	var targets []string
	if roomMap, ok := ns.rooms.Load(room); ok {
		roomMap.(*sync.Map).Range(func(key, value any) bool {
			targets = append(targets, key.(string))
			return true
		})
	}

	return &BroadcastOperator{
		namespace: ns,
		targets:   targets,
	}
}
