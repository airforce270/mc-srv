package server

import (
	"log"

	"github.com/airforce270/mc-srv/flags"
)

const maxLoggedBytes = 15

type readLogger struct {
	log *log.Logger
}

func (r readLogger) Write(b []byte) (int, error) {
	if *flags.Verbose {
		if len(b) > maxLoggedBytes {
			r.log.Printf("READ  %x... (%d)", b[:maxLoggedBytes], len(b))
		} else {
			r.log.Printf("READ  %x (%d)", b, len(b))
		}
	}
	return len(b), nil
}

type writeLogger struct {
	log *log.Logger
}

func (w writeLogger) Write(b []byte) (int, error) {
	if *flags.Verbose {
		if len(b) > maxLoggedBytes {
			w.log.Printf("WRITE %x... (%d)", b[:maxLoggedBytes], len(b))
		} else {
			w.log.Printf("WRITE %x (%d)", b, len(b))
		}
	}
	return len(b), nil
}
