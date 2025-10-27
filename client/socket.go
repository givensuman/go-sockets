package client

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/givensuman/go-sockets"
	"github.com/givensuman/go-sockets/internal/emitter"
	"github.com/givensuman/go-sockets/internal/parser"
	"github.com/gorilla/websocket"
)

// Socket represents a client-side WebSocket connection.
type Socket struct {
	emitter.EventEmitter
	Conn      *websocket.Conn
	writeChan chan sockets.Packet
	closeOnce sync.Once
}

func (s *Socket) readLoop() {
	defer s.Close()

	for {
		_, data, err := s.Conn.ReadMessage()
		if err != nil {
			return
		}

		packet, err := parser.Decode(data)
		if err != nil {
			log.Println("decode error:", err)
			continue
		}

		switch packet.Type {
		case sockets.Event:
			eventName, ok := packet.GetEventName()
			if !ok {
				continue
			}


			eventArgs, ok := packet.GetEventArgs()
			if !ok {
				continue
			}

			s.EventEmitter.Emit(*eventName, eventArgs...)
		case sockets.Disconnect:
			s.EventEmitter.Emit("disconnect", "server request")
			return
		}
	}
}

func (s *Socket) writeLoop() {
	for packet := range s.writeChan {
		data := parser.Encode(packet)
		err := s.Conn.WriteMessage(websocket.TextMessage, data)
		if err != nil {
			log.Println("write error:", err)
			return
		}
	}
}

func (s *Socket) Emit(event string, args ...any) {
	eventData := append([]any{event}, args...)
	data, _ := json.Marshal(eventData)
	packet := sockets.Packet{
		Type:      sockets.Event,
		Data:      json.RawMessage(data),
		Namespace: "/",
	}
	select {
	case s.writeChan <- packet:
		return
	default:
		s.Close()
	}
}

func (s *Socket) Close() {
	s.closeOnce.Do(func() {
		s.Conn.Close()
		close(s.writeChan)
	})
}
