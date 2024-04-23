package packet_test

import (
	"bytes"
	"testing"

	"github.com/airforce270/mc-srv/packet"
	"github.com/airforce270/mc-srv/packet/configtest"
	"github.com/airforce270/mc-srv/packet/id"
	"github.com/google/go-cmp/cmp"
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
		want  packet.ConfigClientInformation
	}{
		{
			desc:  "notchian example",
			input: configtest.NotchianClientInformation,
			want: packet.ConfigClientInformation{
				Header:              inHeader,
				Locale:              "en_us",
				ViewDistance:        12,
				ChatMode:            packet.ChatModeEnabled,
				ChatColorsEnabled:   true,
				DisplayedSkinParts:  0b01111111,
				MainHand:            packet.MainHandRight,
				EnableTextFiltering: false,
				AllowServerListings: true,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()

			got, err := packet.ReadConfigClientInformation(bytes.NewReader(tc.input), inHeader)
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
		want  packet.ConfigServerboundPlugin
	}{
		{
			desc:  "notchian example",
			input: configtest.NotchianServerboundPlugin,
			want: packet.ConfigServerboundPlugin{
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

			got, err := packet.ReadConfigServerboundPlugin(bytes.NewReader(tc.input), inHeader)
			if err != nil {
				t.Fatalf("ReadConfigServerboundPlugin() unexpected err: %v", err)
			}

			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("ReadConfigServerboundPlugin() diff (-want, +got):\n%s", diff)
			}
		})
	}
}

// TODO: TestReadServerboundKeepAlive
// TODO: TestReadConfigPong
// TODO: TestReadConfigResourcePackResponse
