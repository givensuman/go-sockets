# go-sockets

This is a WebSocket framework based off of the [socket.io](https://socket.io) library for the Node.js ecosystem.

It attempts to implement the same [protocol/parser](https://socket.io/docs/v4/socket-io-protocol/), but does not attempt compatibility with the JavaScript library; instead it is meant for use between a Go server and a Go client either through WASM, command line, etc. If you want to write a Go server and client-side JavaScript, use [an alternative](https://github.com/feederco/go-socket.io).

### Install

```bash
go get github.com/givensuman/go-sockets
```

### Import

```go
// Server
import (
  "github.com/givensuman/go-sockets/server"
)
```

```go
// Client
import (
  "github.com/givensuman/go-sockets/client"
)
```

### Differences

|go-sockets|socket.io|
|---|---|
|More performant<sup>[citation needed]</sup>|It's javaScript|
|Concurrent by nature|Can hack [concurrency](https://socket.io/docs/v4/cluster-adapter/)|
|Work in progress|Feature-rich|
### License

[MIT](./LICENSE)
