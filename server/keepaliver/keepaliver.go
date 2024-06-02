package keepaliver

import (
	"context"
	"io"
	"log"
	"math/rand/v2"
	"time"

	"github.com/airforce270/mc-srv/packet/config"
)

func New(interval time.Duration, w io.Writer) KeepAliver {
	return KeepAliver{
		interval: interval,
		w:        w,
	}
}

func NewForTesting(interval time.Duration, w io.Writer, rr rand.Source) KeepAliver {
	return KeepAliver{
		interval: interval,
		w:        w,
		rand:     rand.New(rr),
	}
}

type KeepAliver struct {
	interval time.Duration
	w        io.Writer
	rand     *rand.Rand
}

// StartPinging repeatedly sends pings until its context is cancelled.
// This function is blocking and should be run within a goroutine.
func (p *KeepAliver) StartPinging(ctx context.Context, logger *log.Logger) {
	ticker := time.NewTicker(1 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.Printf("Context done, ending keepalive pinging")
			return
		case <-ticker.C:
			logger.Printf("Sending keepalive")
			pack := config.ClientboundKeepAlive{KeepAliveID: p.randInt64()}
			if err := pack.Write(p.w); err != nil {
				logger.Printf("Failed to write config ping packet: %v", err)
			}
		}
		ticker.Reset(p.interval)
	}
}

func (p *KeepAliver) randInt64() int64 {
	rander := rand.Int64
	if p.rand != nil {
		rander = p.rand.Int64
	}
	return rander()
}
