// Package packet contains packet definitions and parsers.
package packet

import (
	"bytes"
	"fmt"
	"io"
	"log"

	"github.com/airforce270/mc-srv/flags"
	"github.com/airforce270/mc-srv/packet/id"
	"github.com/airforce270/mc-srv/server/serverstate"
	"github.com/airforce270/mc-srv/write"
)

// Packet is a packet in the Minecraft client-server protocol.
type Packet interface {
	// ID returns the ID of the packet.
	ID() id.ID
	// Name returns the human-readable display name of the packet.
	Name() string
}

// Read reads the next packet from the reader.
func Read(r io.Reader, state serverstate.State, logger *log.Logger) (Packet, error) {
	h, err := readHeader(r)
	if err != nil {
		return nil, fmt.Errorf("failed to read header: %w", err)
	}
	if h.Length == 0 {
		return nil, nil
	}
	if *flags.Verbose {
		logger.Printf("Received header {id=0x%x length=%d}", h.PacketID, h.Length)
	}

	packetIDLen := write.VarIntLen(int32(h.PacketID))

	fieldsLength := int(h.Length) - packetIDLen
	if fieldsLength == 0 && h.PacketID == id.StatusRequest {
		return StatusRequest{Header: h}, nil
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
	switch state {
	case serverstate.PreHandshake, serverstate.ClientRequestingStatus:
		switch h.PacketID {
		case id.Handshake:
			p, err = ReadHandshake(&buf, h)
		case id.HandshakePing:
			p, err = ReadPingRequest(&buf, h)
		default:
			logger.Printf("Unhandled packet type (state=%v): %x", state, h.PacketID)
			return nil, nil
		}
	case serverstate.ClientRequestingLogin:
		switch h.PacketID {
		case id.LoginStart:
			p, err = ReadLoginStart(&buf, h)
		}
	case serverstate.EncryptionRequested:
		switch h.PacketID {
		case id.EncryptionResponse:
			p, err = ReadEncryptionResponse(&buf, h)
		}
	case serverstate.LoginCompletePendingAcknowledgement:
		switch h.PacketID {
		case id.LoginAcknowledgement:
			p = LoginAcknowledgement{Header: h}
		}
	case serverstate.LoginComplete:
		switch h.PacketID {
		case id.ClientInformation:
			p, err = ReadConfigClientInformation(&buf, h)
		case id.ServerboundPlugin:
			p, err = ReadConfigServerboundPlugin(&buf, h)
		case id.AcknowledgeFinish:
			p = AcknowledgeFinishConfiguration{Header: h}
		case id.ServerboundKeepAlive:
			p, err = ReadServerboundKeepAlive(&buf, h)
		case id.Pong:
			p, err = ReadConfigPong(&buf, h)
		case id.ResourcePackResponse:
			p, err = ReadConfigResourcePackResponse(&buf, h)
		}
	default:
		logger.Printf("Unhandled packet type (state=%v): %x", state, h.PacketID)
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
func writePacket(w io.Writer, id id.ID, payload readLengther) error {
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
