// Package slp contains packets for the Server List Ping protocol.
// https://wiki.vg/Server_List_Ping
package slp

import (
	"bytes"
	"fmt"
	"io"

	"github.com/airforce270/mc-srv/packet"
	"github.com/airforce270/mc-srv/packet/id"
	"github.com/airforce270/mc-srv/packet/writepacket"
	"github.com/airforce270/mc-srv/read"
	"github.com/airforce270/mc-srv/write"
)

const (
	HandshakeNextStateStatus = 1
	HandshakeNextStateLogin  = 2
)

// Initial packet sent from the client server to establish connection.
type Handshake struct {
	packet.Header
	// The version that the client plans on using to connect to the server
	// (which is not important for the ping).
	// If the client is pinging to determine what version to use,
	// by convention -1 should be set.
	ProtocolVersion int32
	// Hostname or IP, e.g. localhost or 127.0.0.1, that was used to connect.
	ServerAddress string
	// Default is 25565. The Notchian server does not use this information.
	ServerPort uint16
	// Should be 1 for status, but could also be 2 for login.
	NextState int32
}

func (Handshake) Name() string { return "Handshake" }

// ReadHandshake reads a handshake packet from the reader.
// https://wiki.vg/Server_List_Ping#Handshake
func ReadHandshake(r io.Reader, header packet.Header) (Handshake, error) {
	h := Handshake{Header: header}

	var err error

	h.ProtocolVersion, err = read.VarInt(r)
	if err != nil {
		return h, fmt.Errorf("failed to read protocol version: %w", err)
	}

	h.ServerAddress, err = read.String(r)
	if err != nil {
		return h, fmt.Errorf("failed to read server address: %w", err)
	}

	h.ServerPort, err = read.UnsignedShort(r)
	if err != nil {
		return h, fmt.Errorf("failed to read server port: %w", err)
	}

	h.NextState, err = read.VarInt(r)
	if err != nil {
		return h, fmt.Errorf("failed to read next state: %w", err)
	}

	return h, nil
}

// HandshakePingRequest is a request sent from the client to server.
type HandshakePingRequest struct {
	packet.Header
	// A number that doesn't actually matter, but we have to respond with it.
	Payload int64
}

func (HandshakePingRequest) Name() string { return "PingRequest(Handshake)" }

// ReadHandshakePingRequest reads a ping packet from the reader.
// https://wiki.vg/Server_List_Ping#Ping_Request
func ReadHandshakePingRequest(r io.Reader, header packet.Header) (HandshakePingRequest, error) {
	p := HandshakePingRequest{Header: header}

	var err error
	p.Payload, err = read.Long(r)
	if err != nil {
		return p, fmt.Errorf("failed to read payload: %w", err)
	}

	return p, nil
}

// Ping request sent from the client to server.
type HandshakePingResponse struct {
	packet.Header
	// A number that doesn't actually matter, but we have to respond with it.
	Payload int64
}

func (HandshakePingResponse) Name() string { return "PingResponse(Handshake)" }

// Write writes the HandshakePingResponse to the writer.
// https://wiki.vg/Protocol#Pong_Response
func (pr HandshakePingResponse) Write(w io.Writer) error {
	var buf bytes.Buffer
	if err := write.Long(&buf, pr.Payload); err != nil {
		return fmt.Errorf("failed to write ping response payload: %w", err)
	}

	if err := writepacket.Write(w, id.HandshakePong, &buf); err != nil {
		return fmt.Errorf("failed to write packet: %w", err)
	}

	return nil
}
