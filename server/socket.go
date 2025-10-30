package server

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

// Socket represents a server-side WebSocket connection to a client.
// It embeds EventEmitter for event handling and manages acknowledgments.
type Socket struct {
	emitter.EventEmitter
	Conn       *websocket.Conn
	ID         string
	writeChan  chan sockets.Packet
	closeOnce  sync.Once
	Namespace  *Namespace
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
		case sockets.Connect:
			s.EventEmitter.Emit("connect")
			connectPacket := sockets.Packet{
				Type:      sockets.Connect,
				Namespace: packet.Namespace,
			}
			select {
			case s.writeChan <- connectPacket:
			default:
				s.Close()
			}

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
				ackFuncValue := reflect.ValueOf(func(args ...any) {
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
				})

				callbackType := s.GetCallbackType(*eventName)
				if callbackType != nil && callbackType.NumIn() > 1 {
					ackType := callbackType.In(callbackType.NumIn() - 1)
					ackValue := reflect.MakeFunc(ackType, func(in []reflect.Value) []reflect.Value {
						args := make([]any, len(in))
						for i, v := range in {
							args[i] = v.Interface()
						}
						callArgs := make([]reflect.Value, len(args))
						for i, arg := range args {
							callArgs[i] = reflect.ValueOf(arg)
						}
						ackFuncValue.Call(callArgs)
						return nil
					})

					eventArgs = append(eventArgs, ackValue.Interface())
				}
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
			s.EventEmitter.Emit("disconnect", "client request")
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

// Emit sends an event to the client with optional arguments.
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
		Namespace: s.Namespace.name,
		ID:        ackID,
	}
	select {
	case s.writeChan <- packet:
		return
	default:
		s.Close()
	}
}

// Join adds the socket to the specified room in its namespace.
func (s *Socket) Join(room string) {
	if roomMap, ok := s.Namespace.rooms.Load(room); ok {
		roomMap.(*sync.Map).Store(s.ID, true)
	} else {
		roomMap := &sync.Map{}
		roomMap.Store(s.ID, true)
		s.Namespace.rooms.Store(room, roomMap)
	}
}

// Leave removes the socket from the specified room in its namespace.
func (s *Socket) Leave(room string) {
	if roomMap, ok := s.Namespace.rooms.Load(room); ok {
		roomMap.(*sync.Map).Delete(s.ID)
	}
}

// Broadcast returns a BroadcastOperator for sending events to other sockets in the namespace.
func (s *Socket) Broadcast() *BroadcastOperator {
	var targets []string
	s.Namespace.sockets.Range(func(key, value any) bool {
		id := key.(string)
		if id != s.ID {
			targets = append(targets, id)
		}
		return true
	})

	return &BroadcastOperator{
		namespace: s.Namespace,
		targets:   targets,
	}
}

// Close closes the WebSocket connection and cleans up resources.
func (s *Socket) Close() {
	s.closeOnce.Do(func() {
		s.Conn.Close()
		close(s.writeChan)
	})
}
