// Package write handles writing packets back to the client.
package write

import (
	"encoding/binary"
	"fmt"
	"io"

	"github.com/google/uuid"
)

const (
	segmentBits = 0x7F
	continueBit = 0x80

	maxStrLen = (32767 * 3) + 3
)

var (
	// ErrStringTooLong is returned
	// if the string attempting to be encoded is too long.
	ErrStringTooLong = fmt.Errorf("string is too long (max=%d)", maxStrLen)
)

// Bool writes a boolean to the given writer.
func Bool(w io.Writer, b bool) error {
	var v byte = 0x00 // false
	if b {
		v = 0x01 // true
	}
	return Byte(w, v)
}

// Byte writes a byte to the given writer.
func Byte(w io.Writer, b byte) error {
	return Bytes(w, []byte{b})
}

// Int writes an int to the given writer.
func Int(w io.Writer, v int32) error {
	if err := binary.Write(w, binary.BigEndian, v); err != nil {
		return fmt.Errorf("failed to write int32 %d: %w", v, err)
	}

	return nil
}

// Long writes an int64 to the given writer.
func Long(w io.Writer, v int64) error {
	if err := binary.Write(w, binary.BigEndian, v); err != nil {
		return fmt.Errorf("failed to write long %d: %w", v, err)
	}

	return nil
}

// String writes a string to the given writer.
func String(w io.Writer, s string) error {
	l := int32(len(s))
	if l > maxStrLen {
		return ErrStringTooLong
	}
	if err := VarInt(w, l); err != nil {
		return fmt.Errorf("failed to write string's length (%d): %w", l, err)
	}

	if len(s) == 0 {
		return nil
	}

	if _, err := w.Write([]byte(s)); err != nil {
		return fmt.Errorf("failed to write string (%s): %w", s, err)
	}
	return nil
}

// VarInt writes a int32 to the given writer.
func VarInt(w io.Writer, v int32) error {
	for {
		if (v & ^segmentBits) == 0 {
			if err := Byte(w, byte(v)); err != nil {
				return fmt.Errorf("failed to write terminal varint byte %x: %w", v, err)
			}
			return nil
		}
		b := byte((v & segmentBits) | continueBit)
		if err := Byte(w, b); err != nil {
			return fmt.Errorf("failed to write varint %d: %w", b, err)
		}
		v = int32(uint32(v) >> 7)
	}
}

// VarIntLen returns the serialized len of the given varint.
func VarIntLen(v int32) int {
	var buf discardingWriter
	if err := VarInt(&buf, v); err != nil {
		panic(fmt.Sprintf("Failed to write varint, this should never happen"+
			" as discardingWriter does not return errors ever: %v", err))
	}
	return buf.Len()
}

// UUID writes a UUID to the given writer.
func UUID(w io.Writer, v uuid.UUID) error {
	b, err := v.MarshalBinary()
	if err != nil {
		return fmt.Errorf("failed to marshal UUID %s to bytes: %v", v, err)
	}
	return Bytes(w, b)
}

// Bytes writes the bytes to the given writer.
func Bytes(w io.Writer, b []byte) error {
	n, err := w.Write(b)
	if err != nil {
		return fmt.Errorf("failed to write byte %x: %w", b, err)
	}
	if n != len(b) {
		return fmt.Errorf("expected to write %d bytes, wrote %d", len(b), n)
	}
	return nil
}
