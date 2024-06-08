package keepaliver_test

import (
	"bytes"
	"context"
	"log"
	"testing"
	"time"

	"github.com/airforce270/mc-srv/packet"
	"github.com/airforce270/mc-srv/read"
	"github.com/airforce270/mc-srv/server/keepaliver"
)

type fakeRandSource struct {
	val uint64
}

func (s fakeRandSource) Uint64() uint64 { return s.val }

func TestSend(t *testing.T) {
	const dur = 50 * time.Millisecond
	const timeout = dur * 100
	const buffer = 25 * time.Millisecond
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	var buf bytes.Buffer

	source := fakeRandSource{val: 9999999999999999999}
	const want = 776627963145224191 // just so happens to be what the above val resolves to
	p := keepaliver.NewForTesting(dur, timeout, &buf, &source)

	go p.StartPinging(ctx, log.Default())

	const wait = dur + buffer
	time.Sleep(wait)
	cancel()

	h, err := packet.ReadHeader(&buf)
	if err != nil {
		t.Fatalf("Failed to read header: %v", err)
	}
	const wantLength = 9
	if h.Length != wantLength {
		t.Fatalf("Header length is %d, expected %d", h.Length, wantLength)
	}
	val, err := read.Long(&buf)
	if err != nil {
		t.Fatalf("Failed to read long from keepalive packet: %v", err)
	}
	if val != want {
		t.Errorf("KeepAliveID = %d, want %d", val, want)
	}
}
