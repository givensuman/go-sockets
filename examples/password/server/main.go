package main

import (
	"log"
	"net/http"

	sockets "github.com/givensuman/go-sockets/server"
)

func main() {
	server := sockets.NewServer()

	// Root namespace for regular users with chat functionality
	io := server.Of("/")
	io.On("connection", func(s *sockets.Socket) {
		log.Printf("User %s connected to root", s.ID)

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
			log.Printf("User %s disconnected from root", s.ID)
		})
	})

	// Admin namespace with password protection
	admin := server.Of("/admin")
	admin.On("connection", func(s *sockets.Socket) {
		log.Printf("User %s attempting to connect to admin", s.ID)

		authenticated := false

		s.On("auth", func(password string) {
			if password == "admin123" {
				authenticated = true
				s.Emit("authenticated", "Welcome to admin panel")
				log.Printf("User %s authenticated for admin", s.ID)
			} else {
				s.Emit("auth_failed", "Invalid password")
				s.Close()
			}
		})

		s.On("admin_broadcast", func(msg string) {
			if authenticated {
				admin.Emit("admin_message", s.ID, msg)
			} else {
				s.Emit("error", "Not authenticated")
			}
		})

		s.On("disconnect", func() {
			log.Printf("User %s disconnected from admin", s.ID)
		})
	})

	log.Println("Server starting on :3000")
	log.Fatal(http.ListenAndServe(":3000", server))
}
