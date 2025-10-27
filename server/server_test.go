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
	server.On("connection", func(s *Socket) {
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
