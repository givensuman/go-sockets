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
		log.Printf("User %s connected", s.ID)

		s.On("join", func(room string) {
			s.Join(room)
			s.Emit("message", "You joined "+room)
			s.Broadcast().To(room).Emit("message", "User "+s.ID+" joined "+room)
		})

		s.On("leave", func(room string) {
			s.Leave(room)
			s.Emit("message", "You left "+room)
			s.Broadcast().To(room).Emit("message", "User "+s.ID+" left "+room)
		})

		s.On("chat message", func(msg string) {
			log.Printf("Message from %s: %s", s.ID, msg)
			// Broadcast to all clients in the same rooms
			s.Broadcast().Emit("chat message", s.ID, msg)
		})

		s.On("disconnect", func() {
			log.Printf("User %s disconnected", s.ID)
		})
	})

	log.Println("Server starting on :3000")
	log.Fatal(http.ListenAndServe(":3000", server))
}
