# CLAUDE.md — mermaid-to-png

Organization rules: https://github.com/nlink-jp/.github/blob/main/CONVENTIONS.md
Project summary, commands and gotchas: [AGENTS.md](AGENTS.md).

## Rules specific to this project

- **The engine is mermaid-render, and it is not changed from here.** What it
  draws, refuses and reports is the engine's (organization ADR-025). A need
  there is an issue on nlink-jp/mermaid-render. This tool only reads input,
  calls the engine, writes the PNG and reports.
- **No third-party dependencies** (organization ADR-024): the standard library
  and nlink-jp/mermaid-render only. That is why `cmd` uses `flag`, not cobra.
- **The design source is the RFP**: `docs/ja/mermaid-to-png-rfp.ja.md`
  (primary) and `docs/en/mermaid-to-png-rfp.md`. Its Discussion Log records
  why there is no MCP server.
- Always `make build` (output in `dist/`), never `go build` directly.
