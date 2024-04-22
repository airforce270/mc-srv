// Package pingtest provides test data for the ping packet.
package pingtest

import "github.com/airforce270/mc-srv/packet/headertest"

var (
	Notchian       = []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x68, 0x53, 0xa8} // 6837160
	NotchianHeader = headertest.PingLen9
)
