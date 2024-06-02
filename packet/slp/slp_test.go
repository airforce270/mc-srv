package slp_test

import (
	"bytes"
	"testing"

	"github.com/airforce270/mc-srv/packet"
	"github.com/airforce270/mc-srv/packet/id"
	"github.com/airforce270/mc-srv/packet/pingtest"
	"github.com/airforce270/mc-srv/packet/slp"
	"github.com/airforce270/mc-srv/packet/slp/slptest"
	"github.com/google/go-cmp/cmp"
)

func TestReadHandshake(t *testing.T) {
	t.Parallel()

	inHeader := packet.Header{
		Length:   16,
		PacketID: id.Handshake,
	}

	tests := []struct {
		desc  string
		input []byte
		want  slp.Handshake
	}{
		{
			desc:  "notchian example",
			input: slptest.NotchianHandshake,
			want: slp.Handshake{
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

			got, err := slp.ReadHandshake(bytes.NewReader(tc.input), inHeader)
			if err != nil {
				t.Fatalf("ReadHandshake() unexpected err: %v", err)
			}

			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("ReadHandshake() diff (-want, +got):\n%s", diff)
			}
		})
	}
}

func TestReadHandshakePingRequest(t *testing.T) {
	t.Parallel()

	inHeader := packet.Header{
		Length:   9,
		PacketID: id.HandshakePing,
	}

	tests := []struct {
		desc  string
		input []byte
		want  slp.HandshakePingRequest
	}{
		{
			desc:  "notchian client example",
			input: pingtest.Notchian,
			want: slp.HandshakePingRequest{
				Header:  inHeader,
				Payload: 6837160,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()

			got, err := slp.ReadHandshakePingRequest(bytes.NewReader(tc.input), inHeader)
			if err != nil {
				t.Fatalf("ReadPing() unexpected err: %v", err)
			}

			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("ReadPing() diff (-want, +got):\n%s", diff)
			}
		})
	}
}

func TestWriteHandshakePingResponse(t *testing.T) {
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

			pr := slp.HandshakePingResponse{Payload: tc.input}

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
