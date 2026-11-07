# RFP: mermaid-to-png

> Generated: 2026-11-07 (second draft)
> Status: Draft

## 1. Problem Statement

This machine has no command that turns a mermaid diagram into an image. The organization's own engine, mermaid-render
(lib-series), is a library, not a command, and its only command, `mmdpng`, is the library's debugging tool and is not
distributed — because the mermaid-render RFP (2026-09-28) decided not to make a distributed CLI. With no tool, every time
Claude Code wants a mermaid diagram as an image (to put a figure in a document, slides or a note, or to check a diagram
it wrote) it assembles a way on the spot: it looks for headless Chrome, mermaid.js from a CDN, npx or mmdc, or draws the
figure by hand. mermaid-to-png releases, as one project, a CLI that is a front end to the engine and a skill that guides
its use. Installed with brew, the CLI puts the means on PATH; the skill teaches that this is what to use when mermaid has
to become an image, and what to do with a diagram it cannot draw.

**Target users**: Claude Code (the skill and the CLI). People may use it directly from a shell or a script.

**Scope boundary**: making a whole Word document or PDF with diagrams in it belongs to report-studio-mcp. mermaid-to-png
turns one diagram into one image file.

## 2. Functional Specification

### Commands / API Surface

```
mermaid-to-png [-o out.png] [-scale 2] [-font path [-font-name name]] [-bold path [-bold-name name]] [file.mmd | -]
mermaid-to-png -version
```

- Flags are read with the standard library's `flag` (single dash, as in mmdpng).
- Input is a file, or standard input (`-`, or no file name).
- Output defaults to the input file name with its extension changed to `.png`; an output that would be the input's own
  path is refused. Reading from standard input requires `-o`. `-o -` writes the PNG to standard output (for pipes).
