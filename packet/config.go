package packet

import (
	"fmt"
	"io"

	"github.com/airforce270/mc-srv/read"
	"github.com/google/uuid"
)

// Player's chat mode, for ConfigClientInformation.
type ChatMode uint8

const (
	ChatModeEnabled      ChatMode = 0
	ChatModeCommandsOnly ChatMode = 1
	ChatModeHidden       ChatMode = 2
)

// Player's main hand, for ConfigClientInformation.
type MainHand uint8

const (
	MainHandLeft  MainHand = 0
	MainHandRight MainHand = 1
)

// Packet sent from the client to convey its config.
type ConfigClientInformation struct {
	Header
	// Player's locale, e.g. "en_GB".
	Locale string
	// Player's (client-side) view distance, in chunks.
	ViewDistance byte
	// Player's chat mode.
	// See https://wiki.vg/Chat#Client_chat_mode for more info.
	ChatMode ChatMode
	// "Colors" multiplayer setting. Whether the chat can be colored.
	ChatColorsEnabled bool
	// Displayed skin parts.
	//
	// Bit 0 (0x01): Cape enabled
	// Bit 1 (0x02): Jacket enabled
	// Bit 2 (0x04): Left Sleeve enabled
	// Bit 3 (0x08): Right Sleeve enabled
	// Bit 4 (0x10): Left Pants Leg enabled
	// Bit 5 (0x20): Right Pants Leg enabled
	// Bit 6 (0x40): Hat enabled
	// The most significant bit (bit 7, 0x80) appears to be unused.
	DisplayedSkinParts byte
	// Player's main hand.
	MainHand MainHand
	// Enables filtering of text on signs and written book titles.
	// Currently always false (i.e. the filtering is disabled)
	EnableTextFiltering bool
	// Servers usually list online players,
	// this option allows the client to not appear in that list.
	AllowServerListings bool
}

func (c ConfigClientInformation) CapeEnabled() bool          { return c.DisplayedSkinParts&0x01 == 1 }
func (c ConfigClientInformation) JacketEnabled() bool        { return c.DisplayedSkinParts&0x02 == 1 }
func (c ConfigClientInformation) LeftSleeveEnabled() bool    { return c.DisplayedSkinParts&0x04 == 1 }
func (c ConfigClientInformation) RightSleeveEnabled() bool   { return c.DisplayedSkinParts&0x08 == 1 }
func (c ConfigClientInformation) LeftPantsLegEnabled() bool  { return c.DisplayedSkinParts&0x10 == 1 }
func (c ConfigClientInformation) RightPantsLegEnabled() bool { return c.DisplayedSkinParts&0x20 == 1 }
func (c ConfigClientInformation) HatEnabled() bool           { return c.DisplayedSkinParts&0x40 == 1 }

func (ConfigClientInformation) Name() string { return "ClientInformation(config)" }

// ReadConfigClientInformation reads a Client Information (config) packet
// from the reader.
// https://wiki.vg/Protocol#Client_Information_.28configuration.29
func ReadConfigClientInformation(r io.Reader, header Header) (ConfigClientInformation, error) {
	p := ConfigClientInformation{Header: header}

	var err error

	p.Locale, err = read.String(r)
	if err != nil {
		return p, fmt.Errorf("failed to read locale: %w", err)
	}

	p.ViewDistance, err = read.Byte(r)
	if err != nil {
		return p, fmt.Errorf("failed to read view distance: %w", err)
	}

	chatMode, err := read.VarInt(r)
	if err != nil {
		return p, fmt.Errorf("failed to read chat mode: %w", err)
	}
	p.ChatMode = ChatMode(chatMode)

	p.ChatColorsEnabled, err = read.Bool(r)
	if err != nil {
		return p, fmt.Errorf("failed to read chat colors enabled: %w", err)
	}

	p.DisplayedSkinParts, err = read.Byte(r)
	if err != nil {
		return p, fmt.Errorf("failed to read displayed skin parts: %w", err)
	}

	mainHand, err := read.VarInt(r)
	if err != nil {
		return p, fmt.Errorf("failed to read main hand: %w", err)
	}
	p.MainHand = MainHand(mainHand)

	p.EnableTextFiltering, err = read.Bool(r)
	if err != nil {
		return p, fmt.Errorf("failed to read enable text filtering: %w", err)
	}

	p.AllowServerListings, err = read.Bool(r)
	if err != nil {
		return p, fmt.Errorf("failed to read allow server listings: %w", err)
	}

	return p, nil
}

