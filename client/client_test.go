package client

import (
	"net/http"
	"testing"
	"time"

	srv "github.com/givensuman/go-sockets/server"
)

func TestClientE2E(t *testing.T) {
	// Start server
	server := srv.NewServer()
	httpServer := &http.Server{
		Addr:    ":8081",
		Handler: server,
	}
	go httpServer.ListenAndServe()
	defer httpServer.Close()

	time.Sleep(100 * time.Millisecond) // Wait for server to start

	// Register handler on server
	ns := server.Of("/")
	ns.On("connection", func(s *srv.Socket) {
		s.On("ping", func() {
			s.Emit("pong")
		})
	})

	// Connect with client
	connectReceived := make(chan bool, 1)
	clientSocket, err := Connect("ws://localhost:8081", "/", func(s *Socket) {
		s.On("connect", func() {
			connectReceived <- true
		})
	})
	if err != nil {
		t.Fatal(err)
	}
	defer clientSocket.Close()

	select {
	case <-connectReceived:
	case <-time.After(1 * time.Second):
		t.Fatal("connect event not received")
	}

	// Send ping
	pongReceived := make(chan bool, 1)
	clientSocket.On("pong", func() {
		pongReceived <- true
	})
	clientSocket.Emit("ping")

	select {
	case <-pongReceived:
	case <-time.After(1 * time.Second):
		t.Fatal("pong event not received")
	}
}

func TestClientAckE2E(t *testing.T) {
	// Start server
	server := srv.NewServer()
	httpServer := &http.Server{
		Addr:    ":8082",
		Handler: server,
	}
	go httpServer.ListenAndServe()
	defer httpServer.Close()

	time.Sleep(100 * time.Millisecond) // Wait for server to start

	// Register handler on server
	ns := server.Of("/")
	ns.On("connection", func(s *srv.Socket) {
		s.On("get_data", func(d string, ack func(string)) {
			ack("echo:" + d)
		})
	})

	// Connect with client
	connectReceived := make(chan bool, 1)
	clientSocket, err := Connect("ws://localhost:8082", "/", func(s *Socket) {
		s.On("connect", func() {
			connectReceived <- true
		})
	})
	if err != nil {
		t.Fatal(err)
	}
	defer clientSocket.Close()

	select {
	case <-connectReceived:
	case <-time.After(1 * time.Second):
		t.Fatal("connect event not received")
	}

	// Send get_data with ack
	responseReceived := make(chan string, 1)
	clientSocket.Emit("get_data", "foo", func(response string) {
		responseReceived <- response
	})

	select {
	case resp := <-responseReceived:
		if resp != "echo:foo" {
			t.Errorf("expected 'echo:foo', got %s", resp)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("ack response not received")
	}
}
