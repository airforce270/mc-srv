// Package handshaketest contains testdata for the handshake packet.
package handshaketest

import "github.com/airforce270/mc-srv/packet/headertest"

var (
	Notchian = []byte{
		0xfd, 0x05, 0x09, // protocol version
		0x31, 0x32, 0x37, 0x2e, 0x30, 0x2e, 0x30, 0x2e, 0x31, // address
		0x63, 0xdd, // port
		0x01, // next state
	}
	NotchianHeader = headertest.HandshakeLen16
)
