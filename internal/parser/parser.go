// Package parser provides functions to encode and decode sockets.Packet structures.
package parser

import (
	"encoding/json"
	"errors"
	"strconv"
	"strings"

	"github.com/givensuman/go-sockets"
)

// Encode converts a gosock.Packet into its byte slice representation.
func Encode(p sockets.Packet) []byte {
	var sb strings.Builder
	sb.WriteString(strconv.Itoa(int(p.Type)))
	if p.Namespace != "" && p.Namespace != "/" {
		sb.WriteString(p.Namespace)
		sb.WriteString(",")
	}
	if p.ID != nil {
		sb.WriteString(strconv.FormatUint(*p.ID, 10))
	}
	if len(p.Data) > 0 {
		sb.Write(p.Data)
	}
	return []byte(sb.String())
}

// Decode converts a byte slice representation into a gosock.Packet.
func Decode(data []byte) (sockets.Packet, error) {
	s := string(data)
	if len(s) == 0 {
		return sockets.Packet{}, errors.New("empty data")
	}

	typeDigit := s[0]
	if typeDigit < '0' || typeDigit > '9' {
		return sockets.Packet{}, errors.New("invalid packet type")
	}
	p := sockets.Packet{
		Type:      sockets.PacketType(typeDigit - '0'),
		Namespace: "/",
	}

	s = s[1:]

	// Check for namespace
	if len(s) > 0 && s[0] == '/' {
		commaIndex := strings.Index(s, ",")
		if commaIndex != -1 {
			p.Namespace = s[:commaIndex]
			s = s[commaIndex+1:]
		} else {
			p.Namespace = s
			s = ""
		}
	}

	// For ACK, check if starts with digit for ID
	if p.Type == sockets.Ack && len(s) > 0 && s[0] >= '0' && s[0] <= '9' {
		idStr := ""
		i := 0
		for i < len(s) && s[i] >= '0' && s[i] <= '9' {
			idStr += string(s[i])
			i++
		}
		id, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil {
			return sockets.Packet{}, err
		}
		p.ID = &id
		s = s[i:]
	}
	p.Data = json.RawMessage(s)
	return p, nil
}
