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
make build            # dist/mermaid-to-png (signed when a Developer ID is in the keychain)
make test             # go test ./... + tests/validate-skill.sh
make vet              # go vet ./...
make check            # vet + test + build
make package          # 4 platforms, archives, darwin notarized, skill zip
make verify-release   # also runs verify-skill-zip
make brew             # after package: generate the formula into the local homebrew-tap checkout
make brew-print       # render the formula to stdout only
make install-skill    # unpack the built skill zip into ~/.claude/skills
make uninstall-skill
make clean
```

## Structure

```
mermaid-to-png/
├── main.go                 # calls cmd.Execute()
├── cmd/
│   ├── root.go             # Execute, Run, Version
│   ├── run.go              # flags (NewFlagSet, parseArgs), exit statuses, read → engine → write
│   ├── root_test.go        # --version and -version agree
│   ├── run_test.go         # CLI behaviour against the real engine
│   ├── pin_test.go         # behaviour the release review found unguarded
│   └── skill_test.go       # holds the skill to the flags, statuses and engine
├── mermaid-to-png/SKILL.md # the bundled skill (name = directory)
├── tests/validate-skill.sh # org skill validator, vendored byte-identical
├── scripts/                # org templates: codesign, notarize, Homebrew (vendored verbatim)
└── docs/{ja,en}/           # RFP
```

## Gotchas

- Go's `flag` package reads `-version` and `--version` as the same flag; the
  Homebrew formula's test runs `--version`.
- `flag` stops at the first non-flag, so `parseArgs` resumes after each file
  name: `file -o out` works. Everything after `--` is a file name.
- The input is read with a bound of 3 × `MaxTextSize` bytes: a UTF-16 unit is
  at most 3 bytes of UTF-8, so a longer input is certainly over the engine's
  limit and is refused in the engine's form without calling it. The engine
  judges anything shorter.
- The output is opened and written in place (`O_CREATE|O_TRUNC`, 0666 less
  the umask), like a shell's `>`. Writing beside it and renaming was tried
  and removed: it replaced read-only files, widened modes and broke links.
- An output that is the input (by path or `SameFile`) or the regular file a
  redirected standard input reads is a usage error. A pipe cannot be told
  apart and is not checked; a device on stdin (/dev/null) is not compared.
- An output that is the file standard output writes to (`-o /dev/stdout`) is
  treated as `-o -`: written through the stream, no path line, no reopen
  (which on Linux could truncate a file stdout appends to).
- Accepted: a failure while writing leaves the output truncated (as `>`
  does; exit 3 says so), and `-o --` takes `--` as the separator, not the
  value (a usage error, never a wrong write).
- There is no terminal detection: the standard library cannot tell a tty from
  `/dev/null` (both are character devices). Reading a terminal waits, as `cat`
  does.
- `defaultFont` is a variable so `TestDefaultFontMissing` can stand in for a
  system without Hiragino. Drawing tests skip where Hiragino is absent
  (`needFont`).
- `TestSkillFlagsExist` drops spans fenced by four backticks before pairing
  single backticks; the skill shows the ```` ```mermaid ```` fence that way.
- `install-skill` / `uninstall-skill` remove `$(SKILLS_DEST)/mermaid-to-png`
  before unpacking; `SKILLS_DEST` must be set to a real directory.
