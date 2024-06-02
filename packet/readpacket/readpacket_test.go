package readpacket_test

import (
	"bytes"
	"fmt"
	"log"
	"slices"
	"testing"

	"github.com/airforce270/mc-srv/packet"
	"github.com/airforce270/mc-srv/packet/id"
	"github.com/airforce270/mc-srv/packet/login"
	"github.com/airforce270/mc-srv/packet/login/logintest"
	"github.com/airforce270/mc-srv/packet/pingtest"
	"github.com/airforce270/mc-srv/packet/readpacket"
	"github.com/airforce270/mc-srv/packet/slp"
	"github.com/airforce270/mc-srv/packet/slp/slptest"
	"github.com/airforce270/mc-srv/server/serverstate"
	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
)

func TestRead(t *testing.T) {
	t.Parallel()

	tests := []struct {
		state serverstate.State
		input []byte
		want  packet.Packet
	}{
		{
			state: serverstate.PreHandshake,
			input: slices.Concat(slptest.NotchianHandshakeHeader, slptest.NotchianHandshake),
			want: slp.Handshake{
				Header: packet.Header{
					Length:   16,
					PacketID: id.Handshake,
				},
				ProtocolVersion: 765,
				ServerAddress:   "127.0.0.1",
				ServerPort:      25565,
				NextState:       1,
			},
		},

		{
			state: serverstate.ClientRequestingStatus,
			input: slices.Concat(slptest.NotchianHandshakeHeader, slptest.NotchianHandshake),
			want: slp.Handshake{
				Header: packet.Header{
					Length:   16,
					PacketID: id.Handshake,
				},
				ProtocolVersion: 765,
				ServerAddress:   "127.0.0.1",
				ServerPort:      25565,
				NextState:       1,
			},
		},
		{
			state: serverstate.PreHandshake,
			input: slices.Concat(pingtest.NotchianHeader, pingtest.Notchian),
			want: slp.HandshakePingRequest{
				Header: packet.Header{
					Length:   9,
					PacketID: id.HandshakePing,
				},
				Payload: 6837160,
			},
		},
		{
			state: serverstate.ClientRequestingStatus,
			input: slices.Concat(pingtest.NotchianHeader, pingtest.Notchian),
			want: slp.HandshakePingRequest{
				Header: packet.Header{
					Length:   9,
					PacketID: id.HandshakePing,
				},
				Payload: 6837160,
			},
		},
		{
			state: serverstate.ClientRequestingLogin,
			input: slices.Concat(logintest.NotchianLoginStartHeader, logintest.NotchianLoginStart),
			want: login.LoginStart{
				Header: packet.Header{
					Length:   25,
					PacketID: id.LoginStart,
				},
				PlayerName: "airfors",
				PlayerUUID: uuid.MustParse("8996cb86-cb63-4c2d-8b45-7cdfd7b542c8"),
			},
		},
	}

	for _, tc := range tests {
		t.Run(fmt.Sprintf("[%d]%T", tc.state, tc.want), func(t *testing.T) {
			t.Parallel()

			got, err := readpacket.Read(bytes.NewReader(tc.input), tc.state, log.Default())
			if err != nil {
				t.Fatalf("Read() unexpected err: %v", err)
			}

			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("Read() diff (-want, +got):\n%s", diff)
			}
		})
	}
}
