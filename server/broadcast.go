package server

import (
	"encoding/json"
	"sync"

	"github.com/givensuman/go-sockets"
)

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
func (bo *BroadcastOperator) Emit(event string, args ...any) {
	eventData := append([]any{event}, args...)
	data, _ := json.Marshal(eventData)
	packet := sockets.Packet{
		Type:      sockets.Event,
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

