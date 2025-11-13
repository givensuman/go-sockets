# go-sockets

A Go implementation of the Socket.IO API for real-time bidirectional communication between Go clients and Go servers.

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
// client.go
socket.Join("some room")
socket.Leave("some room")
```

### Broadcasting

```go
// server.go
// Broadcast to all other clients in namespace
s.Broadcast().Emit("message", "Hello everyone!")

// Broadcast to clients in specific room
s.Broadcast().To("some room").Emit("message", "Hello room!")

// Broadcast to entire room, including sender
s.To("some room").Emit("message", "Hello room!")
```

## Acknowledging

```go
// client.go
// Client requests data with acknowledgment
socket.Emit("get_data", "request", func(response string) {
    log.Println("Received response:", response)
})

// server.go
// Server responds to acknowledgment
s.On("get_data", func(data string, ack func(string)) {
    ack("Response to: " + data)
})
```

## Examples

See the [examples/](./examples) directory for complete implementations, including a chat application.

## License

[MIT](./LICENSE)
