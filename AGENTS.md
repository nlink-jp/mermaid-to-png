# AGENTS.md — mermaid-to-png

## Summary

`mermaid-to-png` turns one mermaid diagram into one PNG, locally and with no
browser. It is a command-line front end to the organization's engine,
[mermaid-render](https://github.com/nlink-jp/mermaid-render) (lib-series),
and bundles the `mermaid-to-png` skill (organization ADR-027 §2).
Series: util-series. Module: `github.com/nlink-jp/mermaid-to-png`.
Only dependency: `github.com/nlink-jp/mermaid-render` (x/image, x/text and
x/sys come through it).

Design source: `docs/ja/mermaid-to-png-rfp.ja.md` (primary) and
`docs/en/mermaid-to-png-rfp.md`. Its Discussion Log records why there is no
MCP server.

## Commands

```bash
make build           # dist/mermaid-to-png (signed when a Developer ID is in the keychain)
make test            # go test ./... + tests/validate-skill.sh
make check           # vet + test + build
make package         # 4 platforms, archives, darwin notarized, skill zip
make verify-release  # also runs verify-skill-zip
make install-skill   # unpack the built skill zip into ~/.claude/skills
```

## Structure

```
mermaid-to-png/
├── main.go                 # calls cmd.Execute()
├── cmd/
│   ├── root.go             # Execute, Run, Version
│   ├── run.go              # flags (NewFlagSet), exit statuses, read → engine → write
│   ├── run_test.go         # CLI behaviour against the real engine
│   └── skill_test.go       # holds the skill to the flags, statuses and engine
├── mermaid-to-png/SKILL.md # the bundled skill (name = directory)
├── tests/validate-skill.sh # org skill validator, vendored byte-identical
├── scripts/                # org templates: codesign, notarize, Homebrew (vendored verbatim)
└── docs/{ja,en}/           # RFP
```

## Gotchas

- Go's `flag` package reads `-version` and `--version` as the same flag; the
  Homebrew formula's test runs `--version`.
- The input is read with a bound of 3 × `MaxTextSize` bytes: a UTF-16 unit is
  at most 3 bytes of UTF-8, so a longer input is certainly over the engine's
  limit (exit 1 without calling it). The engine judges anything shorter.
- A regular-file output is written beside the destination and renamed in. A
  destination that exists and is not a regular file (`/dev/null`, a pipe) is
  written in place — a rename cannot replace it.
- There is no terminal detection: the standard library cannot tell a tty from
  `/dev/null` (both are character devices). Reading a terminal waits, as `cat`
  does.
- Drawing tests skip where Hiragino is absent (`needFont`); off macOS run them
  with a font or expect skips.
- `TestSkillFlagsExist` drops spans fenced by four backticks before pairing
  single backticks; the skill shows the ```` ```mermaid ```` fence that way.
- `install-skill` / `uninstall-skill` remove `$(SKILLS_DEST)/mermaid-to-png`
  before unpacking; `SKILLS_DEST` must be set to a real directory.
