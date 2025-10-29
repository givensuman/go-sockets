package server

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/givensuman/go-sockets"
	"github.com/givensuman/go-sockets/internal/parser"
	"github.com/gorilla/websocket"
)

func TestServerE2E(t *testing.T) {
	server := NewServer()
	httpServer := &http.Server{
		Addr:    ":8080",
		Handler: server,
	}
	go httpServer.ListenAndServe()
	defer httpServer.Close()
	time.Sleep(100 * time.Millisecond) // Wait for server to start

	// Register handler
	ns := server.Of("/")
	ns.On("connection", func(s *Socket) {
		s.On("ping", func() {
			s.Emit("pong")
		})
	})

	// Connect
	url := "ws://localhost:8080"
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	// Send ping
	pingPacket := sockets.Packet{
		Type: sockets.Event,
		Data: json.RawMessage(`["ping"]`),
	}
	pingData := parser.Encode(pingPacket)
	err = conn.WriteMessage(websocket.TextMessage, pingData)
	if err != nil {
		t.Fatal(err)
	}

	// Read pong
	conn.SetReadDeadline(time.Now().Add(1 * time.Second))
	_, pongData, err := conn.ReadMessage()
	if err != nil {
		t.Fatal(err)
	}
	pongPacket, err := parser.Decode(pongData)
	if err != nil {
		t.Fatal(err)
	}
	if pongPacket.Type != sockets.Event {
		t.Error("expected EVENT packet")
	}
	var pongEvent []any
	json.Unmarshal(pongPacket.Data, &pongEvent)
	if len(pongEvent) == 0 || pongEvent[0] != "pong" {
		t.Error("expected pong event")
	}
}

func TestRoomsAndBroadcasting(t *testing.T) {
	server := NewServer()
	httpServer := &http.Server{
		Addr:    ":8081",
		Handler: server,
	}
	go httpServer.ListenAndServe()
	defer httpServer.Close()
	time.Sleep(500 * time.Millisecond)

	// Register handler on default namespace
	ns := server.Of("/")
	ns.On("connection", func(s *Socket) {
		s.On("msg", func(m string) {
			s.Broadcast().To("room1").Emit("broadcast", m)
		})
	})

	// Connect client A
	connA, _, err := websocket.DefaultDialer.Dial("ws://localhost:8081/", nil)
	if err != nil {
		t.Fatal(err)
	}
	defer connA.Close()

	// Send join for client A
	joinPacket := sockets.Packet{
		Type: sockets.Event,
		Data: json.RawMessage(`["join", "room1"]`),
	}
	joinData := parser.Encode(joinPacket)
	connA.WriteMessage(websocket.TextMessage, joinData)

	// Connect client B
	connB, _, err := websocket.DefaultDialer.Dial("ws://localhost:8081/", nil)
	if err != nil {
		t.Fatal(err)
	}
	defer connB.Close()

	// Send join for client B
	connB.WriteMessage(websocket.TextMessage, joinData)

	// Connect client C
	connC, _, err := websocket.DefaultDialer.Dial("ws://localhost:8081/", nil)
	if err != nil {
		t.Fatal(err)
	}
	defer connC.Close()

	// Send join for client C to room2
	joinPacketC := sockets.Packet{
		Type: sockets.Event,
		Data: json.RawMessage(`["join", "room2"]`),
	}
	joinDataC := parser.Encode(joinPacketC)
	connC.WriteMessage(websocket.TextMessage, joinDataC)

	// Channels to receive messages
	recvA := make(chan string, 1)
	recvB := make(chan string, 1)
	recvC := make(chan string, 1)

	// Goroutines to read
	go func() {
		connA.SetReadDeadline(time.Now().Add(2 * time.Second))
		_, data, err := connA.ReadMessage()
		if err == nil {
			packet, _ := parser.Decode(data)
			if packet.Type == sockets.Event {
				var event []any
				json.Unmarshal(packet.Data, &event)
				if len(event) > 0 && event[0] == "broadcast" {
					recvA <- event[1].(string)
				}
			}
		}
	}()

	go func() {
		connB.SetReadDeadline(time.Now().Add(2 * time.Second))
		_, data, err := connB.ReadMessage()
		if err == nil {
			packet, _ := parser.Decode(data)
			if packet.Type == sockets.Event {
				var event []any
				json.Unmarshal(packet.Data, &event)
				if len(event) > 0 && event[0] == "broadcast" {
					recvB <- event[1].(string)
				}
			}
		}
	}()

	go func() {
		connC.SetReadDeadline(time.Now().Add(2 * time.Second))
		_, data, err := connC.ReadMessage()
		if err == nil {
			packet, _ := parser.Decode(data)
			if packet.Type == sockets.Event {
				var event []any
				json.Unmarshal(packet.Data, &event)
				if len(event) > 0 && event[0] == "broadcast" {
					recvC <- event[1].(string)
				}
			}
		}
	}()

	// Client A sends msg
	msgPacket := sockets.Packet{
		Type: sockets.Event,
		Data: json.RawMessage(`["msg", "hello"]`),
	}
	msgData := parser.Encode(msgPacket)
	connA.WriteMessage(websocket.TextMessage, msgData)

	// Wait a bit
	time.Sleep(500 * time.Millisecond)

	// Check receives
	select {
	case <-recvA:
		t.Error("Client A should not receive broadcast")
	default:
	}

	select {
	case msg := <-recvB:
		if msg != "hello" {
			t.Error("Client B should receive 'hello'")
		}
	default:
		t.Error("Client B should receive broadcast")
	}

	select {
	case <-recvC:
		t.Error("Client C should not receive broadcast")
	default:
	}
}
