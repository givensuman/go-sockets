// Package sockets provides structures and constants for socket communication packets.
package sockets

import (
	"encoding/json"
	"log"
)

type PacketType int

const (
	Connect PacketType = iota
	Disconnect
	Event
	Ack
	Error
	BinaryEvent
	BinaryAck
)

type Packet struct {
	Type      PacketType
	Namespace string
	Data      json.RawMessage
	ID        *uint64
}

func (p *Packet) GetEventName() (*string, bool) {
	if p.Type != Event && p.Type != BinaryEvent {
		return nil, false
	}

	var eventData []any
	if err := json.Unmarshal(p.Data, &eventData); err != nil {
		log.Println("unmarshal error:", err)
		return nil, false
	}

	eventName, ok := eventData[0].(string)
	if !ok {
		return nil, false
	}

	return &eventName, true
}

func (p *Packet) GetEventArgs() ([]any, bool) {
	if p.Type != Event && p.Type != BinaryEvent {
		return nil, false
	}

	var eventData []any
	if err := json.Unmarshal(p.Data, &eventData); err != nil {
		log.Println("unmarshal error:", err)
		return nil, false
	}

	args := eventData[1:]
	return args, true
}