// Packet for mods and plugins to send data client->server.
type ConfigServerboundPlugin struct {
	Header
	// Any data, depending on the channel.
	// `minecraft:` channels are documented here: https://wiki.vg/Plugin_channel
	Data []byte
}

func (ConfigServerboundPlugin) Name() string { return "ServerboundPlugin(config)" }

// ReadConfigServerboundPlugin a Serverbound Plugin (config) packet
// from the reader.
// https://wiki.vg/Protocol#Serverbound_Plugin_.28configuration.29
func ReadConfigServerboundPlugin(r io.Reader, header Header) (ConfigServerboundPlugin, error) {
	p := ConfigServerboundPlugin{Header: header}

	var err error

	p.Data, err = io.ReadAll(r)
	if err != nil {
		return p, fmt.Errorf("failed to read data: %w", err)
	}

	// TODO: convert/decode the data

	return p, nil
}

// Packet sent by the client to notify the server
// the configuration process has finished.
// Sent in response to FinishConfiguration.
type AcknowledgeFinishConfiguration struct {
	Header
}

func (AcknowledgeFinishConfiguration) Name() string { return "AcknowledgeFinishConfiguration" }

// Client->server response to the server->client keep alive packets.
type ServerboundKeepAlive struct {
	Header
	// Should be the same number that the server sent in its keep alive packet.
	KeepAliveID int64
}

func (ServerboundKeepAlive) Name() string { return "ServerboundKeepAlive" }

// ReadServerboundKeepAlive reads a Serverbound Keep Alive packet
// from the reader.
// https://wiki.vg/Protocol#Serverbound_Keep_Alive_.28configuration.29
func ReadServerboundKeepAlive(r io.Reader, header Header) (ServerboundKeepAlive, error) {
	p := ServerboundKeepAlive{Header: header}

	var err error

	p.KeepAliveID, err = read.Long(r)
	if err != nil {
		return p, fmt.Errorf("failed to read ID: %w", err)
	}

	return p, nil
}

// Client->server response to the server->client ping packets.
type ConfigPong struct {
	Header
	// Should be the same number that the server sent in its ping packet.
	PingID int32
}

func (ConfigPong) Name() string { return "Pong(config)" }

// ReadConfigPong reads a Pong (config) packet from the reader.
// https://wiki.vg/Protocol#Pong_.28configuration.29
func ReadConfigPong(r io.Reader, header Header) (ConfigPong, error) {
	p := ConfigPong{Header: header}

	var err error

	p.PingID, err = read.Int(r)
	if err != nil {
		return p, fmt.Errorf("failed to read ID: %w", err)
	}

	return p, nil
}

// Resource pack result ID, for ConfigResourcePackResponse.
type ResourcePackResult uint8

const (
	ResourcePackResultSuccessfullyDownloaded ResourcePackResult = 0
	ResourcePackResultDeclined               ResourcePackResult = 1
	ResourcePackResultFailedToDownload       ResourcePackResult = 2
	ResourcePackResultAccepted               ResourcePackResult = 3
	ResourcePackResultDownloaded             ResourcePackResult = 4
	ResourcePackResultInvalidURL             ResourcePackResult = 5
	ResourcePackResultFailedToReload         ResourcePackResult = 6
	ResourcePackResultDiscarded              ResourcePackResult = 7
)

// Resource pack information client->server (?)
type ConfigResourcePackResponse struct {
	Header
	// The unique identifier of the resource pack
	// received in ConfigAddResourcePack.
	ResourcePackUUID uuid.UUID
	// The result ID.
	Result ResourcePackResult
}

func (ConfigResourcePackResponse) Name() string { return "ResourcePackResponse(config)" }

// ReadConfigResourcePackResponse a Resource Pack Response (config) packet
// from the reader.
// https://wiki.vg/Protocol#Resource_Pack_Response_.28configuration.29
func ReadConfigResourcePackResponse(r io.Reader, header Header) (ConfigResourcePackResponse, error) {
	p := ConfigResourcePackResponse{Header: header}

	var err error

	p.ResourcePackUUID, err = read.UUID(r)
	if err != nil {
		return p, fmt.Errorf("failed to read resource pack uuid: %w", err)
	}

	result, err := read.VarInt(r)
	if err != nil {
		return p, fmt.Errorf("failed to read resource pack uuid: %w", err)
	}
	p.Result = ResourcePackResult(result)

	return p, nil
}
