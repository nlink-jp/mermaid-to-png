---
name: mermaid-to-png
description: Turn a mermaid diagram into a PNG image file on this machine with the mermaid-to-png command — to put a diagram into a document, slides, a note, a Word or PDF file, or to look at a diagram you just wrote. Covers flowchart, sequence, ER, state, gantt, pie and mindmap, choosing the size, and what to change when a diagram is refused. Use whenever mermaid has to become an image. Also for mermaid を画像にする・図を PNG にする・mermaid のレンダリング・構成図・フローチャート・シーケンス図・ER 図・状態遷移図・ガントチャート・円グラフ・マインドマップ・図を資料やスライドに入れる. Needs the mermaid-to-png command.
argument-hint: "<a .mmd file, or the diagram to draw>"
---

# mermaid-to-png

`mermaid-to-png` draws one mermaid diagram as one PNG, on this machine, with
the organization's own engine (mermaid-render). `mermaid-to-png -h` lists the
flags; this skill is when to use it and how to act on what it reports.

## The run

1. **Write the diagram to a file** in your working directory, such as
   `diagram.mmd` (the source only, without the ```` ```mermaid ```` fence).
2. **Draw it**: `mermaid-to-png diagram.mmd` writes `diagram.png` beside it
   and prints the path. `-o name.png` names the output (before or after the
   file name); a PNG already at that path is replaced, so give `-o` another
   name to keep one. The source can also come on standard input with `-o`,
   and `-o -` writes the PNG to standard output.
3. **Look at the PNG** by reading it before you hand it on or put it into a
   document. Read the labels: the image is the diagram the reader gets.
4. **Act on the exit status**:

| Status | Meaning | Next |
|---|---|---|
| 0 | written | use the path it printed |
| 1 | the engine refused the diagram | the message reads `<input>:<line>: <kind>: <message>` (`<input>: <kind>: <message>` when no one line is at fault, as with a limit); see below |
| 2 | a flag or argument | the usage line follows the message |
| 3 | the input, the output or the font | the message names which |

## Status 1: change the diagram, or ask

The engine draws 7 types: `flowchart` (also `graph`), `sequenceDiagram`,
`erDiagram`, `stateDiagram` (also `stateDiagram-v2`), `gantt`, `pie` and
`mindmap`. The kind in the message says what to do:

- **`syntax error`** — the line is not valid mermaid. Fix that line and draw
  again. A first line the engine does not know at all lands here too (a typo,
  or a diagram type newer than the engine).
- **`unsupported diagram type`** — a valid mermaid type outside the 7, such as
  `classDiagram` or `block-beta`. Tell the user, give them the source, and
  offer the choices: keep it as a mermaid code block, or say the same thing
  as one of the 7 types (an ER diagram or a flowchart often carries what a
  class diagram does). Ask before changing the type.
- **`unsupported construct`** — rewrite the line it names and draw again.
  The ones that come up:
  - *HTML in a label* (`<b>`, `<code>`, …): write the label as plain text.
    `<br>` line breaks are drawn.
  - *Markdown string* or *markdown formatting* (backquoted Markdown in a
    flowchart label, `**bold**` in a mindmap): write the label as plain text.
  - *nested subgraph*: keep one level of subgraphs; split the inner group into
    a subgraph of its own beside the others.
  - *no font can draw* a character: replace it with one the font has
    (`➔` → `→`, an emoji → a word).
  - *a limit* (the message names it, such as too many nodes or too many
    pixels): split the diagram into two, or, when it is the image size,
    draw with a lower `-scale`.
- **`layout fault`** — the diagram is valid; the engine laid it out wrong.
  Give the user the source and the message. It is a bug report for
  mermaid-render, and rearranging the diagram (another direction, fewer
  crossing links) may get past it.

Styling is the engine's to decide: `style`, `classDef` and themes in
`%%{init}%%` are read and set aside, and the card is always white.

## Size and fonts

- `-scale` defaults to 2: body text is 28 px, sharp on screens and in
  documents. Use 1 for a small preview, and more only for a large print.
- Text is drawn in Hiragino Sans, macOS's own Japanese font. For another face,
  or on a system without Hiragino (status 3 says the default font cannot be
  loaded), give a font file with `-font` (`-font-name` picks a face inside a
  collection; `-bold` and `-bold-name` set the bold text).

## Hand-offs

- **A whole report with diagrams, as PDF and Word**: report-studio typesets
  the Markdown and draws its mermaid blocks itself. Draw with this command
  when only the image is wanted, or when the document is made some other way
  (pandoc, a slide tool, Google Docs, Obsidian).
- **Mermaid inside an HTML page or an Artifact** is drawn by that page's own
  mermaid.js; it needs no image.
