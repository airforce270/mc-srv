package id

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
	// Handshake
	Handshake     ID = 0x00
	StatusRequest ID = 0x00
	HandshakePing ID = 0x01

	// Login
	LoginStart         ID = 0x00
	EncryptionResponse ID = 0x01

	// Configuration
	ClientInformation    ID = 0x00
	ServerboundPlugin    ID = 0x01
	AcknowledgeFinish    ID = 0x02
	ServerboundKeepAlive ID = 0x03
	Pong                 ID = 0x04
	ResourcePackResponse ID = 0x05

	// Play
)

// Response (Server->Client) packet IDs.
const (
	// Handshake
	StatusResponse ID = 0x00
	HandshakePong  ID = 0x01

	// Login
	EncryptionRequest    ID = 0x01
	LoginSuccess         ID = 0x02
	LoginAcknowledgement ID = 0x03

	// Configuration
	ClientboundPlugin        ID = 0x00
	ConfigDisconnect         ID = 0x01
	FinishConfiguration      ID = 0x02
	ClientboundKeepAlive     ID = 0x03
	Ping                     ID = 0x04
	RegistryData             ID = 0x05
	ConfigRemoveResourcePack ID = 0x06
	ConfigAddResourcePack    ID = 0x06
	FeatureFlags             ID = 0x08
	ConfigUpdateTags         ID = 0x09

	// Play
)

var (
	idLengthCache    = map[ID]int{}
	idLengthCacheMtx sync.Mutex
)
