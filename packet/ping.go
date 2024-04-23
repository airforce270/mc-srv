package packet

import (
	"bytes"
	"fmt"
	"io"

	"github.com/airforce270/mc-srv/packet/id"
	"github.com/airforce270/mc-srv/read"
	"github.com/airforce270/mc-srv/write"
)

// PingRequest request sent from the client to server.
type PingRequest struct {
	Header
	// A number that doesn't actually matter, but we have to respond with it.
	Payload int64
}

func (PingRequest) Name() string { return "PingRequest" }

// ReadPingRequest reads a ping packet from the reader.
// https://wiki.vg/Server_List_Ping#Ping_Request
func ReadPingRequest(r io.Reader, header Header) (PingRequest, error) {
	p := PingRequest{Header: header}

	var err error
	p.Payload, err = read.Long(r)
	if err != nil {
		return p, fmt.Errorf("failed to read payload: %w", err)
	}

	return p, nil
}

// Ping request sent from the client to server.
type PingResponse struct {
	Header
	// A number that doesn't actually matter, but we have to respond with it.
	Payload int64
}

// Write writes the PingResponse to the writer.
// https://wiki.vg/Protocol#Pong_Response
func (pr PingResponse) Write(w io.Writer) error {
	var buf bytes.Buffer
	if err := write.Long(&buf, pr.Payload); err != nil {
		return fmt.Errorf("failed to write ping response payload: %w", err)
	}

	if err := writePacket(w, id.HandshakePong, &buf); err != nil {
		return fmt.Errorf("failed to write packet: %w", err)
	}

	return nil
}
