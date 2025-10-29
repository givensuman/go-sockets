package client

import (
	"encoding/json"
	"log"
	"reflect"
	"sync"
	"sync/atomic"
	"time"

	"github.com/givensuman/go-sockets"
	"github.com/givensuman/go-sockets/internal/emitter"
	"github.com/givensuman/go-sockets/internal/parser"
	"github.com/gorilla/websocket"
)

// Socket represents a client-side WebSocket connection to a Socket.IO server.
// It embeds EventEmitter for event handling and manages acknowledgments.
type Socket struct {
	emitter.EventEmitter
	Conn       *websocket.Conn
	writeChan  chan sockets.Packet
	closeOnce  sync.Once
	Namespace  string
	ackCounter uint64
	ackMap     sync.Map // uint64 -> reflect.Value
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

			if packet.ID != nil {
				ackFunc := func(args ...any) {
					ackData, _ := json.Marshal(args)
					ackPacket := sockets.Packet{
						Type:      sockets.Ack,
						Data:      json.RawMessage(ackData),
						Namespace: packet.Namespace,
						ID:        packet.ID,
					}

					select {
					case s.writeChan <- ackPacket:
					default:
						s.Close()
					}
				}

				eventArgs = append(eventArgs, ackFunc)
			}

			s.EventEmitter.Emit(*eventName, eventArgs...)

		case sockets.Ack:
			if packet.ID != nil {
				if callback, ok := s.ackMap.Load(*packet.ID); ok {
					s.ackMap.Delete(*packet.ID)

					var ackArgs []any
					json.Unmarshal(packet.Data, &ackArgs)
					var callArgs []reflect.Value

					for _, arg := range ackArgs {
						callArgs = append(callArgs, reflect.ValueOf(arg))
					}
					
					callback.(reflect.Value).Call(callArgs)
				}
			}

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

// Emit sends an event to the server with optional arguments.
// If the last argument is a function, it sets up an acknowledgment callback.
func (s *Socket) Emit(event string, args ...any) {
	var ackID *uint64
	if len(args) > 0 {
		lastArg := args[len(args)-1]
		if reflect.TypeOf(lastArg).Kind() == reflect.Func {
			id := atomic.AddUint64(&s.ackCounter, 1)
			ackID = &id
			s.ackMap.Store(id, reflect.ValueOf(lastArg))
			// Timeout after 10 seconds
			time.AfterFunc(10*time.Second, func() {
				s.ackMap.Delete(id)
			})
			args = args[:len(args)-1]
		}
	}
	eventData := append([]any{event}, args...)
	data, _ := json.Marshal(eventData)
	packet := sockets.Packet{
		Type:      sockets.Event,
		Data:      json.RawMessage(data),
		Namespace: s.Namespace,
		ID:        ackID,
	}
	select {
	case s.writeChan <- packet:
		return
	default:
		s.Close()
	}
}

// Join sends a "join" event to the server to join the specified room.
func (s *Socket) Join(room string) {
	s.Emit("join", room)
}

// Leave sends a "leave" event to the server to leave the specified room.
func (s *Socket) Leave(room string) {
	s.Emit("leave", room)
}

// Close closes the WebSocket connection and cleans up resources.
func (s *Socket) Close() {
	s.closeOnce.Do(func() {
		s.Conn.Close()
		close(s.writeChan)
	})
}
