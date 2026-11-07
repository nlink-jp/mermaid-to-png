# mermaid-to-png

[日本語](README.ja.md)

Turn one mermaid diagram into one PNG, on this machine and with no browser.
`mermaid-to-png` is a command-line front end to
[mermaid-render](https://github.com/nlink-jp/mermaid-render), the
organization's own mermaid engine, and ships with a skill that tells an agent
when to use it and what to do when a diagram is refused.

## Why

Without a command that draws mermaid, an agent asked for a diagram image
assembles one on the spot: headless Chrome, mermaid.js from a CDN, mmdc
through npx, or a figure drawn by hand. mermaid-render already draws mermaid
locally, but it is a library. This is its command.

## What it draws

The 7 types mermaid-render draws: flowchart (`flowchart`, `graph`),
sequence, ER, state, gantt, pie and mindmap, on a white card. Styling
(`style`, `classDef`, themes) is set aside. What the engine refuses, and why,
is in its [README](https://github.com/nlink-jp/mermaid-render#readme).

## Installation

The command and the skill install separately: Homebrew and the binary
archives carry the command only, and every release attaches the skill as its
own asset, `mermaid-to-png-vX.Y.Z.zip` (no platform in the name), with the
skill folder at the zip's root.

### The command

```bash
brew install nlink-jp/tap/mermaid-to-png
```

Or from source:

```bash
git clone https://github.com/nlink-jp/mermaid-to-png.git
cd mermaid-to-png
make build          # → dist/mermaid-to-png
```

Release archives for macOS (arm64, signed and notarized), Linux (amd64,
arm64) and Windows (amd64) are on the
[releases page](https://github.com/nlink-jp/mermaid-to-png/releases).
Text is drawn in Hiragino Sans, which only macOS has: elsewhere, give a font
with `-font`.

### The skill

Take `mermaid-to-png-vX.Y.Z.zip` from the release that matches your command.

- **Claude (desktop app / claude.ai)** — Customize → Skills → upload the zip.
  Upload the new zip again after each update.
- **Claude Code, by hand** — `unzip mermaid-to-png-vX.Y.Z.zip -d ~/.claude/skills/`

The skill needs the command where the agent runs. From a clone,
`make install-skill` is the development loop: it builds the zip and unpacks
that into `~/.claude/skills/` (`SKILLS_DEST=<dir>` for another directory).

## Usage

```bash
mermaid-to-png diagram.mmd                 # → diagram.png, prints its path
mermaid-to-png -o figure.png diagram.mmd
mermaid-to-png diagram.mmd -o figure.png   # flags may follow the file name
cat diagram.mmd | mermaid-to-png -o figure.png
mermaid-to-png -o - diagram.mmd > figure.png
mermaid-to-png -scale 1 diagram.mmd        # smaller image
mermaid-to-png -font /path/to/font.ttc -font-name "Face Name" diagram.mmd
```

| Flag | Meaning |
|---|---|
| `-o path` | output PNG; `-` for standard output. Default: the input with `.png`. Required when the source comes on standard input |
| `-scale n` | scale factor in (0, 8], default 2 (body text 28 px) |
| `-font path` | font file for the text (default Hiragino Sans W3) |
| `-font-name name` | face inside `-font`, by full or PostScript name |
| `-bold path`, `-bold-name name` | the face for bold text (default: the body face) |
| `-version` | print the version |

- The input is the diagram's source only (no ```` ```mermaid ```` fence).
  With no file name, or `-`, it is read from standard input, which is named
  `<stdin>` in messages; reading a terminal waits for input, as `cat` does.
  After `--` every argument is a file name.
- The output is written the way a shell's `>` writes: a file already at that
  path is replaced in place, keeping its permissions and links, and a
  read-only one is refused. An output that is the input itself — by name,
  through a link, or the file standard input is redirected from — is refused.
  `-o -` (or `-o /dev/stdout`) writes only the PNG to standard output. As
  with `>`, a run that fails while writing (exit 3) leaves the output file
  unusable.

### Exit status

| Status | Meaning |
|---|---|
| 0 | the PNG was written |
| 1 | the engine refused the diagram — `mermaid-to-png: <input>:<line>: <kind>: <message>`, or `<input>: <kind>: <message>` when no one line is at fault (a limit, an empty source) |
| 2 | a flag or argument (the usage line follows) |
| 3 | the input, the output or the font cannot be read, written or loaded |

`<kind>` is the engine's: `syntax error`, `unsupported diagram type`,
`unsupported construct` (also a limit exceeded, or a character no font can
draw) or `layout fault`.

## Skill (bundled): mermaid-to-png

The repository bundles the **mermaid-to-png** skill: when to draw a diagram
with this command, how to act on each exit status, what to rewrite when the
engine refuses a construct (HTML or Markdown in a label, a nested subgraph, a
character the font lacks), and when the work belongs to report-studio or to a
page's own mermaid.js instead. How to install it is under
[Installation](#the-skill).

`make test` runs the organization's skill validator. `cmd/skill_test.go`
holds the skill to the command: every flag it writes in backticks must exist,
its exit-status table must list exactly the statuses the command returns, it
must name all four engine kinds and every type keyword the engine draws, the
types it gives as unsupported must be refused, and each rewrite it
recommends must be refused and accepted by the engine as it says.

## Build

```bash
make build      # dist/mermaid-to-png
make test       # unit tests + skill validator
make package    # all platforms, darwin notarized, plus the skill zip
```

## Design

The design and its history are in the RFP:
[docs/en/mermaid-to-png-rfp.md](docs/en/mermaid-to-png-rfp.md)
([日本語](docs/ja/mermaid-to-png-rfp.ja.md)).

## License

MIT — see [LICENSE](LICENSE).
