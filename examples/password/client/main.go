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
	fmt.Print("Are you an administrator? (y/n): ")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	isAdmin := strings.ToLower(strings.TrimSpace(scanner.Text())) == "y"

	namespace := "/"
	if isAdmin {
		namespace = "/admin"
	}

	socket, err := sockets.Connect("ws://localhost:3000", namespace, nil)
	if err != nil {
		log.Fatal("Failed to connect:", err)
	}
	defer socket.Close()

	authenticated := false

	socket.On("connect", func() {
		if isAdmin {
			fmt.Println("Connected to admin namespace")
			fmt.Print("Enter password: ")
		} else {
			fmt.Println("Connected as regular user")
			fmt.Println("Commands:")
			fmt.Println("  /join <room>  - Join a room")
			fmt.Println("  /leave <room> - Leave a room")
			fmt.Println("  <message>     - Send chat message")
			fmt.Println("  /quit         - Exit")
		}
	})

	if isAdmin {
		socket.On("authenticated", func(msg string) {
			authenticated = true
			fmt.Println(msg)
			fmt.Println("Commands:")
			fmt.Println("  /broadcast <message> - Broadcast admin message")
			fmt.Println("  /quit                - Exit")
		})

		socket.On("auth_failed", func(msg string) {
			fmt.Println("Authentication failed:", msg)
			socket.Close()
			os.Exit(1)
		})

		socket.On("admin_message", func(userID, msg string) {
			fmt.Printf("[Admin Broadcast from %s] %s\n", userID, msg)
		})

		socket.On("error", func(msg string) {
			fmt.Println("Error:", msg)
		})
	} else {
		socket.On("message", func(msg string) {
			fmt.Println("System:", msg)
		})

		socket.On("chat message", func(userID, msg string) {
			fmt.Printf("<%s> %s\n", userID, msg)
		})
	}

	go func() {
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" {
				continue
			}

			if isAdmin && !authenticated {
				// First input is password
				socket.Emit("auth", line)
				continue
			}

			if strings.HasPrefix(line, "/") {
				parts := strings.SplitN(line[1:], " ", 2)
				cmd := parts[0]
				switch cmd {
				case "join":
					if !isAdmin {
						if len(parts) > 1 {
							socket.Join(parts[1])
						} else {
							fmt.Println("Usage: /join <room>")
						}
					} else {
						fmt.Println("Unknown command:", cmd)
					}
				case "leave":
					if !isAdmin {
						if len(parts) > 1 {
							socket.Leave(parts[1])
						} else {
							fmt.Println("Usage: /leave <room>")
						}
					} else {
						fmt.Println("Unknown command:", cmd)
					}
				case "broadcast":
					if isAdmin && authenticated {
						if len(parts) > 1 {
							socket.Emit("admin_broadcast", parts[1])
						} else {
							fmt.Println("Usage: /broadcast <message>")
						}
					} else {
						fmt.Println("Unknown command:", cmd)
					}
				case "quit":
					socket.Close()
					os.Exit(0)
				default:
					fmt.Println("Unknown command:", cmd)
				}
			} else {
				if isAdmin {
					fmt.Println("Use /broadcast <message> to send admin messages")
				} else {
					socket.Emit("chat message", line)
				}
			}
		}
	}()

	// Keep the connection alive
	select {}
}
