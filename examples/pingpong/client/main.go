package main

import (
	"log"

	sockets "github.com/givensuman/go-sockets/client"
)

func main() {
	socket, err := sockets.Connect("ws://localhost:3000", "/", nil)
	if err != nil {
		log.Fatal("Failed to connect:", err)
	}
	defer socket.Close()

	socket.On("connect", func() {
		log.Println("Connected to server")
		socket.Emit("ping")
	})

	socket.On("pong", func() {
		log.Println("Received pong from server")
		socket.Close()
	})

	// Keep the connection alive
	select {}
}
