package packet

import (
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
	length = write.VarIntLen(int32(i))
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

	LoginStartID         ID = 0x00
	EncryptionResponseID ID = 0x01
)

// Response (Server->Client) packet IDs.
const (
	StatusResponseID ID = 0x00
	PingResponseID   ID = 0x01

	EncryptionRequestID ID = 0x01
	LoginSuccessID      ID = 0x02
)

var (
	idLengthCache    = map[ID]int{}
	idLengthCacheMtx sync.Mutex
)
