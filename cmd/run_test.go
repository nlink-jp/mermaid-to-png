package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	mr "github.com/nlink-jp/mermaid-render"
	"github.com/nlink-jp/mermaid-render/raster"
)

var pngMagic = []byte("\x89PNG\r\n\x1a\n")

const flowchart = "flowchart TD\n  A[開始] --> B{判定}\n  B -->|はい| C[終わり]\n"

// needFont skips a drawing test where the default font (Hiragino) is absent:
// the engine cannot draw there without -font, which TestDefaultFontMissing
// covers by its exit status instead.
func needFont(t *testing.T) {
	t.Helper()
	if _, err := raster.DefaultFont(); err != nil {
		t.Skipf("no default font on this system: %v", err)
	}
}

type result struct {
	code           int
	stdout, stderr string
}

func runCLI(t *testing.T, stdin string, args ...string) result {
	t.Helper()
	var out, errb bytes.Buffer
	code := Run(args, strings.NewReader(stdin), &out, &errb)
	return result{code, out.String(), errb.String()}
}

func writeInput(t *testing.T, name, src string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(p, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

func assertPNG(t *testing.T, b []byte) {
	t.Helper()
	if !bytes.HasPrefix(b, pngMagic) {
		t.Fatalf("not a PNG (first bytes %q)", b[:min(len(b), 8)])
	}
}

// The default output is the input with .png, and its path is the one line on
// standard output.
func TestFileToDefaultOutput(t *testing.T) {
	needFont(t)
	in := writeInput(t, "diagram.mmd", flowchart)
	r := runCLI(t, "", in)
	if r.code != ExitOK {
		t.Fatalf("exit %d: %s", r.code, r.stderr)
	}
	want := strings.TrimSuffix(in, ".mmd") + ".png"
	if r.stdout != want+"\n" {
		t.Errorf("stdout %q, want %q", r.stdout, want+"\n")
	}
	b, err := os.ReadFile(want)
	if err != nil {
		t.Fatal(err)
	}
	assertPNG(t, b)
	// Nothing but the input and the output is left in the directory.
	ents, _ := os.ReadDir(filepath.Dir(in))
	if len(ents) != 2 {
		var names []string
		for _, e := range ents {
			names = append(names, e.Name())
		}
		t.Errorf("directory holds %v, want only the input and the PNG", names)
	}
}

func TestDefaultOutputName(t *testing.T) {
	needFont(t)
	for in, want := range map[string]string{
		"a.mermaid": "a.png",
		"noext":     "noext.png",
		"x.y.mmd":   "x.y.png",
	} {
		p := writeInput(t, in, flowchart)
		if r := runCLI(t, "", p); r.code != ExitOK {
			t.Fatalf("%s: exit %d: %s", in, r.code, r.stderr)
		}
		if _, err := os.Stat(filepath.Join(filepath.Dir(p), want)); err != nil {
			t.Errorf("%s: no %s: %v", in, want, err)
		}
	}
}

// Flags may stand before or after the file name; after -- everything is a
// file name, even one that starts with a dash.
func TestFlagsAroundTheFileName(t *testing.T) {
	needFont(t)
	dir := t.TempDir()
	in := writeInput(t, "d.mmd", flowchart)
	for name, args := range map[string][]string{
		"after":   {in, "-o", filepath.Join(dir, "after.png")},
		"between": {"-scale", "1", in, "-o", filepath.Join(dir, "between.png")},
	} {
		r := runCLI(t, "", args...)
		if r.code != ExitOK {
			t.Errorf("%s: exit %d: %s", name, r.code, r.stderr)
			continue
		}
		if _, err := os.Stat(filepath.Join(dir, name+".png")); err != nil {
			t.Errorf("%s: %v", name, err)
		}
	}
	dash := filepath.Join(dir, "-d.mmd")
	if err := os.WriteFile(dash, []byte(flowchart), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Chdir(dir)
	if r := runCLI(t, "", "-o", "dash.png", "--", "-d.mmd"); r.code != ExitOK {
		t.Errorf("after --: exit %d: %s", r.code, r.stderr)
	}
	if r := runCLI(t, "", in, in); r.code != ExitUsage {
		t.Errorf("two files: exit %d, want %d", r.code, ExitUsage)
	}
}

func TestStdinToFile(t *testing.T) {
	needFont(t)
	out := filepath.Join(t.TempDir(), "out.png")
	r := runCLI(t, flowchart, "-o", out)
	if r.code != ExitOK {
		t.Fatalf("exit %d: %s", r.code, r.stderr)
	}
	if r.stdout != out+"\n" {
		t.Errorf("stdout %q", r.stdout)
	}
	b, _ := os.ReadFile(out)
	assertPNG(t, b)
}

// The output is written the way a shell's > writes: an existing file keeps
// its mode and stays the same file, a link is written through, and a
// read-only file is refused and left as it was.
func TestExistingOutput(t *testing.T) {
	needFont(t)
	dir := t.TempDir()
	in := writeInput(t, "d.mmd", flowchart)

	priv := filepath.Join(dir, "priv.png")
	if err := os.WriteFile(priv, []byte("old"), 0o600); err != nil {
		t.Fatal(err)
	}
	before, _ := os.Stat(priv)
	if r := runCLI(t, "", "-o", priv, in); r.code != ExitOK {
		t.Fatalf("existing file: exit %d: %s", r.code, r.stderr)
	}
	after, _ := os.Stat(priv)
	if after.Mode().Perm() != 0o600 {
		t.Errorf("mode became %v, want 0600", after.Mode().Perm())
	}
	if !os.SameFile(before, after) {
		t.Errorf("the destination was replaced by another file")
	}

	target := filepath.Join(dir, "target.png")
	link := filepath.Join(dir, "link.png")
	if err := os.WriteFile(target, []byte("old"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, link); err != nil {
		t.Fatal(err)
	}
	if r := runCLI(t, "", "-o", link, in); r.code != ExitOK {
		t.Fatalf("symlink: exit %d: %s", r.code, r.stderr)
	}
	if st, _ := os.Lstat(link); st.Mode()&os.ModeSymlink == 0 {
		t.Errorf("the link was replaced by a regular file")
	}
	b, _ := os.ReadFile(target)
	assertPNG(t, b)

	ro := filepath.Join(dir, "ro.png")
	if err := os.WriteFile(ro, []byte("old"), 0o444); err != nil {
		t.Fatal(err)
	}
	if os.Geteuid() == 0 {
		t.Skip("root writes read-only files")
	}
	if r := runCLI(t, "", "-o", ro, in); r.code != ExitIO || !strings.Contains(r.stderr, "writing "+ro) {
		t.Errorf("read-only: exit %d, %q", r.code, r.stderr)
	}
	if b, _ := os.ReadFile(ro); string(b) != "old" {
		t.Errorf("the read-only file was changed")
	}
}

// A destination that is not a regular file is written like any other.
func TestNonRegularOutput(t *testing.T) {
	needFont(t)
	r := runCLI(t, flowchart, "-o", os.DevNull)
	if r.code != ExitOK {
		t.Fatalf("exit %d: %s", r.code, r.stderr)
	}
}

// A directory is not a destination.
func TestDirectoryOutput(t *testing.T) {
	needFont(t)
	if r := runCLI(t, flowchart, "-o", t.TempDir()); r.code != ExitIO {
		t.Errorf("exit %d, want %d (%s)", r.code, ExitIO, r.stderr)
	}
}

// -o - writes the PNG itself to standard output, and nothing else.
func TestStdoutOutput(t *testing.T) {
	needFont(t)
	for _, args := range [][]string{{"-o", "-"}, {"-o", "-", "-"}} {
		r := runCLI(t, flowchart, args...)
		if r.code != ExitOK {
			t.Fatalf("%v: exit %d: %s", args, r.code, r.stderr)
		}
		assertPNG(t, []byte(r.stdout))
	}
}

func TestScaleChangesSize(t *testing.T) {
	needFont(t)
	size := func(scale string) int {
		r := runCLI(t, flowchart, "-o", "-", "-scale", scale)
		if r.code != ExitOK {
			t.Fatalf("-scale %s: exit %d: %s", scale, r.code, r.stderr)
		}
		return len(r.stdout)
	}
	if small, large := size("1"), size("3"); small >= large {
		t.Errorf("-scale 1 gave %d bytes, -scale 3 gave %d", small, large)
	}
}

// Every keyword of the engine's 7 types draws.
func TestEveryDiagramKeyword(t *testing.T) {
	needFont(t)
	srcs := map[string]string{
		"flowchart":       "flowchart LR\n  A --> B\n",
		"graph":           "graph TD\n  A --> B\n",
		"flowchart-elk":   "flowchart-elk TD\n  A --> B\n",
		"sequenceDiagram": "sequenceDiagram\n  Alice->>Bob: こんにちは\n",
		"erDiagram":       "erDiagram\n  USER ||--o{ ORDER : places\n",
		"stateDiagram":    "stateDiagram\n  [*] --> Still\n  Still --> [*]\n",
		"stateDiagram-v2": "stateDiagram-v2\n  [*] --> Moving\n",
		"gantt":           "gantt\n  dateFormat YYYY-MM-DD\n  section A\n  Task :a1, 2026-01-01, 3d\n",
		"pie":             "pie\n  \"犬\" : 3\n  \"猫\" : 2\n",
		"mindmap":         "mindmap\n  root((中心))\n    枝1\n    枝2\n",
	}
	for kw, src := range srcs {
		r := runCLI(t, src, "-o", "-")
		if r.code != ExitOK {
			t.Errorf("%s: exit %d: %s", kw, r.code, r.stderr)
			continue
		}
		assertPNG(t, []byte(r.stdout))
	}
}

// An engine refusal ends with status 1 and reads <input>:<line>: <kind>: <message>.
func TestNotDrawn(t *testing.T) {
	needFont(t)
	cases := []struct {
		name, src, want string
	}{
		{"syntax", "flowchart TD\n  A -->\n", "<stdin>:2: " + mr.SyntaxError.String() + ": "},
		{"type", "classDiagram\n  A <|-- B\n", "<stdin>:1: " + mr.UnsupportedType.String() + ": classDiagram"},
		{"construct", "flowchart TD\n  A[\"<b>太字</b>\"] --> B\n", ": " + mr.UnsupportedConstruct.String() + ": "},
		{"glyph", "flowchart TD\n  A[🦜] --> B\n", ": " + mr.UnsupportedConstruct.String() + ": "},
	}
	for _, c := range cases {
		r := runCLI(t, c.src, "-o", "-")
		if r.code != ExitNotDrawn {
			t.Errorf("%s: exit %d, want %d (%s)", c.name, r.code, ExitNotDrawn, r.stderr)
			continue
		}
		if r.stdout != "" {
			t.Errorf("%s: wrote %d bytes to stdout", c.name, len(r.stdout))
		}
		if !strings.HasPrefix(r.stderr, "mermaid-to-png: ") || !strings.Contains(r.stderr, c.want) {
			t.Errorf("%s: stderr %q does not contain %q", c.name, r.stderr, c.want)
		}
	}
}

// A refused diagram leaves no file behind, and the file name is the input's.
func TestNotDrawnWritesNothing(t *testing.T) {
	needFont(t)
	in := writeInput(t, "bad.mmd", "classDiagram\n  A <|-- B\n")
	r := runCLI(t, "", in)
	if r.code != ExitNotDrawn {
		t.Fatalf("exit %d", r.code)
	}
	if !strings.Contains(r.stderr, in+":1: ") {
		t.Errorf("stderr %q does not name %s:1", r.stderr, in)
	}
	ents, _ := os.ReadDir(filepath.Dir(in))
	if len(ents) != 1 {
		t.Errorf("directory holds %d entries, want only the input", len(ents))
	}
}

func TestTooLong(t *testing.T) {
	// Over maxInputBytes: refused while reading, before the engine.
	r := runCLI(t, "flowchart TD\n"+strings.Repeat("A-->B\n", maxInputBytes/6+1), "-o", "-")
	if r.code != ExitNotDrawn || !strings.Contains(r.stderr, "longer than mermaid draws") {
		t.Errorf("exit %d, stderr %q", r.code, r.stderr)
	}
}

func TestUsageErrors(t *testing.T) {
	in := writeInput(t, "d.mmd", flowchart)
	png := writeInput(t, "d.png", flowchart)
	for name, args := range map[string][]string{
		"unknown flag":       {"-bogus", in},
		"two inputs":         {in, in},
		"stdin without -o":   {},
		"stdin dash, no -o":  {"-"},
		"scale zero":         {"-scale", "0", in},
		"scale negative":     {"-scale", "-1", in},
		"scale too large":    {"-scale", "9", in},
		"scale NaN":          {"-scale", "NaN", in},
		"scale not a number": {"-scale", "big", in},
		"font-name alone":    {"-font-name", "X", in},
		"bold alone":         {"-bold", "/x.ttf", in},
		"bold-name alone":    {"-font", "/x.ttf", "-bold-name", "X", in},
		"output is input":    {png},
		"output named input": {"-o", in, in},
	} {
		r := runCLI(t, flowchart, args...)
		if r.code != ExitUsage {
			t.Errorf("%s: exit %d, want %d (%s)", name, r.code, ExitUsage, r.stderr)
			continue
		}
		if !strings.Contains(r.stderr, usageLine) {
			t.Errorf("%s: stderr lacks the usage line: %q", name, r.stderr)
		}
	}
	// The input the refused run named is untouched.
	if b, _ := os.ReadFile(png); string(b) != flowchart {
		t.Errorf("the input was overwritten")
	}
}

func TestHelp(t *testing.T) {
	r := runCLI(t, "", "-h")
	if r.code != ExitOK || !strings.HasPrefix(r.stdout, usageLine) || !strings.Contains(r.stdout, "-scale") {
		t.Errorf("exit %d, stdout %q", r.code, r.stdout)
	}
}

func TestIOErrors(t *testing.T) {
	dir := t.TempDir()
	in := writeInput(t, "d.mmd", flowchart)
	for name, args := range map[string][]string{
		"missing input":      {filepath.Join(dir, "none.mmd")},
		"directory input":    {dir},
		"missing font":       {"-font", filepath.Join(dir, "none.ttf"), in},
		"not a font":         {"-font", in, in},
		"missing output dir": {"-o", filepath.Join(dir, "no", "such", "out.png"), in},
	} {
		r := runCLI(t, "", args...)
		if r.code != ExitIO {
			t.Errorf("%s: exit %d, want %d (%s)", name, r.code, ExitIO, r.stderr)
		}
	}
}

// Every status the exits table names is distinct, and the documented order
// is the one the README and the skill use.
func TestExitStatuses(t *testing.T) {
	want := []int{0, 1, 2, 3}
	if len(ExitStatuses) != len(want) {
		t.Fatalf("ExitStatuses %v", ExitStatuses)
	}
	for i, s := range ExitStatuses {
		if s != want[i] {
			t.Errorf("ExitStatuses[%d] = %d, want %d", i, s, want[i])
		}
	}
}
