package keepaliver

import (
	"context"
	"io"
	"log"
	"math/rand/v2"
	"sync"
	"time"

	"github.com/airforce270/mc-srv/packet/config"
)

const (
	monitorInterval = 50 * time.Millisecond
)

// New creates a new KeepAliver.
func New(interval time.Duration, w io.Writer) KeepAliver {
	return KeepAliver{
		sendInterval:  interval,
		mustRespondIn: 5 * time.Second,
		w:             w,
		rand:          nil,
		pending:       map[int64]time.Time{},
		cancel:        make(chan struct{}, 10),
	}
}

// NewForTesting creates a new KeepAliver for testing.
// Notably, it allows providing a source of random data for predictability.
func NewForTesting(sendInterval, mustRespondIn time.Duration, w io.Writer, rr rand.Source) KeepAliver {
	return KeepAliver{
		sendInterval:  sendInterval,
		mustRespondIn: mustRespondIn,
		w:             w,
		rand:          rand.New(rr),
		pending:       map[int64]time.Time{},
		cancel:        make(chan struct{}, 10),
	}
}

// A KeepAliver sends KeepAlive packets to the client
// and notifies when the client doesn't respond in time.
type KeepAliver struct {
	sendInterval  time.Duration
	mustRespondIn time.Duration
	w             io.Writer
	rand          *rand.Rand

	pending    map[int64]time.Time
	pendingMtx sync.RWMutex // protects pending

	cancel chan struct{}
}

// StartPinging repeatedly sends keepalives until its context is cancelled.
// This function is blocking and should be run within a goroutine.
func (k *KeepAliver) StartPinging(ctx context.Context, logger *log.Logger) {
	go k.startMonitoring(ctx, logger)

	ticker := time.NewTicker(1 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.Printf("Context done, ending keepalive pinging")
			return
		case <-ticker.C:
			keepAliveID := k.randInt64()
			logger.Printf("Sending keepalive %d", keepAliveID)
			pack := config.ClientboundKeepAlive{KeepAliveID: keepAliveID}
			if err := pack.Write(k.w); err != nil {
				logger.Printf("Failed to write keepalive packet: %v", err)
			}
			k.pendingMtx.Lock()
			k.pending[keepAliveID] = time.Now()
			k.pendingMtx.Unlock()
		}
		ticker.Reset(k.sendInterval)
	}
}

// Receive marks a keepalive ID as received.
func (p *KeepAliver) Receive(id int64) {
	p.pendingMtx.Lock()
	delete(p.pending, id)
	p.pendingMtx.Unlock()
}

// Notifier returns a channel that is sent a value
// when the client doesn't respond in time.
func (p *KeepAliver) Notifier() <-chan struct{} {
	return p.cancel
}

// startMonitoring starts monitoring for keepalives
// that haven't been responded to in time.
func (p *KeepAliver) startMonitoring(ctx context.Context, logger *log.Logger) {
	ticker := time.NewTicker(1 * time.Millisecond)
	for {
		select {
		case <-ctx.Done():
			logger.Printf("Context done, ending keepalive monitoring")
			return
		case <-ticker.C:
			now := time.Now()
			p.pendingMtx.RLock()
			for id, sendTime := range p.pending {
				if diff := now.Sub(sendTime); diff > p.mustRespondIn {
					logger.Printf("Client didn't respond to keepalive %d in time", id)
					p.cancel <- struct{}{}
				}
			}
			p.pendingMtx.RUnlock()
		}
		ticker.Reset(monitorInterval)
	}
}

func (p *KeepAliver) randInt64() int64 {
	rander := rand.Int64
	if p.rand != nil {
		rander = p.rand.Int64
	}
	return rander()
}
