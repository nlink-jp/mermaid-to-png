# Changelog

All notable changes to this project are documented here. The format follows
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and versions follow
[Semantic Versioning](https://semver.org/).

## [Unreleased]

### Changed

- `make package` runs `govulncheck ./...` first (`make vuln`) and stops on a
  finding, so a release cannot be built past one.
- `make vuln` — and so `make package` — scans every platform in `PLATFORMS`,
  with the environment `build-all` builds each with. It scanned the host's
  build alone, so an advisory in a Linux- or Windows-only package could not
  stop a release.

## [0.1.0] - 2026-11-08

### Added

- First release: `mermaid-to-png`, a command-line front end to
  [mermaid-render](https://github.com/nlink-jp/mermaid-render) that turns one
  mermaid diagram into one PNG with no browser. Reads a file or standard
  input, writes a file or standard output (`-o -`), takes flags before or
  after the file name, and exits 0 (written), 1 (the engine refused the
  diagram, as `<input>:<line>: <kind>: <message>`, or `<input>: <kind>:
  <message>` when no one line is at fault), 2 (usage) or 3 (input, output or
  font).
- The bundled `mermaid-to-png` skill: when to draw with this command, how to
  act on each exit status and engine refusal, and when the work belongs to
  report-studio. Released as its own asset, `mermaid-to-png-vX.Y.Z.zip`.
