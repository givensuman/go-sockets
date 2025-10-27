package parser

import (
	"encoding/json"
	"testing"

	sockets "github.com/givensuman/go-sockets"
)

func TestEncode(t *testing.T) {
	// Test CONNECT
	p := sockets.Packet{Type: sockets.Connect}
	data := Encode(p)
	if string(data) != "0" {
		t.Errorf("expected '0', got %s", string(data))
	}

	// Test EVENT
	p = sockets.Packet{Type: sockets.Event, Data: json.RawMessage(`["test", {"key": "value"}]`), Namespace: "/"}
	data = Encode(p)
	if string(data) != `2["test", {"key": "value"}]` {
		t.Errorf("expected '2[\"test\", {\"key\": \"value\"}]', got %s", string(data))
	}

	// Test with namespace
	p = sockets.Packet{Type: sockets.Event, Data: json.RawMessage(`["test"]`), Namespace: "/chat"}
	data = Encode(p)
	if string(data) != "2/chat,[\"test\"]" {
		t.Errorf("expected '2/chat,[\"test\"]', got %s", string(data))
	}

	// Test DISCONNECT
	p = sockets.Packet{Type: sockets.Disconnect}
	data = Encode(p)
	if string(data) != "1" {
		t.Errorf("expected '1', got %s", string(data))
	}

	// Test ERROR
	p = sockets.Packet{Type: sockets.Error, Data: json.RawMessage(`"error message"`), Namespace: "/"}
	data = Encode(p)
	if string(data) != `4"error message"` {
		t.Errorf("expected '4\"error message\"', got %s", string(data))
	}

	// Test BINARY_EVENT
	p = sockets.Packet{Type: sockets.BinaryEvent, Data: json.RawMessage(`["event", "data"]`), Namespace: "/"}
	data = Encode(p)
	if string(data) != `5["event", "data"]` {
		t.Errorf("expected '5[\"event\", \"data\"]', got %s", string(data))
	}

	// Test EVENT with empty data
	p = sockets.Packet{Type: sockets.Event, Data: json.RawMessage(`[]`), Namespace: "/"}
	data = Encode(p)
	if string(data) != "2[]" {
		t.Errorf("expected '2[]', got %s", string(data))
	}

	// Test ACK with ID
	id := uint64(123)
	p = sockets.Packet{Type: sockets.Ack, Data: json.RawMessage(`["response"]`), ID: &id}
	data = Encode(p)
	if string(data) != "3123[\"response\"]" {
		t.Errorf("expected '3123[\"response\"]', got %s", string(data))
	}
}

func TestDecode(t *testing.T) {
	// Test CONNECT
	data := []byte("0")
	p, err := Decode(data)
	if err != nil {
		t.Fatal(err)
	}
	if p.Type != sockets.Connect {
		t.Errorf("expected CONNECT, got %v", p.Type)
	}
	if p.Namespace != "/" {
		t.Errorf("expected namespace '/', got %s", p.Namespace)
	}

	// Test EVENT
	data = []byte(`2["test", {"key": "value"}]`)
	p, err = Decode(data)
	if err != nil {
		t.Fatal(err)
	}
	if p.Type != sockets.Event {
		t.Errorf("expected EVENT, got %v", p.Type)
	}
	if string(p.Data) != `["test", {"key": "value"}]` {
		t.Errorf("expected data, got %s", string(p.Data))
	}
	if p.Namespace != "/" {
		t.Errorf("expected namespace '/', got %s", p.Namespace)
	}

	// Test with namespace
	data = []byte("2/chat,[\"test\"]")
	p, err = Decode(data)
	if err != nil {
		t.Fatal(err)
	}
	if p.Namespace != "/chat" {
		t.Errorf("expected /chat, got %s", p.Namespace)
	}
	if string(p.Data) != `["test"]` {
		t.Errorf("expected data, got %s", string(p.Data))
	}

	// Test DISCONNECT
	data = []byte("1")
	p, err = Decode(data)
	if err != nil {
		t.Fatal(err)
	}
	if p.Type != sockets.Disconnect {
		t.Errorf("expected DISCONNECT, got %v", p.Type)
	}

	// Test ERROR
	data = []byte(`4"error message"`)
	p, err = Decode(data)
	if err != nil {
		t.Fatal(err)
	}
	if p.Type != sockets.Error {
		t.Errorf("expected ERROR, got %v", p.Type)
	}
	if string(p.Data) != `"error message"` {
		t.Errorf("expected data, got %s", string(p.Data))
	}

	// Test BINARY_EVENT
	data = []byte(`5["event", "data"]`)
	p, err = Decode(data)
	if err != nil {
		t.Fatal(err)
	}
	if p.Type != sockets.BinaryEvent {
		t.Errorf("expected BINARY_EVENT, got %v", p.Type)
	}
	if string(p.Data) != `["event", "data"]` {
		t.Errorf("expected data, got %s", string(p.Data))
	}

	// Test EVENT with empty data
	data = []byte("2[]")
	p, err = Decode(data)
	if err != nil {
		t.Fatal(err)
	}
	if p.Type != sockets.Event {
		t.Errorf("expected EVENT, got %v", p.Type)
	}
	if string(p.Data) != "[]" {
		t.Errorf("expected data '[]', got %s", string(p.Data))
	}

	// Test ACK
	data = []byte("3123[\"response\"]")
	p, err = Decode(data)
	if err != nil {
		t.Fatal(err)
	}
	if p.Type != sockets.Ack {
		t.Errorf("expected ACK, got %v", p.Type)
	}
	if p.ID == nil || *p.ID != 123 {
		t.Errorf("expected ID 123, got %v", p.ID)
	}
	if string(p.Data) != `["response"]` {
		t.Errorf("expected data, got %s", string(p.Data))
	}
}