- `-scale` is in (0, 8], default 2 (the engine's range and default; out of range is refused before the engine is called).
- The font flags are passed to the engine's `raster.FontSpec` as they are. Without them, Hiragino Sans W3/W6.
  `-font-name`, `-bold` and `-bold-name` are refused without `-font` (as in mmdpng).
- On success, prints the path of the written file on one line to standard output (nothing with `-o -`).

### Input / Output

- Input: mermaid text. Limit: the engine's `MaxTextSize` (50,000 UTF-16 units).
- Output: PNG (a white card, exactly as the engine draws it).
- Failure output (standard error): `mermaid-to-png: <input>:<line>: <kind>: <message>`, where `<kind>` is the name of the
  engine's `ErrorKind` (`syntax error`, `unsupported diagram type`, `unsupported construct`, `layout fault`). The line is
  left out when it is not known.
- Exit status:

| Status | Meaning |
|---|---|
| 0 | written |
| 1 | the diagram could not be drawn (the engine returned a `*mermaidrender.Error`; exceeded limits and characters no font can draw land here too) |
| 2 | usage error (flags, arguments) |
| 3 | input/output or font failure (cannot read, cannot write, cannot load a font) |

Details (how `<input>` reads for standard input, what happens when input is a terminal, and the like) are settled in
Phase 1 and written in the README and the tests.

### Skill

`mermaid-to-png/SKILL.md` at the repository root (skills-series layout, organization ADR-027 §2). It holds only the craft
the CLI's `-h` cannot carry:

- **When to use it**: whenever a mermaid diagram has to become an image file (to put it in a document, slides or a note,
  or to look at a diagram just written). Once the image exists, reading that PNG shows the diagram.
- **What it draws**: the engine's 7 types (flowchart, sequence, ER, state, gantt, pie, mindmap).
- **On exit status 1**: rewrite the construct it cannot draw (what was actually needed: drop `<b>` and `<code>` from
  labels, flatten nested subgraphs, replace characters the font lacks), or, for a type it does not draw, give the user
  the source and ask how to proceed.
- **Hand-offs**: a whole document with diagrams goes to report-studio-mcp. Mermaid inside an HTML Artifact is drawn by
  that page's own mermaid.js and is not this tool's job.
- Written in the positive (no "do not use X" — knowledge `llm-integration.md`, "a 'do not' over-generalizes"). The
  description stays within 1,024 characters.

### Configuration

None.

### External Dependencies

None (no network, outside service or credentials). The only Go dependency is `github.com/nlink-jp/mermaid-render`
(`golang.org/x/image`, `golang.org/x/text` and `golang.org/x/sys` through the engine).

## 3. Design Decisions

- **A CLI and a skill**: the root cause is that no tool is installed, so the answer is to install one. Installed commands
  get found (on 11-07 Claude Code found the unregistered `report-studio-mcp` on PATH and used it). The skill teaches when
  to use it and what to do with a diagram it cannot draw. No MCP server (Discussion Log 4).
- **Go**: the engine is Go. No cgo.
- **Sources to carry over**: the CLI's body from mmdpng (read and copied; mermaid-render is not changed). Bundling the
  skill, building and verifying its zip, installing it into `~/.claude/skills`, and the skill validator in `make test`
  from report-studio-mcp (`report-manuscript`).
- **nlink-jp tools it complements**:
  - mermaid-render — the engine. Not changed; a need there becomes an issue on mermaid-render (organization ADR-025).
    mmdpng stays as the debugging tool.
  - report-studio-mcp — document typesetting; the skill hands off to it.
  - gem-agent and lagent — already draw on screen; unaffected.
- **Not in mcp-tactics**: not an MCP server, and it only produces an artifact on this machine (organization ADR-027 §1).
- **Out of scope**: an MCP server, drawing every diagram in a Markdown file, text-art diagrams, SVG, themes and colours,
  diagram types beyond the 7 (the engine's question), document typesetting.

## 4. Development Plan

### Phase 1: Core

- Scaffold in `_wip/mermaid-to-png` (CONVENTIONS.md templates).
- The CLI (§2).
- The skill (§2).
- Tests:
  - flags and input/output paths (file, standard input, `-o -`, refusing an output equal to the input)
  - exit status (each engine `ErrorKind` → 1, flag errors → 2, unreadable input or a bad font → 3)
  - the 7 types drawn by the real engine (10 keywords: flowchart, graph, flowchart-elk, sequenceDiagram, erDiagram,
    stateDiagram, stateDiagram-v2, gantt, pie, mindmap)
  - every flag and exit status the skill names exists in the CLI (organization ADR-027 §2); the organization's skill validator
  - `-version`

### Phase 2: Checking in the field

- Install a development build of the CLI and the skill, use them several times in real work (making a document with
  diagrams, for example), and see whether the work gets done without reaching for Chrome and the like. Installing the
  skill changes the user's environment: ask first, and put it back afterwards.
- Fix the skill's wording from what is seen.
- One independent verification pass.

### Phase 3: Release

- README.md / README.ja.md (using the CLI, installing the skill), CHANGELOG.md, AGENTS.md.
- The organization's release steps: create the repository, build for 4 platforms (darwin signed and notarized), each zip
  plus the separate skill zip (`mermaid-to-png-vX.Y.Z.zip`, the organization's skill zip name), `verify-release`,
  Homebrew formula, util-series submodule and catalogue row, organization profile, website card, knowledge feedback,
  check-org. The catalogue rows list the skill as part of the tool ("bundled mermaid-to-png skill", CONVENTIONS §Skill
  bundled with a tool).
- The README's Installation section leads the skill with the app's registration, keeps unzipping into
  `~/.claude/skills/` as the fallback, and describes `make install-skill` as the development loop (CONVENTIONS §Skill
  project scaffold).
- Watch the repository (`gh api -X PUT /repos/nlink-jp/mermaid-to-png/subscription -F subscribed=true`) and set its
  About (description, topics) (CONVENTIONS §Scaffold checklist). Name the skill zip in the release notes (its name is
  close to the source zip GitHub attaches).
- One independent verification pass before release.

**Independently reviewable points**: the RFP / end of Phase 1 / before release.

## 5. Required API Scopes / Permissions

None. It reads the input and fonts and writes the output PNG, nothing else.

## 6. Series Placement

Series: util-series
Reason: a pipe-friendly conversion CLI (the same "X-to-Y" as `csv-to-json` and `pptx-to-markdown`). lib-series holds the
engine and does not hold distributed tools.

## 7. External Platform Constraints

- **Platforms**: util-series' usual 4 (darwin/arm64, linux/amd64, linux/arm64, windows/amd64). Outside macOS there is no
  Hiragino, so without `-font` the exit status is 3 (stated in the README). The Homebrew tap is arm64 only.
- Outside API rate limits and the like: none.

---

## Discussion Log

1. **Start (2026-11-07)**: the user: "a CLI tool that converts mermaid to images would be good to have", then "Claude
   Code keeps trying odd ways to render mermaid (forcing it through Chrome), so this should be provided as a proper toolset."
2. **What the transcripts show** (tool calls in `~/.claude/projects/*/*.jsonl`):
   - 09-06, 09-12, 09-13: looked for and used a way to draw mermaid (node, npx, the mermaid CLI, headless Chrome). These
     predate the engine (first commit 09-28): there was no means at all. 09-12 and 09-13 are the same document, and
     Chrome also typeset the whole PDF.
   - 11-07: Claude Code found `report-studio-mcp` on PATH and finished 4 documents with diagrams. When the user asked to
     "finish it without report-studio (the page breaks are poor)", it tried fetching mermaid.js from a CDN, then loading
     chrome-pilot; after the user said "don't use Chrome" it drew the figures as SVG by hand in Python. A standalone tool
     that turns a diagram into an image would have been enough.
   - 10-30 (QLMarkdown) is work on an app that itself uses mermaid.js and is not counted.
3. **The root cause**: no command on PATH turns mermaid into an image (neither `mmdpng` nor `mmdc`). The engine is a
   library; `mmdpng` is a debugging tool sitting in `lib-series/mermaid-render/dist`. The cause is the 09-28 decision not
   to make a distributed CLI. I (Claude) first diagnosed "the tool exists but is not visible" and stacked an MCP server,
   discovery by name, steering by server instructions and user-scope registration on it (first draft). Two independent
   reviews found the 11-07 account misreported; the user then said the root cause had not been identified simply, and the
   diagnosis was changed to "it is not installed".
4. **Shape (the user's proposal)**: "mmdpng was set up as the library's debugging tool. Build the library's front end as
   a CLI and release it as an independent project together with a skill that guides its use." The MCP server came from
   the wrong diagnosis ("make it visible") and was dropped. The skill carries craft beyond the CLI's `-h` — when to use it
   and what to do with a diagram it cannot draw — so it meets organization ADR-027 §2's condition for a bundled skill.
5. **Reversing the earlier decision**: the mermaid-render RFP's "a distributed CLI (not worth the signing, notarization
   and versioning burden)" is not followed for this tool: that decision created the missing means, and Claude Code's
   fumbling is the price being paid. The mermaid-render documents are not edited (organization ADR-025); nothing is
   needed from mermaid-render, so no issue is opened.
6. **Review findings on the first draft that shaped the second**: a "do not use X" opens a third path (knowledge and
   gem-agent ADR-0063) → the skill is written in the positive. Exceeded limits and missing glyphs are `UnsupportedConstruct`
   in the engine and cannot be told apart → the exit status does not split them. The engine also pulls in `x/text` and
   `x/sys`. `MaxTextSize` counts UTF-16 units. Mermaid inside an HTML Artifact is a legitimate browser path. The rest
   (the work_dir response shape, file-name rules, structured-error spellings, cleaning up an MCP registration) went away
   with the MCP server.
7. **Platforms**: the first draft said darwin/arm64 only; a review pointed out that the precedents (report-studio-mcp,
   agent-report) are bound to macOS by structure and do not justify it. The CLI is pure Go and draws on other systems with
   `-font`, so it takes util-series' usual 4 platforms.
8. **Phase 2 check in the field (2026-11-07)**: a development build (624a2fd) was put in `~/.local/bin`, and the user
   registered the skill through the app. Two new sessions were started from this RFP's session with ordinary requests
   that named neither the tool nor the trial.
   - Scene 1 (draw gem-agent's MCP flow as a PNG): it loaded the skill first, drew with `mermaid-to-png`, looked at the
     image and redrew it with notes.
   - Scene 2 (the types' relations as a `classDiagram` PNG): it loaded the skill, tried a small class diagram first, and
     when it stopped at `unsupported diagram type` it offered choices (redraw as a flowchart, hand over the code block,
     an ER diagram) and asked. When the user chose the flowchart, it redrew it and made the PNG.
   - In neither scene did it reach for Chrome, npx, mmdc, a CDN, kroki, mermaid.ink or hand-drawn SVG (checked against
     the tool calls in the session transcripts). The skill's wording was not changed.
9. **Independent review of the implementation (2026-11-08, two passes)**: fixed — writing beside the output and
   renaming it in (which I added in Phase 1; §2 does not have it) replaced read-only files, widened modes to 0644,
   ignored the umask, broke links and broke `-o /dev/stdout` → the output is written in place like a shell's `>` (the
   PNG is complete in memory before the file is opened, so only the disk can fail midway). Flags after the file name
   are accepted. An output that is the file a redirected standard input reads is refused. A source over the read bound
   is reported in the engine's form. The lineless refusal form, an overstated test description and a reference to a
   test that did not exist were corrected in the documents and tests. Not changed — writing to a FIFO with no reader
   blocks (as a shell's `>` does), `-scale 1e-300` gives a 1×1 image (within the range), and `-o /dev/stdout` gets the
   path line after the PNG (the README says to use `-o -` for standard output).
