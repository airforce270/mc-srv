package packet

import (
	"bytes"
	"log"
	"testing"

	"github.com/airforce270/mc-srv/packet/headertest"
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
				PacketID: HandshakeID,
			},
		},
		{
			desc:  "ping header",
			input: headertest.PingLen9,
			want: Header{
				Length:   9,
				PacketID: PingRequestID,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()

			got, err := readHeader(bytes.NewReader(tc.input), log.Default())
			if err != nil {
				t.Fatalf("readHeader() unexpected err: %v", err)
			}

			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("readHeader() diff (-want, +got):\n%s", diff)
			}
		})
	}
}
