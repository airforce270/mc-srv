package packet_test

import (
	"bytes"
	"testing"

	"github.com/airforce270/mc-srv/packet"
	"github.com/airforce270/mc-srv/packet/pingtest"
	"github.com/google/go-cmp/cmp"
)

func TestReadPing(t *testing.T) {
	t.Parallel()

	inHeader := packet.Header{
		Length:   9,
		PacketID: packet.PingRequestID,
	}

	tests := []struct {
		desc  string
		input []byte
		want  packet.PingRequest
	}{
		{
			desc:  "notchian client example",
			input: pingtest.Notchian,
			want: packet.PingRequest{
				Header:  inHeader,
				Payload: 6837160,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()

			got, err := packet.ReadPingRequest(bytes.NewReader(tc.input), inHeader)
			if err != nil {
				t.Fatalf("ReadPing() unexpected err: %v", err)
			}

			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("ReadPing() diff (-want, +got):\n%s", diff)
			}
		})
	}
}

func TestWritePingResponse(t *testing.T) {
	t.Parallel()

	tests := []struct {
		desc  string
		input int64
		want  []byte
	}{
		{
			desc:  "standard",
			input: 8663213,
			want: []byte{
				0x09, 0x01, // header
				0x00, 0x00, 0x00, 0x00, 0x00, 0x84, 0x30, 0xad, // payload
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()

			pr := packet.PingResponse{Payload: tc.input}

			var out bytes.Buffer

			if err := pr.Write(&out); err != nil {
				t.Fatalf("WritePingResponse() unexpected err: %v", err)
			}

			got := out.Bytes()

			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("WritePingResponse() diff (-want, +got):\n%s", diff)
			}
		})
	}
}
