package read_test

import (
	"bytes"
	"fmt"
	"slices"
	"testing"

	"github.com/airforce270/mc-srv/read"
	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
)

func TestBool(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input byte
		want  bool
	}{
		{0x00, false},
		{0x01, true},
	}

	for _, tc := range tests {
		t.Run(fmt.Sprintf("%x->%t", tc.input, tc.want), func(t *testing.T) {
			t.Parallel()

			got, err := read.Bool(bytes.NewReader([]byte{tc.input}))
			if err != nil {
				t.Fatalf("Bool() unexpected error: %v", err)
			}
			if got != tc.want {
				t.Errorf("Bool() = %t, want %t", got, tc.want)
			}
		})
	}
}

func TestByte(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input byte
		want  byte
	}{
		{'a', 'a'},
		{'3', '3'},
	}

	for _, tc := range tests {
		t.Run(fmt.Sprintf("%x->%x", tc.input, tc.want), func(t *testing.T) {
			t.Parallel()

			got, err := read.Byte(bytes.NewReader([]byte{tc.input}))
			if err != nil {
				t.Fatalf("Byte() unexpected error: %v", err)
			}
			if got != tc.want {
				t.Errorf("Byte() = %x, want %x", got, tc.want)
			}
		})
	}
}

func TestUnsignedShort(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input [2]byte
		want  uint16
	}{
		{[2]byte{0x0, 0x0}, 0},
		{[2]byte{0x4, 0x0}, 1024},
		{[2]byte{0x4, 0x4}, 1028},
	}

	for _, tc := range tests {
		t.Run(fmt.Sprintf("%x->%d", tc.input, tc.want), func(t *testing.T) {
			t.Parallel()

			got, err := read.UnsignedShort(bytes.NewReader(tc.input[:]))
			if err != nil {
				t.Fatalf("UnsignedShort() unexpected error: %v", err)
			}
			if got != tc.want {
				t.Errorf("UnsignedShort() = %d, want %d", got, tc.want)
			}
		})
	}
}

func TestLong(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input [8]byte
		want  int64
	}{
		{[8]byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}, 0},
		{[8]byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x04, 0xd2}, 1234},
		{[8]byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}, -1},
		{[8]byte{0xff, 0xff, 0xff, 0xff, 0xf8, 0xa4, 0x32, 0xeb}, -123456789},
	}

	for _, tc := range tests {
		t.Run(fmt.Sprintf("%x->%d", tc.input, tc.want), func(t *testing.T) {
			t.Parallel()

			got, err := read.Long(bytes.NewReader(tc.input[:]))
			if err != nil {
				t.Fatalf("Long() unexpected error: %v", err)
			}
			if got != tc.want {
				t.Errorf("Long() = %d, want %d", got, tc.want)
			}
		})
	}
}

func TestString(t *testing.T) {
	t.Parallel()

	const strWithLen128 = "hi this is a test of a somewhat longer string blah blah blah blah blah blah blah blah blah blah blah blah blah blah blah blah bl"

	tests := []struct {
		input []byte
		want  string
	}{
		{[]byte{0x00}, ""},
		{[]byte{0x02, 'h', 'i'}, "hi"},
		{slices.Concat([]byte{0x80, 0x01}, []byte(strWithLen128)), strWithLen128},
	}

	for _, tc := range tests {
		t.Run(fmt.Sprintf("%x->%s", tc.input, tc.want), func(t *testing.T) {
			t.Parallel()

			got, err := read.String(bytes.NewReader(tc.input))
			if err != nil {
				t.Fatalf("String() unexpected error: %v", err)
			}
			if got != tc.want {
				t.Errorf("String() = %s, want %s", got, tc.want)
			}
		})
	}
}

func TestVarInt(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input []byte
		want  int32
	}{
		{[]byte{0x00}, 0},
		{[]byte{0x01}, 1},
		{[]byte{0x02}, 2},
		{[]byte{0x7f}, 127},
		{[]byte{0x80, 0x01}, 128},
		{[]byte{0xe1, 0x01}, 225},
		{[]byte{0xe2, 0x01}, 226},
		{[]byte{0xff, 0x01}, 255},
		{[]byte{0xdd, 0xc7, 0x01}, 25565},
		{[]byte{0xff, 0xff, 0xff}, 2097151},
		{[]byte{0xff, 0xff, 0xff, 0xff, 0x07}, 2147483647},
		{[]byte{0xff, 0xff, 0xff, 0xff, 0x0f}, -1},
		{[]byte{0x80, 0x80, 0x80, 0x80, 0x08}, -2147483648},
	}

	for _, tc := range tests {
		t.Run(fmt.Sprintf("%x->%d", tc.input, tc.want), func(t *testing.T) {
			t.Parallel()

			got, err := read.VarInt(bytes.NewReader(tc.input))
			if err != nil {
				t.Fatalf("VarInt() unexpected error: %v", err)
			}
			if got != tc.want {
				t.Errorf("VarInt() = %d, want %d", got, tc.want)
			}
		})
	}
}

func TestUUID(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input []byte
		want  uuid.UUID
	}{
		{
			input: []byte{
				0x89, 0x96, 0xcb, 0x86, 0xcb, 0x63, 0x4c, 0x2d,
				0x8b, 0x45, 0x7c, 0xdf, 0xd7, 0xb5, 0x42, 0xc8,
			},
			want: uuid.MustParse("8996cb86-cb63-4c2d-8b45-7cdfd7b542c8"),
		},
	}

	for _, tc := range tests {
		t.Run(fmt.Sprintf("%x->%d", tc.input, tc.want), func(t *testing.T) {
			t.Parallel()

			got, err := read.UUID(bytes.NewReader(tc.input))
			if err != nil {
				t.Fatalf("UUID() unexpected error: %v", err)
			}
			if got != tc.want {
				t.Errorf("UUID() = %d, want %d", got, tc.want)
			}
		})
	}
}

func TestBytes(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input []byte
		want  []byte
	}{
		{[]byte{0x00}, []byte{0x00}},
		{[]byte{0x11, 0x12, 0x13}, []byte{0x11, 0x12, 0x13}},
	}

	for _, tc := range tests {
		t.Run(fmt.Sprintf("%x->%x", tc.input, tc.want), func(t *testing.T) {
			t.Parallel()

			got, err := read.Bytes(bytes.NewReader(tc.input), len(tc.input))
			if err != nil {
				t.Fatalf("Bytes() unexpected error: %v", err)
			}
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("Bytes() diff (-want, +got):\n%s", diff)
			}
		})
	}
}
