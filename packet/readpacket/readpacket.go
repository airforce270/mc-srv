// Package readpacket handles reading packets.
package readpacket

import (
	"bytes"
	"fmt"
	"io"
	"log"

	"github.com/airforce270/mc-srv/flags"
	"github.com/airforce270/mc-srv/packet"
	"github.com/airforce270/mc-srv/packet/config"
	"github.com/airforce270/mc-srv/packet/id"
	"github.com/airforce270/mc-srv/packet/login"
	"github.com/airforce270/mc-srv/packet/slp"
	"github.com/airforce270/mc-srv/server/serverstate"
	"github.com/airforce270/mc-srv/write"
)

// Read reads the next packet from the reader.
func Read(r io.Reader, state serverstate.State, logger *log.Logger) (packet.Packet, error) {
	h, err := packet.ReadHeader(r)
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
		return slp.StatusRequest{Header: h}, nil
	}

	var buf bytes.Buffer
	readN, err := io.CopyN(&buf, r, int64(fieldsLength))
	if err != nil {
		return nil, fmt.Errorf("failed to read packet bytes: %w", err)
	}
	if readN != int64(fieldsLength) {
		return nil, fmt.Errorf("expected to read %d bytes, only read %d", fieldsLength, readN)
	}

	var p packet.Packet
	switch state {
	case serverstate.PreHandshake, serverstate.ClientRequestingStatus:
		switch h.PacketID {
		case id.Handshake:
			p, err = slp.ReadHandshake(&buf, h)
		case id.HandshakePing:
			p, err = slp.ReadHandshakePingRequest(&buf, h)
		default:
			logger.Printf("Unhandled packet type (state=%v): %x", state, h.PacketID)
			return nil, nil
		}
	case serverstate.ClientRequestingLogin:
		switch h.PacketID {
		case id.LoginStart:
			p, err = login.ReadLoginStart(&buf, h)
		}
	case serverstate.EncryptionRequested:
		switch h.PacketID {
		case id.EncryptionResponse:
			p, err = login.ReadEncryptionResponse(&buf, h)
		}
	case serverstate.LoginCompletePendingAcknowledgement:
		switch h.PacketID {
		case id.LoginAcknowledgement:
			p = login.LoginAcknowledgement{Header: h}
		}
	case serverstate.LoginComplete:
		switch h.PacketID {
		case id.ClientInformation:
			p, err = config.ReadConfigClientInformation(&buf, h)
		case id.ServerboundPlugin:
			p, err = config.ReadConfigServerboundPlugin(&buf, h)
		case id.AcknowledgeFinish:
			p = config.AcknowledgeFinishConfiguration{Header: h}
		case id.ServerboundKeepAlive:
			p, err = config.ReadServerboundKeepAlive(&buf, h)
		case id.Pong:
			p, err = config.ReadConfigPong(&buf, h)
		case id.ResourcePackResponse:
			p, err = config.ReadConfigResourcePackResponse(&buf, h)
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
