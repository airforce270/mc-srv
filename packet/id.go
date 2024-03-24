package packet

import (
	"fmt"
	"sync"

	"github.com/airforce270/mc-srv/write"
)

// ID is a packet ID.
type ID int32

// Len returns the length of the packet ID, in serialized bytes.
func (i ID) Len() int {
	length, ok := idLengthCache[i]
	if ok {
		return length
	}
	length = mustCalcIDLength(i)
	idLengthCacheMtx.Lock()
	idLengthCache[i] = length
	idLengthCacheMtx.Unlock()
	return length
}

// Request (Client->Server) packet IDs.
const (
	HandshakeID     ID = 0x00
	StatusRequestID ID = 0x00
	PingRequestID   ID = 0x01
)

// Response (Server->Client) packet IDs.
const (
	StatusResponseID ID = 0x00
	PingResponseID   ID = 0x01
)

var (
	idLengthCache    = map[ID]int{}
	idLengthCacheMtx sync.Mutex
)

func mustCalcIDLength(id ID) int {
	length, err := write.VarIntLen(int32(id))
	if err != nil {
		panic(fmt.Sprintf("Failed to calculate id len (%d): %v", id, err))
	}
	return length
}
