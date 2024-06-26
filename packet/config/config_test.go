package config_test

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/airforce270/mc-srv/packet"
	"github.com/airforce270/mc-srv/packet/config"
	"github.com/airforce270/mc-srv/packet/config/configtest"
	"github.com/airforce270/mc-srv/packet/id"
	"github.com/airforce270/mc-srv/packet/types"
	"github.com/airforce270/mc-srv/read"
	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
)

func TestReadConfigClientInformation(t *testing.T) {
	t.Parallel()

	inHeader := packet.Header{
		Length:   14,
		PacketID: id.ClientInformation,
	}

	tests := []struct {
		desc  string
		input []byte
		want  config.ConfigClientInformation
	}{
		{
			desc:  "notchian example",
			input: configtest.NotchianClientInformation,
			want: config.ConfigClientInformation{
				Header:              inHeader,
				Locale:              "en_us",
				ViewDistance:        12,
				ChatMode:            config.ChatModeEnabled,
				ChatColorsEnabled:   true,
				DisplayedSkinParts:  0b01111111,
				MainHand:            config.MainHandRight,
				EnableTextFiltering: false,
				AllowServerListings: true,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()

			got, err := config.ReadConfigClientInformation(bytes.NewReader(tc.input), inHeader)
			if err != nil {
				t.Fatalf("ReadConfigClientInformation() unexpected err: %v", err)
			}

			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("ReadConfigClientInformation() diff (-want, +got):\n%s", diff)
			}
		})
	}
}

func TestReadConfigServerboundPlugin(t *testing.T) {
	t.Parallel()

	inHeader := packet.Header{
		Length:   25,
		PacketID: id.ServerboundPlugin,
	}

	tests := []struct {
		desc  string
		input []byte
		want  config.ConfigServerboundPlugin
	}{
		{
			desc:  "notchian example",
			input: configtest.NotchianServerboundPlugin,
			want: config.ConfigServerboundPlugin{
				Header: inHeader,
				Data: []byte{
					15, 109, 105, 110, 101, 99, 114, 97, 102, 116, 58, 98,
					114, 97, 110, 100, 7, 118, 97, 110, 105, 108, 108, 97,
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()

			got, err := config.ReadConfigServerboundPlugin(bytes.NewReader(tc.input), inHeader)
			if err != nil {
				t.Fatalf("ReadConfigServerboundPlugin() unexpected err: %v", err)
			}

			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("ReadConfigServerboundPlugin() diff (-want, +got):\n%s", diff)
			}
		})
	}
}

func TestWriteDisconnect(t *testing.T) {
	in := types.TextComponent{Text: "some message"}

	p := config.Disconnect{Reason: in}

	var buf bytes.Buffer
	if err := p.Write(&buf); err != nil {
		t.Fatalf("Disconnect.Write() unexpected error writing: %v", err)
	}

	h, err := packet.ReadHeader(&buf)
	if err != nil {
		t.Fatalf("Disconnect.Write() unexpected error reading header: %v", err)
	}

	const wantLength = 25
	if h.Length != wantLength {
		t.Errorf("Disconnect.Write() header length = %d, want %d", h.Length, wantLength)
	}

	reasonStr, err := read.String(&buf)
	if err != nil {
		t.Fatalf("Disconnect.Write() unexpected error reading reason: %v", err)
	}

	var gotReason types.TextComponent
	if err := json.Unmarshal([]byte(reasonStr), &gotReason); err != nil {
		t.Fatalf("Disconnect.Write() unexpected error unmarshaling reason: %v", err)
	}

	if diff := cmp.Diff(in, gotReason); diff != "" {
		t.Errorf("Disconnect.Write() reason diff (-want, +got):\n%s", diff)
	}
}

func TestWriteFinishConfiguration(t *testing.T) {
	p := config.FinishConfiguration{}

	var buf bytes.Buffer
	if err := p.Write(&buf); err != nil {
		t.Fatalf("FinishConfiguration.Write() unexpected error writing: %v", err)
	}

	h, err := packet.ReadHeader(&buf)
	if err != nil {
		t.Fatalf("FinishConfiguration.Write() unexpected error reading header: %v", err)
	}

	const wantLength = 1
	if h.Length != wantLength {
		t.Errorf("FinishConfiguration.Write() header length = %d, want %d", h.Length, wantLength)
	}
}

