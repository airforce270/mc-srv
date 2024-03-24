// Package packet contains packet definitions and parsers.
package packet

import (
	"bytes"
	"fmt"
	"io"
	"log"

	"github.com/airforce270/mc-srv/write"
)

// Packet is a packet in the Minecraft client-server protocol.
type Packet interface {
	// ID returns the ID of the packet.
	ID() ID
	// Name returns the human-readable display name of the packet.
	Name() string
}

// Read reads the next packet from the reader.
func Read(r io.Reader) (Packet, error) {
	h, err := readHeader(r)
	if err != nil {
		return nil, fmt.Errorf("failed to read header: %w", err)
	}
	if h.Length == 0 {
		return nil, nil
	}

	packetIDLen, err := write.VarIntLen(int32(h.PacketID))
	if err != nil {
		return nil, fmt.Errorf("failed to calculate serialized packet id len (%d): %w", h.PacketID, err)
	}

	fieldsLength := int(h.Length) - packetIDLen
	if fieldsLength == 0 {
		if h.PacketID == StatusRequestID {
			return StatusRequest{Header: h}, nil
		}
		return nil, nil
	}

	var buf bytes.Buffer
	readN, err := io.CopyN(&buf, r, int64(fieldsLength))
	if err != nil {
		return nil, fmt.Errorf("failed to read packet bytes: %w", err)
	}
	if readN != int64(fieldsLength) {
		return nil, fmt.Errorf("expected to read %d bytes, only read %d", fieldsLength, readN)
	}

	var p Packet
	switch h.PacketID {
	case HandshakeID:
		p, err = ReadHandshake(&buf, h)
	case PingRequestID:
		p, err = ReadPing(&buf, h)
	default:
		log.Printf("Unhandled packet type: %x", h.PacketID)
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to read packet (header=%+v): %w", h, err)
	}

	return p, nil
}

// readLengther is any io.Reader with a Len() method.
type readLengther interface {
	io.Reader
	// Len returns the number of bytes available to be read.
	Len() int
}

// writePacket writes a packet to the writer.
func writePacket(w io.Writer, id ID, payload readLengther) error {
	payloadLen := payload.Len()
	h := Header{
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
