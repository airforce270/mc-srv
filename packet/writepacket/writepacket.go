// Package writepacket handles writing packets.
package writepacket

import (
	"fmt"
	"io"

	"github.com/airforce270/mc-srv/packet"
	"github.com/airforce270/mc-srv/packet/id"
)

// Write writes a packet to the writer.
func Write(w io.Writer, id id.ID, payload readLengther) error {
	payloadLen := payload.Len()
	h := packet.Header{
		Length:   int32(id.Len() + payloadLen),
		PacketID: id,
	}
	if err := h.Write(w); err != nil {
		return fmt.Errorf("failed to write packet header (%+v): %w", h, err)
	}

	wroteLen, err := io.Copy(w, payload)
	if err != nil {
		return fmt.Errorf("failed to write packet payload: %w", err)
	}
	if wroteLen != int64(payloadLen) {
		return fmt.Errorf("writing packet payload expected to write %d bytes, but wrote %d", payloadLen, wroteLen)
	}

	return nil
}

// readLengther is any io.Reader with a Len() method.
type readLengther interface {
	io.Reader
	// Len returns the number of bytes available to be read.
	Len() int
}
