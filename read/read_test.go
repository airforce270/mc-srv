package read_test

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/airforce270/mc-srv/read"
)

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
		tc := tc
		t.Run(fmt.Sprintf("%x->%x", tc.input, tc.want), func(t *testing.T) {
			t.Parallel()
			buf := bytes.NewBuffer([]byte{tc.input})

			got, err := read.Byte(buf)
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
		tc := tc
		t.Run(fmt.Sprintf("%x->%d", tc.input, tc.want), func(t *testing.T) {
			t.Parallel()
			buf := bytes.NewBuffer(tc.input[:])

			got, err := read.UnsignedShort(buf)
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
		tc := tc
		t.Run(fmt.Sprintf("%x->%d", tc.input, tc.want), func(t *testing.T) {
			t.Parallel()
			buf := bytes.NewBuffer(tc.input[:])

			got, err := read.Long(buf)
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
		{append([]byte{0x80, 0x01}, []byte(strWithLen128)...), strWithLen128},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(fmt.Sprintf("%x->%s", tc.input, tc.want), func(t *testing.T) {
			t.Parallel()
			buf := bytes.NewBuffer(tc.input[:])

			got, err := read.String(buf)
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
		tc := tc
		t.Run(fmt.Sprintf("%x->%d", tc.input, tc.want), func(t *testing.T) {
			t.Parallel()
			buf := bytes.NewBuffer(tc.input[:])

			got, err := read.VarInt(buf)
			if err != nil {
				t.Fatalf("VarInt() unexpected error: %v", err)
			}
			if got != tc.want {
				t.Errorf("VarInt() = %d, want %d", got, tc.want)
			}
		})
	}
}
