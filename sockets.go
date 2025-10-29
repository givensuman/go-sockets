// Package sockets provides structures and constants for socket communication packets.
// It defines the core types used for encoding and decoding messages in the Socket.IO protocol.
package sockets

import (
	"encoding/json"
	"log"
)

// PacketType represents the type of a Socket.IO packet.
type PacketType int

// Packet type constants as defined in the Socket.IO protocol.
const (
	// Connect is sent when a client connects to a namespace.
	Connect PacketType = iota
	// Disconnect is sent when a client disconnects from a namespace.
	Disconnect
	// Event is a regular event packet containing an event name and data.
	Event
	// Ack is an acknowledgment packet sent in response to an event with an ID.
	Ack
	// Error indicates an error occurred.
	Error
	// BinaryEvent is an event packet with binary data.
	BinaryEvent
	// BinaryAck is an acknowledgment packet with binary data.
	BinaryAck
)

// Packet represents a Socket.IO protocol packet.
type Packet struct {
	// Type is the packet type (e.g., Event, Ack).
	Type PacketType
	// Namespace is the namespace for the packet (e.g., "/", "/chat").
	Namespace string
	// Data contains the packet payload as JSON raw message.
	Data json.RawMessage
	// ID is an optional packet ID used for acknowledgments.
	ID *uint64
}

// GetEventName extracts the event name from an Event or BinaryEvent packet.
// It returns the event name and true if successful, or nil and false otherwise.
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

// GetEventArgs extracts the event arguments from an Event or BinaryEvent packet.
// It returns the arguments slice and true if successful, or nil and false otherwise.
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
