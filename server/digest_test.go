package server

import (
	"crypto/sha1"
	"fmt"
	"io"
	"testing"
)

func TestMinecraftDigest(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"Notch", "4ed1f46bbe04bc756bcb17c0c7ce3e4632f06a48"},
		{"jeb_", "-7c9d5b0044c130109a5d7b5fb5c317c02b4e28c1"},
		{"simon", "88e16a1019277b15d58faf0541e11910eb756f6"},
	}

	for _, tc := range tests {
		t.Run(fmt.Sprintf("%s->%s", tc.input, tc.want), func(t *testing.T) {
			in := sha1.New()
			if _, err := io.WriteString(in, tc.input); err != nil {
				t.Fatalf("Failed to write input to sha1: %v", err)
			}

			if got := minecraftDigest(in); got != tc.want {
				t.Errorf("minecraftDigest() = %q, want %q", got, tc.want)
			}
		})
	}
}
