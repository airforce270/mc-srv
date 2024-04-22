package packet_test

import (
	"bytes"
	"testing"

	"github.com/airforce270/mc-srv/packet"
	"github.com/airforce270/mc-srv/packet/handshaketest"
	"github.com/google/go-cmp/cmp"
)

func TestReadHandshake(t *testing.T) {
	t.Parallel()

	inHeader := packet.Header{
		Length:   16,
		PacketID: packet.HandshakeID,
	}

	tests := []struct {
		desc  string
		input []byte
		want  packet.Handshake
	}{
		{
			desc:  "notchian example",
			input: handshaketest.Notchian,
			want: packet.Handshake{
				Header:          inHeader,
				ProtocolVersion: 765,
				ServerAddress:   "127.0.0.1",
				ServerPort:      25565,
				NextState:       1,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()

			got, err := packet.ReadHandshake(bytes.NewReader(tc.input), inHeader)
			if err != nil {
				t.Fatalf("ReadHandshake() unexpected err: %v", err)
			}

			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("ReadHandshake() diff (-want, +got):\n%s", diff)
			}
		})
	}
}
