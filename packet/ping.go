package packet

import (
	"bytes"
	"fmt"
	"io"

	"github.com/airforce270/mc-srv/read"
	"github.com/airforce270/mc-srv/write"
)

// Ping request sent from the client to server.
type Ping struct {
	Header
	// A number that doesn't actually matter, but we have to respond with it.
	Payload int64
}

func (p Ping) Name() string { return "PingRequest" }

// ReadPing reads a ping packet from the reader.
// https://wiki.vg/Server_List_Ping#Ping_Request
func ReadPing(r io.Reader, header Header) (Ping, error) {
	p := Ping{Header: header}

	var err error
	p.Payload, err = read.Long(r)
	if err != nil {
		return p, fmt.Errorf("failed to read payload: %w", err)
	}

	return p, nil
}

// WritePingResponse writes a PingResponse to the writer.
// https://wiki.vg/Protocol#Pong_Response
func WritePingResponse(w io.Writer, payload int64) error {
	var buf bytes.Buffer
	if err := write.Long(&buf, payload); err != nil {
		return fmt.Errorf("failed to write ping response payload: %w", err)
	}

	if err := writePacket(w, PingResponseID, &buf); err != nil {
		return fmt.Errorf("failed to write packet: %w", err)
	}

	return nil
}
