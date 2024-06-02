package login_test

import (
	"bytes"
	"slices"
	"testing"

	"github.com/airforce270/mc-srv/packet"
	"github.com/airforce270/mc-srv/packet/id"
	"github.com/airforce270/mc-srv/packet/login"
	"github.com/airforce270/mc-srv/packet/login/logintest"
	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
)

func TestReadLoginStart(t *testing.T) {
	t.Parallel()

	inHeader := packet.Header{
		Length:   25,
		PacketID: id.LoginStart,
	}

	tests := []struct {
		desc  string
		input []byte
		want  login.LoginStart
	}{
		{
			desc:  "notchian example",
			input: logintest.NotchianLoginStart,
			want: login.LoginStart{
				Header:     inHeader,
				PlayerName: "airfors",
				PlayerUUID: uuid.MustParse("8996cb86-cb63-4c2d-8b45-7cdfd7b542c8"),
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()

			got, err := login.ReadLoginStart(bytes.NewReader(tc.input), inHeader)
			if err != nil {
				t.Fatalf("ReadLoginStart() unexpected err: %v", err)
			}

			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("ReadLoginStart() diff (-want, +got):\n%s", diff)
			}
		})
	}
}

func TestWriteEncryptionRequest(t *testing.T) {
	t.Parallel()

	tests := []struct {
		desc  string
		input login.EncryptionRequest
		want  []byte
	}{
		{
			desc: "standard",
			input: login.EncryptionRequest{
				ServerID:          "",
				PublicKeyLength:   3,
				PublicKey:         []byte{0x01, 0x02, 0x03},
				VerifyTokenLength: 5,
				VerifyToken:       []byte{0x01, 0x02, 0x03, 0x04, 0x05},
			},
			want: slices.Concat(
				// header
				[]byte{0x0c, 0x01},
				// payload
				[]byte{
					// server id
					0x00,
					// public key length
					0x03,
					// public key
					0x01, 0x02, 0x03,
					// verify token length
					0x05,
					// verify token
					0x01, 0x02, 0x03, 0x04, 0x05,
				},
			),
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()

			var out bytes.Buffer

			if err := tc.input.Write(&out); err != nil {
				t.Fatalf("WriteEncryptionRequest() unexpected err: %v", err)
			}

			got := out.Bytes()

			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("WriteEncryptionRequest() diff (-want, +got):\n%s", diff)
			}
		})
	}
}

func TestReadEncryptionResponse(t *testing.T) {
	t.Parallel()

	inHeader := packet.Header{
		Length:   11,
		PacketID: id.EncryptionResponse,
	}

	tests := []struct {
		desc  string
		input []byte
		want  login.EncryptionResponse
	}{
		{
			desc:  "notchian example",
			input: logintest.NotchianEncryptionResponse,
			want: login.EncryptionResponse{
				Header:             inHeader,
				SharedSecretLength: 5,
				SharedSecret:       []byte{0x01, 0x02, 0x03, 0x04, 0x05},
				VerifyTokenLength:  3,
				VerifyToken:        []byte{0x01, 0x02, 0x03},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()

			got, err := login.ReadEncryptionResponse(bytes.NewReader(tc.input), inHeader)
			if err != nil {
				t.Fatalf("ReadEncryptionResponse() unexpected err: %v", err)
			}

			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("ReadEncryptionResponse() diff (-want, +got):\n%s", diff)
			}
		})
	}
}
