package write_test

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/airforce270/mc-srv/write"
	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
)

func TestBool(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input bool
		want  []byte
	}{
		{true, []byte{0x01}},
		{false, []byte{0x00}},
	}

	for _, tc := range tests {
		t.Run(fmt.Sprintf("%t->%x", tc.input, tc.want), func(t *testing.T) {
			t.Parallel()
			var buf bytes.Buffer

			if err := write.Bool(&buf, tc.input); err != nil {
				t.Fatalf("Bool() unexpected error: %v", err)
			}
			got := buf.Bytes()
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("Bool() diff (-want, +got):\n%s", diff)
			}
		})
	}
}

func TestByte(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input byte
		want  []byte
	}{
		{'a', []byte{'a'}},
		{'3', []byte{'3'}},
	}

	for _, tc := range tests {
		t.Run(fmt.Sprintf("%x->%x", tc.input, tc.want), func(t *testing.T) {
			t.Parallel()
			var buf bytes.Buffer

			if err := write.Byte(&buf, tc.input); err != nil {
				t.Fatalf("Byte() unexpected error: %v", err)
			}
			got := buf.Bytes()
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("Byte() diff (-want, +got):\n%s", diff)
			}
		})
	}
}

func TestLong(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input int64
		want  []byte
	}{
		{0, []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}},
		{1234, []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x04, 0xd2}},
		{-1, []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}},
		{-123456789, []byte{0xff, 0xff, 0xff, 0xff, 0xf8, 0xa4, 0x32, 0xeb}},
	}

	for _, tc := range tests {
		t.Run(fmt.Sprintf("%x->%x", tc.input, tc.want), func(t *testing.T) {
			t.Parallel()
			var buf bytes.Buffer

			if err := write.Long(&buf, tc.input); err != nil {
				t.Fatalf("Long() unexpected error: %v", err)
			}
			got := buf.Bytes()
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("Long() diff (-want, +got):\n%s", diff)
			}
		})
	}
}

func TestString(t *testing.T) {
	t.Parallel()

	const strWithLen128 = "hi this is a test of a somewhat longer string blah blah blah blah blah blah blah blah blah blah blah blah blah blah blah blah bl"

	tests := []struct {
		input string
		want  []byte
	}{
		{"", []byte{0x00}},
		{"hi", []byte{0x02, 'h', 'i'}},
		{strWithLen128, append([]byte{0x80, 0x01}, []byte(strWithLen128)...)},
	}

	for _, tc := range tests {
		t.Run(fmt.Sprintf("%s->%x", tc.input, tc.want), func(t *testing.T) {
			t.Parallel()
			var buf bytes.Buffer

			if err := write.String(&buf, tc.input); err != nil {
				t.Fatalf("String() unexpected error: %v", err)
			}
			got := buf.Bytes()
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("String() diff (-want, +got):\n%s", diff)
			}
		})
	}
}

func TestVarInt(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input int32
		want  []byte
	}{
		{0, []byte{0x00}},
		{1, []byte{0x01}},
		{2, []byte{0x02}},
		{127, []byte{0x7f}},
		{128, []byte{0x80, 0x01}},
		{255, []byte{0xff, 0x01}},
		{25565, []byte{0xdd, 0xc7, 0x01}},
		{2097151, []byte{0xff, 0xff, 0x7f}},
		{2147483647, []byte{0xff, 0xff, 0xff, 0xff, 0x07}},
		{-1, []byte{0xff, 0xff, 0xff, 0xff, 0x0f}},
		{-2147483648, []byte{0x80, 0x80, 0x80, 0x80, 0x08}},
	}

	for _, tc := range tests {
		t.Run(fmt.Sprintf("%d->%x", tc.input, tc.want), func(t *testing.T) {
			t.Parallel()
			var buf bytes.Buffer

			if err := write.VarInt(&buf, tc.input); err != nil {
				t.Fatalf("VarInt() unexpected error: %v", err)
			}
			got := buf.Bytes()
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("VarInt() diff (-want, +got):\n%s", diff)
			}
		})
	}
}

func TestVarIntLen(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input int32
		want  int
	}{
		{0, 1},
		{1, 1},
		{2, 1},
		{127, 1},
		{128, 2},
		{255, 2},
		{25565, 3},
		{2097151, 3},
		{2147483647, 5},
		{-1, 5},
		{-2147483648, 5},
	}

	for _, tc := range tests {
		t.Run(fmt.Sprintf("%d->%d", tc.input, tc.want), func(t *testing.T) {
			t.Parallel()
			if got := write.VarIntLen(tc.input); got != tc.want {
				t.Errorf("VarIntLen() = %d, want %d", got, tc.want)
			}
		})
	}
}

func TestUUID(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input uuid.UUID
		want  []byte
	}{
		{
			input: uuid.MustParse("8996cb86-cb63-4c2d-8b45-7cdfd7b542c8"),
			want: []byte{
				0x89, 0x96, 0xcb, 0x86, 0xcb, 0x63, 0x4c, 0x2d,
				0x8b, 0x45, 0x7c, 0xdf, 0xd7, 0xb5, 0x42, 0xc8,
			},
		},
	}

	for _, tc := range tests {
		t.Run(fmt.Sprintf("%x->%x", tc.input, tc.want), func(t *testing.T) {
			t.Parallel()
			var buf bytes.Buffer

			if err := write.UUID(&buf, tc.input); err != nil {
				t.Fatalf("UUID() unexpected error: %v", err)
			}
			got := buf.Bytes()
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("UUID() diff (-want, +got):\n%s", diff)
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
			var buf bytes.Buffer

			if err := write.Bytes(&buf, tc.input); err != nil {
				t.Fatalf("Bytes() unexpected error: %v", err)
			}
			got := buf.Bytes()
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("Bytes() diff (-want, +got):\n%s", diff)
			}
		})
	}
}
