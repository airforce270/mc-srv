package server

import (
	"encoding/hex"
	"hash"
	"strings"
)

func minecraftDigest(h hash.Hash) string {
	hBytes := h.Sum(nil)

	negative := (hBytes[0] & 0x80) == 0x80
	if negative {
		hBytes = twosComplementLittleEndian(hBytes)
	}

	digest := strings.TrimLeft(hex.EncodeToString(hBytes), "0")
	if negative {
		return "-" + digest
	}
	return digest
}

func twosComplementLittleEndian(p []byte) []byte {
	carry := true
	for i := len(p) - 1; i >= 0; i-- {
		p[i] = byte(^p[i])

		if carry {
			carry = p[i] == 0xff
			p[i]++
		}
	}
	return p
}
