package cmd

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"image/png"
	"io"
	"math"
	"os"
	"path/filepath"
	"strings"

	mr "github.com/nlink-jp/mermaid-render"
	"github.com/nlink-jp/mermaid-render/raster"
)

// Exit statuses. The bundled skill names them, and its test holds it to this
// list (organization ADR-027 §2).
const (
	ExitOK       = 0 // the PNG was written
	ExitNotDrawn = 1 // the engine refused the diagram: syntax, unsupported, over a limit, no glyph, layout fault
	ExitUsage    = 2 // flags or arguments
	ExitIO       = 3 // cannot read the input, write the output, or load a font
)

// ExitStatuses lists every status the command line can return.
var ExitStatuses = []int{ExitOK, ExitNotDrawn, ExitUsage, ExitIO}

const usageLine = "usage: mermaid-to-png [-o out.png] [-scale 2] [-font path [-font-name name]] [-bold path [-bold-name name]] [file.mmd | -]"

// maxInputBytes bounds what is read. A UTF-16 unit takes at most 3 bytes of
// UTF-8, so a longer input is certainly over the engine's MaxTextSize; the
// engine itself judges anything shorter.
const maxInputBytes = 3 * mr.MaxTextSize

type options struct {
	out     string
	scale   float64
	font    raster.FontSpec
	version bool
}

// NewFlagSet defines the command line's flags. The skill's test takes the
// flag names from here, not from a copied list.
func NewFlagSet(o *options) *flag.FlagSet {
	fs := flag.NewFlagSet("mermaid-to-png", flag.ContinueOnError)
	fs.StringVar(&o.out, "o", "", "output PNG; - for standard output (default: the input with .png)")
	fs.Float64Var(&o.scale, "scale", 2, "scale factor, in (0, 8]")
	fs.StringVar(&o.font.Path, "font", "", "font file for body text (default: Hiragino Sans W3)")
	fs.StringVar(&o.font.Name, "font-name", "", "face in -font, by full or PostScript name")
	fs.StringVar(&o.font.BoldPath, "bold", "", "font file for bold text (default: the body face)")
	fs.StringVar(&o.font.BoldName, "bold-name", "", "face in -bold")
	fs.BoolVar(&o.version, "version", false, "print the version and exit")
	return fs
}

// failure is an error with the exit status it ends the run with.
type failure struct {
	status int
	msg    string
}

func (f *failure) Error() string { return f.msg }

func fail(status int, format string, a ...any) error {
	return &failure{status: status, msg: fmt.Sprintf(format, a...)}
}

// parseArgs reads flags wherever they stand, before or after the file name
// (`diagram.mmd -o out.png` as well as `-o out.png diagram.mmd`): the flag
// package stops at the first non-flag, so parsing resumes after each one.
// Everything after `--` is a file name.
func parseArgs(fs *flag.FlagSet, args []string) ([]string, error) {
	var files []string
	for {
		if err := fs.Parse(args); err != nil {
			return nil, err
		}
		rest := fs.Args()
		if len(rest) == 0 {
			return files, nil
		}
		if used := len(args) - len(rest); used > 0 && args[used-1] == "--" {
			return append(files, rest...), nil
		}
		files = append(files, rest[0])
		args = rest[1:]
	}
}

func run(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	var o options
	fs := NewFlagSet(&o)
	fs.SetOutput(io.Discard)
	files, err := parseArgs(fs, args)
	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			fmt.Fprintln(stdout, usageLine)
			fs.SetOutput(stdout)
			fs.PrintDefaults()
			return ExitOK
		}
		fmt.Fprintf(stderr, "mermaid-to-png: %v\n%s\n", err, usageLine)
		return ExitUsage
	}
	if o.version {
		fmt.Fprintln(stdout, "mermaid-to-png", Version)
		return ExitOK
	}
	if err := convert(&o, files, stdin, stdout); err != nil {
		fmt.Fprintf(stderr, "mermaid-to-png: %s\n", err)
		var f *failure
		if errors.As(err, &f) {
			if f.status == ExitUsage {
				fmt.Fprintln(stderr, usageLine)
			}
			return f.status
		}
		return ExitNotDrawn
	}
	return ExitOK
}

