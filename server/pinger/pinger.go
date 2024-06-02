package pinger

import (
	"context"
	"io"
	"log"
	"math/rand/v2"
	"time"

	"github.com/airforce270/mc-srv/packet/config"
)

func New(interval time.Duration, w io.Writer) Pinger {
	return Pinger{
		interval: interval,
		w:        w,
	}
}

func NewForTesting(interval time.Duration, w io.Writer, rr rand.Source) Pinger {
	return Pinger{
		interval: interval,
		w:        w,
		rand:     rand.New(rr),
	}
}

type Pinger struct {
	interval time.Duration
	w        io.Writer
	rand     *rand.Rand
}

// StartPinging repeatedly sends pings until its context is cancelled.
// This function is blocking and should be run within a goroutine.
func (p *Pinger) StartPinging(ctx context.Context, logger *log.Logger) {
	ticker := time.NewTicker(1 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.Printf("Context done, ending pinging")
			return
		case <-ticker.C:
			logger.Printf("Sending ping")
			pack := config.ConfigPing{PingID: p.randInt32()}
			if err := pack.Write(p.w); err != nil {
				logger.Printf("Failed to write config ping packet: %v", err)
			}
		}
		ticker.Reset(p.interval)
	}
}

func (p *Pinger) randInt32() int32 {
	rander := rand.Int32
	if p.rand != nil {
		rander = p.rand.Int32
	}
	return rander()
}
