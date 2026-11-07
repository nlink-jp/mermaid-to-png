package cmd

import (
	"bytes"
	"io"
	"strings"
	"testing"
)

// The shared Homebrew formula template runs `<binary> --version` in its test
// block, and Go's flag package reads -version as the same flag: both spellings
// must print the same line, and it must carry the version.
func TestVersionSpellingsAgree(t *testing.T) {
	run := func(arg string) string {
		t.Helper()
		var out bytes.Buffer
		if code := Run([]string{arg}, strings.NewReader(""), &out, io.Discard); code != 0 {
			t.Fatalf("%s: exit status %d", arg, code)
		}
		return out.String()
	}
	long, short := run("--version"), run("-version")
	if want := "mermaid-to-png " + Version + "\n"; long != want {
		t.Errorf("--version printed %q, want %q", long, want)
	}
	if long != short {
		t.Errorf("--version printed %q but -version printed %q; they must agree", long, short)
	}
}
