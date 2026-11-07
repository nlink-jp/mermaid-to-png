package cmd

import (
	"fmt"
	"io"
	"slices"
)

func run(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	if slices.Contains(args, "--version") || slices.Contains(args, "-version") {
		fmt.Fprintln(stdout, "mermaid-to-png", Version)
		return 0
	}
	fmt.Fprintln(stderr, "mermaid-to-png: not implemented yet")
	return 2
}
