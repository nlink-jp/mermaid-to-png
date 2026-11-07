// Package cmd implements the mermaid-to-png command line.
//
// It takes no third-party dependencies beyond the organization's own engine
// (organization ADR-024), so the usual cobra scaffold is replaced with the
// standard library's flag package — which also keeps mmdpng's single-dash
// long flags (-scale, -font) working.
package cmd

import (
	"io"
	"os"
)

// Version is overridden at build time via -ldflags "-X .../cmd.Version=<vX.Y.Z>".
var Version = "dev"

// Execute runs the command line and exits with its status.
func Execute() {
	os.Exit(Run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}

// Run is the testable entry point behind Execute: it returns the exit status
// instead of exiting.
func Run(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	return run(args, stdin, stdout, stderr)
}
