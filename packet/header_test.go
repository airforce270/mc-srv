package packet

import (
	"bytes"
	"testing"

	"github.com/airforce270/mc-srv/packet/headertest"
	"github.com/airforce270/mc-srv/packet/id"
	"github.com/google/go-cmp/cmp"
)

func TestReadHeader(t *testing.T) {
	t.Parallel()

	tests := []struct {
		desc  string
		input []byte
		want  Header
	}{
		{
			desc:  "handshake header",
			input: headertest.HandshakeLen16,
			want: Header{
				Length:   16,
				PacketID: id.Handshake,
			},
		},
		{
			desc:  "ping header",
			input: headertest.PingLen9,
			want: Header{
				Length:   9,
				PacketID: id.HandshakePing,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()

			got, err := ReadHeader(bytes.NewReader(tc.input))
			if err != nil {
				t.Fatalf("readHeader() unexpected err: %v", err)
			}

			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("readHeader() diff (-want, +got):\n%s", diff)
			}
		})
	}
}
