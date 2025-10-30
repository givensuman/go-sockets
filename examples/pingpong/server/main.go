package main

import (
	"log"
	"net/http"

	sockets "github.com/givensuman/go-sockets/server"
)

func main() {
	server := sockets.NewServer()
	io := server.Of("/")

	io.On("connection", func(s *sockets.Socket) {
		log.Printf("Client %s connected", s.ID)

		s.On("ping", func() {
			log.Printf("Received ping from %s", s.ID)
			s.Emit("pong")
		})

		s.On("disconnect", func() {
			log.Printf("Client %s disconnected", s.ID)
		})
	})

	log.Println("Ping/Pong server starting on :3000")
	log.Fatal(http.ListenAndServe(":3000", server))
}
