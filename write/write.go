// Package write handles writing packets back to the client.
package write

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"log"

	"github.com/airforce270/mc-srv/flags"
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

// Byte writes a byte to the given writer.
func Byte(w io.Writer, b byte) error {
	_, err := w.Write([]byte{b})
	if err != nil {
		return fmt.Errorf("failed to write byte %x: %w", b, err)
	}
	if *flags.Verbose {
		log.Printf("WRITE %x (1)", []byte{b})
	}
	return nil
}

// Long writes an int64 to the given writer.
func Long(w io.Writer, v int64) error {
	var logBuf bytes.Buffer
	mw := io.MultiWriter(w, &logBuf)
	if err := binary.Write(mw, binary.BigEndian, v); err != nil {
		return fmt.Errorf("failed to write long %d: %w", v, err)
	}
	if *flags.Verbose {
		length := logBuf.Len()
		wrote := logBuf.Bytes()
		log.Printf("WRITE %x (%d)", wrote, length)
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
	if _, err := w.Write([]byte(s)); err != nil {
		return fmt.Errorf("failed to write string (%s): %w", s, err)
	}
	if *flags.Verbose {
		log.Printf("WRITE %x (%d)", []byte(s), len([]byte(s)))
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

// VarInt writes a int32 to the given writer.
func VarIntLen(v int32) (int, error) {
	var buf bytes.Buffer
	err := VarInt(&buf, v)
	return buf.Len(), err
}
