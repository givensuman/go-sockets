// Package client provides functionality to connect to a Socket.IO WebSocket server.
// It allows clients to emit events, listen for events, and manage acknowledgments.
package client

import (
	"net/url"

	"github.com/givensuman/go-sockets"
	"github.com/givensuman/go-sockets/internal/emitter"
	"github.com/gorilla/websocket"
)

// Connect establishes a WebSocket connection to the Socket.IO server at the given URL and namespace.
// It calls onConnect with the socket once connected, then emits a "connect" event.
// Namespace defaults to "/" if empty.
func Connect(serverURL string, namespace string, onConnect func(*Socket)) (*Socket, error) {
	if namespace == "" {
		namespace = "/"
	}

	u, err := url.Parse(serverURL)
	if err != nil {
		return nil, err
	}
	u.Path = namespace

	dialer := websocket.Dialer{}
	conn, _, err := dialer.Dial(u.String(), nil)
	if err != nil {
		return nil, err
	}

	socket := &Socket{
		EventEmitter: emitter.EventEmitter{},
		Conn:         conn,
		writeChan:    make(chan sockets.Packet, 10),
		Namespace:    namespace,
	}

	go socket.readLoop()
	go socket.writeLoop()

	if onConnect != nil {
		onConnect(socket)
	}

	socket.EventEmitter.Emit("connect")

	return socket, nil
}
