// Package client provides functionality to connect to a WebSocket server.
package client

import (
	"net/url"

	"github.com/givensuman/go-sockets"
	"github.com/givensuman/go-sockets/internal/emitter"
	"github.com/gorilla/websocket"
)

// Connect establishes a WebSocket connection to the server at the given URL.
func Connect(serverURL string, onConnect func(*Socket)) (*Socket, error) {
	u, err := url.Parse(serverURL)
	if err != nil {
		return nil, err
	}

	dialer := websocket.Dialer{}
	conn, _, err := dialer.Dial(u.String(), nil)
	if err != nil {
		return nil, err
	}

	socket := &Socket{
		EventEmitter: emitter.EventEmitter{},
		Conn:         conn,
		writeChan:    make(chan sockets.Packet, 10),
	}

	go socket.readLoop()
	go socket.writeLoop()

	if onConnect != nil {
		onConnect(socket)
	}

	socket.EventEmitter.Emit("connect")

	return socket, nil
}
