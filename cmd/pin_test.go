package cmd

// Tests that pin behaviour the release review found unguarded: each fails
// when the behaviour it names is broken (checked by mutation when written).

import (
	"bytes"
	"errors"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"

	mr "github.com/nlink-jp/mermaid-render"
	"github.com/nlink-jp/mermaid-render/raster"
)

// Without the default font (any system but macOS), drawing stops with exit 3
// and a message that says to give -font.
func TestDefaultFontMissing(t *testing.T) {
	saved := defaultFont
	defaultFont = func() (*raster.Font, error) { return nil, errors.New("no Hiragino here") }
	t.Cleanup(func() { defaultFont = saved })

	r := runCLI(t, flowchart, "-o", "-")
	if r.code != ExitIO {
		t.Fatalf("exit %d, want %d (%s)", r.code, ExitIO, r.stderr)
	}
	for _, want := range []string{"the default font", "no Hiragino here", "give one with -font"} {
		if !strings.Contains(r.stderr, want) {
			t.Errorf("stderr %q lacks %q", r.stderr, want)
		}
	}
	if r.stdout != "" {
		t.Errorf("wrote %d bytes to stdout", len(r.stdout))
	}
}

// -o - writes one complete PNG and nothing after it.
func TestStdoutIsOnePNG(t *testing.T) {
	needFont(t)
	r := runCLI(t, flowchart, "-o", "-")
	if r.code != ExitOK {
		t.Fatalf("exit %d: %s", r.code, r.stderr)
	}
	if _, err := png.Decode(strings.NewReader(r.stdout)); err != nil {
		t.Fatalf("not a PNG: %v", err)
	}
	if iend := "IEND\xaeB`\x82"; !strings.HasSuffix(r.stdout, iend) {
		t.Errorf("bytes follow the PNG's end (last 16: %q)", r.stdout[max(0, len(r.stdout)-16):])
	}
}

// A refusal not tied to a line reads <input>: <kind>: <message>, with no :0.
func TestRefusalWithoutLine(t *testing.T) {
	needFont(t)
	r := runCLI(t, "", "-o", "-")
	want := "mermaid-to-png: <stdin>: " + mr.SyntaxError.String() + ": no diagram\n"
	if r.code != ExitNotDrawn || r.stderr != want {
		t.Errorf("exit %d, stderr %q, want %q", r.code, r.stderr, want)
	}
}

// The default scale is 2.
func TestDefaultScale(t *testing.T) {
	needFont(t)
	def := runCLI(t, flowchart, "-o", "-")
	two := runCLI(t, flowchart, "-o", "-", "-scale", "2")
	if def.code != ExitOK || two.code != ExitOK || !bytes.Equal([]byte(def.stdout), []byte(two.stdout)) {
		t.Errorf("the default and -scale 2 differ (exits %d, %d)", def.code, two.code)
	}
}

// A face name -font does not hold is an exit 3, naming what was asked for.
func TestFontNameNotFound(t *testing.T) {
	const ttc = "/System/Library/Fonts/Helvetica.ttc"
	if _, err := os.Stat(ttc); err != nil {
		t.Skipf("no %s on this system", ttc)
	}
	r := runCLI(t, flowchart, "-o", "-", "-font", ttc, "-font-name", "No Such Face")
	if r.code != ExitIO || !strings.Contains(r.stderr, "No Such Face") {
		t.Errorf("exit %d, stderr %q", r.code, r.stderr)
	}
}

// A usage error is reported before any input is read: it never waits on a
// terminal or a pipe.
type unreadable struct{ t *testing.T }

func (u unreadable) Read([]byte) (int, error) {
	u.t.Error("input was read before the arguments were checked")
	return 0, errors.New("unreadable")
}

func TestUsageBeforeInput(t *testing.T) {
	for name, args := range map[string][]string{
		"stdin without -o": {},
		"scale":            {"-o", "-", "-scale", "9"},
		"font-name alone":  {"-o", "-", "-font-name", "X"},
	} {
		var out, errb bytes.Buffer
		if code := Run(args, unreadable{t}, &out, &errb); code != ExitUsage {
			t.Errorf("%s: exit %d, want %d", name, code, ExitUsage)
		}
	}
}

// An output that reaches the input through a symbolic link is the input.
func TestOutputLinksToInput(t *testing.T) {
	in := writeInput(t, "d.mmd", flowchart)
	link := filepath.Join(filepath.Dir(in), "link.png")
	if err := os.Symlink(in, link); err != nil {
		t.Fatal(err)
	}
	r := runCLI(t, "", "-o", link, in)
	if r.code != ExitUsage || !strings.Contains(r.stderr, "is the input itself") {
		t.Errorf("exit %d, stderr %q", r.code, r.stderr)
	}
	if b, _ := os.ReadFile(in); string(b) != flowchart {
		t.Errorf("the input was overwritten")
	}
}

// An output that is the file standard output writes to is written through
// the stream, with nothing after the PNG.
func TestOutputIsStdoutFile(t *testing.T) {
	needFont(t)
	dst := filepath.Join(t.TempDir(), "out.png")
	f, err := os.Create(dst)
	if err != nil {
		t.Fatal(err)
	}
	var errb bytes.Buffer
	code := Run([]string{"-o", dst}, strings.NewReader(flowchart), f, &errb)
	f.Close()
	if code != ExitOK {
		t.Fatalf("exit %d: %s", code, errb.String())
	}
	b, _ := os.ReadFile(dst)
	if _, err := png.Decode(bytes.NewReader(b)); err != nil || !bytes.HasSuffix(b, []byte("IEND\xaeB`\x82")) {
		t.Errorf("not exactly one PNG (%v, last bytes %q)", err, b[max(0, len(b)-16):])
	}
}

// A device on standard input (/dev/null) is not a file the output could
// overwrite: the run reaches the engine, which finds no diagram.
func TestDeviceStdinIsNotTheOutput(t *testing.T) {
	needFont(t)
	null, err := os.Open(os.DevNull)
	if err != nil {
		t.Fatal(err)
	}
	defer null.Close()
	var out, errb bytes.Buffer
	if code := Run([]string{"-o", os.DevNull}, null, &out, &errb); code != ExitNotDrawn {
		t.Errorf("exit %d, want %d (%s)", code, ExitNotDrawn, errb.String())
	}
}

// Each I/O failure says which thing failed, not only that one did.
func TestIOErrorMessages(t *testing.T) {
	needFont(t)
	dir := t.TempDir()
	in := writeInput(t, "d.mmd", flowchart)
	for name, c := range map[string]struct {
		args []string
		want string
	}{
		"missing input":      {[]string{filepath.Join(dir, "none.mmd")}, "none.mmd"},
		"directory input":    {[]string{dir}, "is a directory"},
		"missing font":       {[]string{"-font", filepath.Join(dir, "none.ttf"), in}, "none.ttf"},
		"missing output dir": {[]string{"-o", filepath.Join(dir, "no", "out.png"), in}, "writing " + filepath.Join(dir, "no", "out.png")},
	} {
		r := runCLI(t, "", c.args...)
		if r.code != ExitIO || !strings.Contains(r.stderr, c.want) {
			t.Errorf("%s: exit %d, stderr %q, want %q", name, r.code, r.stderr, c.want)
		}
	}
}