// convert reads the source, has the engine draw it, and writes the PNG.
func convert(o *options, args []string, stdin io.Reader, stdout io.Writer) error {
	if len(args) > 1 {
		return fail(ExitUsage, "one input at a time (got %d)", len(args))
	}
	in := "-"
	if len(args) == 1 {
		in = args[0]
	}
	fromStdin := in == "-"
	label := in
	if fromStdin {
		label = "<stdin>"
	}

	// Arguments first, so a usage error never waits on input.
	if math.IsNaN(o.scale) || o.scale <= 0 || o.scale > raster.MaxScale {
		return fail(ExitUsage, "-scale %v is outside (0, %d]", o.scale, raster.MaxScale)
	}
	if o.font.Path == "" && (o.font.Name != "" || o.font.BoldPath != "" || o.font.BoldName != "") {
		return fail(ExitUsage, "-font-name, -bold and -bold-name need -font")
	}
	if o.font.BoldPath == "" && o.font.BoldName != "" {
		return fail(ExitUsage, "-bold-name needs -bold")
	}
	out := o.out
	if out == "" {
		if fromStdin {
			return fail(ExitUsage, "reading standard input needs -o (a file, or - for standard output)")
		}
		out = strings.TrimSuffix(in, filepath.Ext(in)) + ".png"
	}
	toStdout := out == "-"
	if !fromStdin && !toStdout && samePath(in, out) {
		return fail(ExitUsage, "the output %s is the input itself", out)
	}
	if fromStdin && !toStdout && isFile(stdin, out) {
		return fail(ExitUsage, "the output %s is the file standard input reads", out)
	}
	src, err := readSource(in, label, stdin)
	if err != nil {
		return err
	}
	font, err := loadFont(o.font)
	if err != nil {
		return err
	}
	img, err := raster.RenderSource(src, raster.Options{Font: font, Scale: o.scale})
	if err != nil {
		return notDrawn(label, err)
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return fail(ExitIO, "encoding the PNG: %v", err)
	}
	if toStdout {
		if _, err := stdout.Write(buf.Bytes()); err != nil {
			return fail(ExitIO, "writing standard output: %v", err)
		}
		return nil
	}
	if err := writeFile(out, buf.Bytes()); err != nil {
		return err
	}
	fmt.Fprintln(stdout, out)
	return nil
}

// notDrawn reports an engine refusal as <input>:<line>: <kind>: <message>.
func notDrawn(label string, err error) error {
	var e *mr.Error
	if !errors.As(err, &e) {
		return fail(ExitNotDrawn, "%s: %v", label, err)
	}
	where := label
	if e.Line > 0 {
		where = fmt.Sprintf("%s:%d", label, e.Line)
	}
	return fail(ExitNotDrawn, "%s: %s: %s", where, e.Kind, e.Msg)
}

// readSource reads the input. One over the bound is refused in the engine's
// form, <input>: unsupported construct: …, like the engine's own refusal of
// a source just under the bound.
func readSource(in, label string, stdin io.Reader) (string, error) {
	r := stdin
	if in != "-" {
		f, err := os.Open(in)
		if err != nil {
			return "", fail(ExitIO, "%v", err)
		}
		defer f.Close()
		if st, err := f.Stat(); err == nil && st.IsDir() {
			return "", fail(ExitIO, "%s is a directory", in)
		}
		r = f
	}
	b, err := io.ReadAll(io.LimitReader(r, maxInputBytes+1))
	if err != nil {
		return "", fail(ExitIO, "reading %s: %v", label, err)
	}
	if len(b) > maxInputBytes {
		return "", fail(ExitNotDrawn, "%s: %s: the source is over %d bytes (mermaid draws at most %d characters)",
			label, mr.UnsupportedConstruct, maxInputBytes, mr.MaxTextSize)
	}
	return string(b), nil
}

func loadFont(spec raster.FontSpec) (*raster.Font, error) {
	var (
		f   *raster.Font
		err error
	)
	if spec.Path != "" {
		f, err = raster.LoadFont(spec)
	} else {
		f, err = raster.DefaultFont()
	}
	if err != nil {
		if spec.Path == "" {
			return nil, fail(ExitIO, "font: the default font (Hiragino Sans) cannot be loaded: %v; give one with -font", err)
		}
		return nil, fail(ExitIO, "%v", err) // the engine's message names the font
	}
	return f, nil
}

// writeFile writes the PNG to its destination the way a shell's > does: an
// existing file keeps its permissions, links and ownership, a read-only one
// is refused, and a new one gets 0666 less the umask. The PNG is complete in
// memory before the file is opened, so only the disk itself can fail midway.
func writeFile(path string, data []byte) error {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o666)
	if err != nil {
		return fail(ExitIO, "writing %s: %v", path, err)
	}
	_, err = f.Write(data)
	if cerr := f.Close(); err == nil {
		err = cerr
	}
	if err != nil {
		return fail(ExitIO, "writing %s: %v", path, err)
	}
	return nil
}

// isFile reports whether r is an open file that is the file at path. A
// redirect (< diagram.mmd) is such a file; a pipe from cat cannot be told
// apart from any other pipe and is not.
func isFile(r io.Reader, path string) bool {
	f, ok := r.(*os.File)
	if !ok {
		return false
	}
	a, err1 := f.Stat()
	b, err2 := os.Stat(path)
	return err1 == nil && err2 == nil && os.SameFile(a, b)
}

func samePath(a, b string) bool {
	aa, err1 := filepath.Abs(a)
	bb, err2 := filepath.Abs(b)
	if err1 != nil || err2 != nil {
		return false
	}
	if aa == bb {
		return true
	}
	sa, err1 := os.Stat(aa)
	sb, err2 := os.Stat(bb)
	return err1 == nil && err2 == nil && os.SameFile(sa, sb)
}
