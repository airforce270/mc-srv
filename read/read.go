// Package read reads values according to the Minecraft Client-Server protocol.
// https://wiki.vg/Protocol#Type
package read

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log"

	"github.com/airforce270/mc-srv/flags"
)

const (
	segmentBits = 0x7F
	continueBit = 0x80
)

var (
	errVarIntTooBig = errors.New("varint is too big")
)

// Byte reads a single byte from the reader.
func Byte(r io.Reader) (byte, error) {
	b, err := read(r, 1)
	if err != nil {
		return 0, fmt.Errorf("failed to read byte: %w", err)
	}

	return b[0], nil
}

// UnsignedShort reads an unsigned short from the reader.
func UnsignedShort(r io.Reader) (uint16, error) {
	b, err := read(r, 2)
	if err != nil {
		return 0, fmt.Errorf("failed to read bytes: %w", err)
	}

	return 0 | (uint16(b[0]) << 8) | uint16(b[1]), nil
}

// Long reads a long from the reader.
func Long(r io.Reader) (int64, error) {
	b, err := read(r, 8)
	if err != nil {
		return 0, fmt.Errorf("failed to read bytes: %w", err)
	}

	var val int64
	if err := binary.Read(bytes.NewReader(b), binary.BigEndian, &val); err != nil {
		return 0, fmt.Errorf("failed to read int64: %w", err)
	}
	return val, nil
}

// String reads a string from the reader.
func String(r io.Reader) (string, error) {
	length, err := VarInt(r)
	if err != nil {
		return "", fmt.Errorf("failed to read string's length: %w", err)
	}

	bytes, err := read(r, int(length))
	if err != nil {
		return "", fmt.Errorf("failed to read string bytes: %w", err)
	}

	return string(bytes), nil
}

// VarInt reads a VarInt from the reader.
func VarInt(r io.Reader) (int32, error) {
	var val int32
	var pos int32

	for {
		b, err := Byte(r)
		if err != nil {
			if errors.Is(err, io.EOF) {
				return val, nil
			}
			return 0, fmt.Errorf("failed to read byte for varint: %w", err)
		}
		val |= (int32(b) & segmentBits) << pos

		if b&continueBit == 0 {
			return val, nil
		}

		pos += 7

		if pos >= 32 {
			return 0, fmt.Errorf("val=%d pos=%d: %w", val, pos, errVarIntTooBig)
		}
	}
}

// Read the specified number of bytes from the reader.
func read(r io.Reader, count int) ([]byte, error) {
	buf := make([]byte, count)
	readCount, err := r.Read(buf)
	if err != nil {
		return nil, fmt.Errorf("failed to read %d bytes from reader: %w", count, err)
	}
	if readCount != count {
		return nil, fmt.Errorf("expected to read %d bytes, read %d: %w", count, readCount, err)
	}

	if *flags.Verbose {
		log.Printf("READ  %x (%d)", buf, readCount)
	}
	return buf, nil
}