// Package packet contains packet definitions and parsers.
package packet

import (
	"github.com/airforce270/mc-srv/packet/id"
)

// Packet is a packet in the Minecraft client-server protocol.
type Packet interface {
	// ID returns the ID of the packet.
	ID() id.ID
	// Name returns the human-readable display name of the packet.
	Name() string
}
