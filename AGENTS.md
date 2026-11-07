# AGENTS.md — mermaid-to-png

## Summary

`mermaid-to-png` turns one mermaid diagram into one PNG, locally and with no
browser. It is a command-line front end to the organization's engine,
[mermaid-render](https://github.com/nlink-jp/mermaid-render) (lib-series).
Series: util-series. Module: `github.com/nlink-jp/mermaid-to-png`.

Design source: `docs/ja/mermaid-to-png-rfp.ja.md` (primary) and
`docs/en/mermaid-to-png-rfp.md`.

## Commands

```bash
make build     # dist/mermaid-to-png (signed when a Developer ID is in the keychain)
make test      # go test ./...
make check     # vet + test + build
make package   # 4 platforms, archives, darwin notarized
make verify-release
```

## Structure

```
mermaid-to-png/
├── main.go            # calls cmd.Execute()
├── cmd/               # the command line (stdlib flag)
├── scripts/           # org templates: codesign, notarize, Homebrew (vendored verbatim)
└── docs/{ja,en}/      # RFP
```

## Gotchas

- Go's `flag` package reads `-version` and `--version` as the same flag; the
  Homebrew formula's test runs `--version`.
