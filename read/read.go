// Package read reads values according to the Minecraft Client-Server protocol.
// https://wiki.vg/Protocol#Type
package read

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"

	"github.com/google/uuid"
)

const (
	segmentBits = 0x7F
	continueBit = 0x80
)

var (
	errVarIntTooBig = errors.New("varint is too big")
)

// Bool reads a bool from the reader.
func Bool(r io.Reader) (bool, error) {
	b, err := Byte(r)
	if err != nil {
		return false, fmt.Errorf("failed to read bool: %w", err)
	}
	return b == 0x01, nil
}

// Byte reads a single byte from the reader.
func Byte(r io.Reader) (byte, error) {
	b, err := Bytes(r, 1)
	if err != nil {
		return 0, fmt.Errorf("failed to read byte: %w", err)
	}

	return b[0], nil
}

// UnsignedShort reads an unsigned short from the reader.
func UnsignedShort(r io.Reader) (uint16, error) {
	b, err := Bytes(r, 2)
	if err != nil {
		return 0, fmt.Errorf("failed to read bytes: %w", err)
	}

	return 0 | (uint16(b[0]) << 8) | uint16(b[1]), nil
}

// Int reads a int from the reader.
func Int(r io.Reader) (int32, error) {
	b, err := Bytes(r, 4)
	if err != nil {
		return 0, fmt.Errorf("failed to read bytes: %w", err)
	}

	var val int32
	if err := binary.Read(bytes.NewReader(b), binary.BigEndian, &val); err != nil {
		return 0, fmt.Errorf("failed to read int32: %w", err)
	}
	return val, nil
}

// Long reads a long from the reader.
func Long(r io.Reader) (int64, error) {
	b, err := Bytes(r, 8)
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

	bytes, err := Bytes(r, int(length))
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

// UUID reads a UUID from the reader.
//
// UUIDs are encoded as an unsigned 128-bit integer
// (or two unsigned 64-bit integers: the most sig 64 bits
// and then the least sig 64 bits)
func UUID(r io.Reader) (uuid.UUID, error) {
	bytes, err := Bytes(r, 16)
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("failed to read bytes for UUID: %w", err)
	}

	u, err := uuid.FromBytes(bytes)
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("failed to create UUID from bytes: %w", err)
	}

	return u, nil
}

// Bytes reads the specified number of bytes from the reader.
func Bytes(r io.Reader, count int) ([]byte, error) {
	if count == 0 {
		return nil, nil
	}

	buf := make([]byte, count)
	readCount, err := r.Read(buf)
	if err != nil {
		return nil, fmt.Errorf("failed to read %d bytes from reader: %w", count, err)
	}
	if readCount != count {
		return nil, fmt.Errorf("expected to read %d bytes, read %d: %w", count, readCount, err)
	}

	return buf, nil
}
