package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

	sockets "github.com/givensuman/go-sockets/client"
)

func main() {
	socket, err := sockets.Connect("ws://localhost:3000", "/", nil)
	if err != nil {
		log.Fatal("Failed to connect:", err)
	}
	defer socket.Close()

	socket.On("connect", func() {
		fmt.Println("Connected to server")
		fmt.Println("Commands:")
		fmt.Println("  /join <room>  - Join a room")
		fmt.Println("  /leave <room> - Leave a room")
		fmt.Println("  <message>     - Send chat message")
		fmt.Println("  /quit         - Exit")
	})

	socket.On("message", func(msg string) {
		fmt.Println("System:", msg)
	})

	socket.On("chat message", func(userID, msg string) {
		fmt.Printf("<%s> %s\n", userID, msg)
	})

	go func() {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" {
				continue
			}

			if strings.HasPrefix(line, "/") {
				parts := strings.SplitN(line[1:], " ", 2)
				cmd := parts[0]
				switch cmd {
				case "join":
					if len(parts) > 1 {
						socket.Join(parts[1])
					} else {
						fmt.Println("Usage: /join <room>")
					}
				case "leave":
					if len(parts) > 1 {
						socket.Leave(parts[1])
					} else {
						fmt.Println("Usage: /leave <room>")
					}
				case "quit":
					socket.Close()
					os.Exit(0)
				default:
					fmt.Println("Unknown command:", cmd)
				}
			} else {
				socket.Emit("chat message", line)
			}
		}
	}()

	// Keep the connection alive
	select {}
}
