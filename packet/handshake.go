package packet

import (
	"fmt"
	"io"

	"github.com/airforce270/mc-srv/read"
)

const (
	HandshakeNextStateStatus = 1
	HandshakeNextStateLogin  = 2
)

// Initial packet sent from the client server to establish connection.
type Handshake struct {
	Header
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
func ReadHandshake(r io.Reader, header Header) (Handshake, error) {
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
