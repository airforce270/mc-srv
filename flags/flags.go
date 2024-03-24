// Package flags contains top-level flags for the application.
package flags

import "flag"

var (
	// Verbose is whether verbose logging should be enabled.
	Verbose = flag.Bool("verbose", false, "Whether verbose logging should be enabled.")
)