func TestWriteClientboundKeepAlive(t *testing.T) {
	var keepAliveID int64 = 1234
	p := config.ClientboundKeepAlive{KeepAliveID: keepAliveID}

	var buf bytes.Buffer
	if err := p.Write(&buf); err != nil {
		t.Fatalf("ClientboundKeepAlive.Write() unexpected error writing: %v", err)
	}

	h, err := packet.ReadHeader(&buf)
	if err != nil {
		t.Fatalf("ClientboundKeepAlive.Write() unexpected error reading header: %v", err)
	}

	const wantLength = 9
	if h.Length != wantLength {
		t.Errorf("ClientboundKeepAlive.Write() header length = %d, want %d", h.Length, wantLength)
	}

	gotKeepAliveID, err := read.Long(&buf)
	if err != nil {
		t.Fatalf("ClientboundKeepAlive.Write() unexpected err reading id: %v", err)
	}
	if gotKeepAliveID != keepAliveID {
		t.Errorf("ClientboundKeepAlive.Write() keepalive ID = %d, want %d", gotKeepAliveID, keepAliveID)
	}
}

func TestReadServerboundKeepAlive(t *testing.T) {
	t.Parallel()

	inHeader := packet.Header{
		Length:   5,
		PacketID: id.ServerboundKeepAlive,
	}

	tests := []struct {
		desc  string
		input []byte
		want  config.ServerboundKeepAlive
	}{
		{
			desc:  "notchian example",
			input: configtest.NotchianServerboundKeepAlive,
			want: config.ServerboundKeepAlive{
				Header:      inHeader,
				KeepAliveID: 1717283416,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()

			got, err := config.ReadServerboundKeepAlive(bytes.NewReader(tc.input), inHeader)
			if err != nil {
				t.Fatalf("ReadServerboundKeepAlive() unexpected err: %v", err)
			}

			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("ReadServerboundKeepAlive() diff (-want, +got):\n%s", diff)
			}
		})
	}
}

func TestWriteConfigPing(t *testing.T) {
	var pingID int32 = 1234

	p := config.ConfigPing{PingID: pingID}

	var buf bytes.Buffer
	if err := p.Write(&buf); err != nil {
		t.Fatalf("ConfigPing.Write() Unexpected error writing: %v", err)
	}

	h, err := packet.ReadHeader(&buf)
	if err != nil {
		t.Fatalf("ConfigPing.Write() Unexpected error reading header: %v", err)
	}

	const wantLength = 5
	if h.Length != wantLength {
		t.Errorf("ConfigPing.Write() header length = %d, want %d", h.Length, wantLength)
	}

	gotPingID, err := read.Int(&buf)
	if err != nil {
		t.Fatalf("ConfigPing.Write() Unexpected error reading reason: %v", err)
	}

	if gotPingID != pingID {
		t.Errorf("ConfigPing.Write() PingID = %d, want %d", gotPingID, pingID)
	}
}

func TestReadConfigResourcePackResponse(t *testing.T) {
	t.Parallel()

	inHeader := packet.Header{
		Length:   17,
		PacketID: id.ResourcePackResponse,
	}

	tests := []struct {
		desc  string
		input []byte
		want  config.ConfigResourcePackResponse
	}{
		{
			desc:  "notchian example",
			input: configtest.NotchianConfigResourcePackResponse,
			want: config.ConfigResourcePackResponse{
				Header:           inHeader,
				ResourcePackUUID: uuid.MustParse("8996cb86-cb63-4c2d-8b45-7cdfd7b542c8"),
				Result:           config.ResourcePackResultFailedToDownload,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()

			got, err := config.ReadConfigResourcePackResponse(bytes.NewReader(tc.input), inHeader)
			if err != nil {
				t.Fatalf("ReadConfigResourcePackResponse() unexpected err: %v", err)
			}

			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("ReadConfigResourcePackResponse() diff (-want, +got):\n%s", diff)
			}
		})
	}
}
