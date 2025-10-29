# Go-Socket.IO

A Go implementation of the Socket.IO protocol for real-time bidirectional communication between web clients and servers.

It attempts to implement the same [protocol/parser](https://socket.io/docs/v4/socket-io-protocol/), but does not attempt compatibility with the JavaScript library; instead it is meant for use between a Go server and a Go client either through WASM, command line, etc. If you want to write a Go server and client-side JavaScript, use [an alternative](https://github.com/feederco/go-socket.io).

### Install

```bash
go get github.com/givensuman/go-sockets
```

## Quick Start

### Server

```go
package main

import (
    "log"
    "net/http"
    srv "github.com/givensuman/go-sockets/server"
)

func main() {
    server := srv.NewServer()
    ns := server.Of("/")

    ns.On("connection", func(s *srv.Socket) {
        log.Println("Client connected:", s.ID)

        s.On("message", func(msg string) {
            log.Println("Received:", msg)
            s.Emit("response", "Hello from server!")
        })

        s.On("disconnect", func() {
            log.Println("Client disconnected:", s.ID)
        })
    })

    http.ListenAndServe(":3000", server)
}
```

### Client

```go
package main

import (
    "log"
    cli "github.com/givensuman/go-sockets/client"
)

func main() {
    socket, err := cli.Connect("ws://localhost:3000", "/", nil)
    if err != nil {
        log.Fatal(err)
    }
    defer socket.Close()

    socket.On("connect", func() {
        log.Println("Connected to server")
        socket.Emit("message", "Hello from client!")
    })

    socket.On("response", func(msg string) {
        log.Println("Server says:", msg)
    })

    // Keep the connection alive
    select {}
}
```

## Rooms and Broadcasting

### Joining and Leaving Rooms

```go
// Client-side
socket.Join("room1")
socket.Leave("room1")

// Server-side (automatic with client join/leave events)
s.Join("room1")  // Manual
s.Leave("room1")
```

### Broadcasting

```go
// Broadcast to all other clients in namespace
s.Broadcast().Emit("message", "Hello everyone!")

// Broadcast to clients in specific room
s.Broadcast().To("room1").Emit("message", "Hello room1!")

// Broadcast to entire room
ns.To("room1").Emit("message", "Hello room1!")
```

## Acknowledgments

```go
// Client requests data with acknowledgment
socket.Emit("get_data", "request", func(response string) {
    log.Println("Received response:", response)
})

// Server responds to acknowledgment
s.On("get_data", func(data string, ack func(string)) {
    ack("Response to: " + data)
})
```

## Examples

See the `examples/` directory for complete implementations, including a chat application.

## License

MIT
