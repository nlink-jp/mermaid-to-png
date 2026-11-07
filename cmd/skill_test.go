package cmd

// The bundled mermaid-to-png skill is installed into an agent's skill
// directory and read as instruction, so a flag, exit status or engine kind
// it names that the command does not have is the agent being told something
// false. These tests hold the skill to the command in both directions
// (organization ADR-027 §2), taking the names from the flag set, the exit
// statuses and the engine — not from a copied list — and check that the
// rewrites it recommends are what the engine actually refuses and accepts.

import (
	"flag"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"

	mr "github.com/nlink-jp/mermaid-render"
)

const skillDir = "../mermaid-to-png"

func readSkill(t *testing.T) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(skillDir, "SKILL.md"))
	if err != nil {
		t.Fatalf("the skill has no SKILL.md: %v", err)
	}
	return string(b)
}

func TestSkillNameIsItsDirectory(t *testing.T) {
	s := readSkill(t)
	m := regexp.MustCompile(`(?m)^name: (.+)$`).FindStringSubmatch(s)
	if m == nil || m[1] != filepath.Base(skillDir) {
		t.Fatalf("frontmatter name %v, want %q", m, filepath.Base(skillDir))
	}
}

// Every flag the skill puts in backticks exists, and the flags a caller needs
// most are named.
func TestSkillFlagsExist(t *testing.T) {
	s := readSkill(t)
	var o options
	fs := NewFlagSet(&o)
	known := map[string]bool{"h": true} // flag's built-in help
	fs.VisitAll(func(f *flag.Flag) { known[f.Name] = true })

	named := map[string]bool{}
	// A span fenced by four backticks holds backticks itself; drop it so the
	// single-backtick spans pair up.
	s = regexp.MustCompile("````.*?````").ReplaceAllString(s, "")
	for _, m := range regexp.MustCompile("`[^`]*`").FindAllString(s, -1) {
		for _, f := range regexp.MustCompile(`(?:^|\s)-([a-z][a-z-]*)`).FindAllStringSubmatch(strings.Trim(m, "`"), -1) {
			named[f[1]] = true
			if !known[f[1]] {
				t.Errorf("the skill names -%s in %s; the command has no such flag", f[1], m)
			}
		}
	}
	for _, f := range []string{"o", "scale", "font"} {
		if !named[f] {
			t.Errorf("the skill does not name -%s", f)
		}
	}
}

// The skill's exit-status table lists exactly the statuses the command returns.
func TestSkillExitStatuses(t *testing.T) {
	s := readSkill(t)
	rows := regexp.MustCompile(`(?m)^\| (\d+) \|`).FindAllStringSubmatch(s, -1)
	var got []int
	for _, r := range rows {
		n, _ := strconv.Atoi(r[1])
		got = append(got, n)
	}
	if len(got) != len(ExitStatuses) {
		t.Fatalf("the skill's table lists %v, the command returns %v", got, ExitStatuses)
	}
	for i := range got {
		if got[i] != ExitStatuses[i] {
			t.Errorf("row %d is status %d, want %d", i, got[i], ExitStatuses[i])
		}
	}
}

// Every engine kind the message can carry is named, as the engine spells it.
func TestSkillNamesEveryKind(t *testing.T) {
	s := readSkill(t)
	for _, k := range []mr.ErrorKind{mr.SyntaxError, mr.UnsupportedType, mr.UnsupportedConstruct, mr.LayoutFault} {
		if !strings.Contains(s, "`"+k.String()+"`") {
			t.Errorf("the skill does not name the kind %q", k.String())
		}
	}
}

// Every type keyword the engine draws is named, and the unsupported examples
// are really unsupported.
func TestSkillTypes(t *testing.T) {
	s := readSkill(t)
	for _, kw := range []string{"flowchart", "graph", "sequenceDiagram", "erDiagram", "stateDiagram", "stateDiagram-v2", "gantt", "pie", "mindmap"} {
		if !strings.Contains(s, "`"+kw+"`") {
			t.Errorf("the skill does not name the type keyword %q", kw)
		}
		if _, err := mr.Parse(kw + "\n"); isKind(err, mr.UnsupportedType) {
			t.Errorf("%s is named as drawn, but the engine refuses the type", kw)
		}
	}
	for _, kw := range []string{"classDiagram", "block-beta"} {
		if !strings.Contains(s, "`"+kw+"`") {
			t.Errorf("the skill no longer names %q as an unsupported example", kw)
		}
		if _, err := mr.Parse(kw + "\n"); !isKind(err, mr.UnsupportedType) {
			t.Errorf("%s is named as unsupported, but the engine says %v", kw, err)
		}
	}
}

// The rewrites the skill recommends hold: what it says is refused is refused
// with the kind it says, and what it offers instead draws.
func TestSkillRewritesHold(t *testing.T) {
	needFont(t)
	cases := []struct {
		name, refused, accepted string
	}{
		{"HTML in a label", "flowchart TD\n  A[\"<b>太字</b>\"] --> B\n", "flowchart TD\n  A[\"太字\"] --> B\n"},
		{"line break", "", "flowchart TD\n  A[\"行1<br>行2\"] --> B\n"},
		{"Markdown string", "flowchart TD\n  A[\"`**md**`\"] --> B\n", "flowchart TD\n  A[\"md\"] --> B\n"},
		{"markdown formatting", "mindmap\n  root\n    **b**\n", "mindmap\n  root\n    b\n"},
		{"nested subgraph", "flowchart TD\n  subgraph X\n    subgraph Y\n      a --> b\n    end\n  end\n",
			"flowchart TD\n  subgraph X\n    a\n  end\n  subgraph Y\n    b\n  end\n  a --> b\n"},
		{"glyph", "flowchart TD\n  A[次へ ➔] --> B\n", "flowchart TD\n  A[次へ →] --> B\n"},
		{"styling is set aside", "", "%%{init: {\"theme\":\"dark\"}}%%\nflowchart TD\n  A --> B\n  style A fill:#f00\n  classDef c fill:#0f0\n"},
	}
	for _, c := range cases {
		if c.refused != "" {
			r := runCLI(t, c.refused, "-o", "-")
			if r.code != ExitNotDrawn || !strings.Contains(r.stderr, mr.UnsupportedConstruct.String()) {
				t.Errorf("%s: refused source gave exit %d, %q", c.name, r.code, r.stderr)
			}
		}
		if r := runCLI(t, c.accepted, "-o", "-"); r.code != ExitOK {
			t.Errorf("%s: the rewrite does not draw: exit %d, %q", c.name, r.code, r.stderr)
		}
	}
	if !strings.Contains(readSkill(t), "no font can draw") {
		t.Errorf("the skill no longer quotes the engine's missing-glyph message")
	}
}

func isKind(err error, k mr.ErrorKind) bool {
	e, ok := err.(*mr.Error)
	return ok && e.Kind == k
}
