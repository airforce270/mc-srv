package packet

import (
	"fmt"
	"io"
	"log"

	"github.com/airforce270/mc-srv/flags"
	"github.com/airforce270/mc-srv/read"
	"github.com/airforce270/mc-srv/write"
)

// Common fields that every packet has.
type Header struct {
	// Length is the length of the PacketID and following data in the packet.
	Length int32
	// PacketID is the ID of the packet.
	PacketID ID
}

// ID returns the packet ID of the header.
func (h Header) ID() ID { return h.PacketID }

func (h Header) Write(w io.Writer) error {
	if err := write.VarInt(w, h.Length); err != nil {
		return fmt.Errorf("failed to write header length (%d): %w", h.Length, err)
	}
	if err := write.VarInt(w, int32(h.PacketID)); err != nil {
		return fmt.Errorf("failed to write header packet ID (%d): %w", h.PacketID, err)
	}
	return nil
}

func readHeader(r io.Reader, logger *log.Logger) (Header, error) {
	var h Header
	var err error

	h.Length, err = read.VarInt(r)
	if err != nil {
		return h, fmt.Errorf("failed to read packet length: %w", err)
	}
	if h.Length == 0 {
		return h, nil
	}

	packetID, err := read.VarInt(r)
	if err != nil {
		return h, fmt.Errorf("failed to read packet id: %w", err)
	}
	h.PacketID = ID(packetID)

	if *flags.Verbose {
		logger.Printf("Received header: %+v", h)
	}

	return h, nil
}
