# mermaid-to-png

[English](README.md)

mermaid の図 1 つを PNG 1 つにします。手元で動き、ブラウザを使いません。
`mermaid-to-png` は、組織自前の mermaid エンジン
[mermaid-render](https://github.com/nlink-jp/mermaid-render) のコマンドライン版です。
エージェントに「いつ使うか」と「図が拒まれたらどうするか」を教えるスキルが付いています。

## なぜ

mermaid を描くコマンドが無いと、図の画像を頼まれたエージェントはその場で手段を組み立てます。
ヘッドレス Chrome、CDN の mermaid.js、npx 経由の mmdc、あるいは図の手描きです。
mermaid-render は手元で mermaid を描けますが、ライブラリです。これはそのコマンドです。

## 入れ方

コマンドとスキルは別々に入れます。Homebrew とバイナリのアーカイブにはコマンドだけが入っています。
スキルはリリースごとに別の添付物 `mermaid-to-png-vX.Y.Z.zip`（名前にプラットフォームが付かないもの）として付き、
zip の根にスキルのフォルダがあります。

### コマンド

```bash
brew install nlink-jp/tap/mermaid-to-png
```

ソースから入れる場合:

```bash
git clone https://github.com/nlink-jp/mermaid-to-png.git
cd mermaid-to-png
make build          # → dist/mermaid-to-png
```

macOS（arm64、署名・公証済み）、Linux（amd64・arm64）、Windows（amd64）のアーカイブは
[リリースのページ](https://github.com/nlink-jp/mermaid-to-png/releases)にあります。
文字はヒラギノ角ゴシックで描きます。ヒラギノは macOS にしか無いので、他の OS では `-font` でフォントを指定してください。

### スキル

使っているコマンドと同じリリースの `mermaid-to-png-vX.Y.Z.zip` を取ってください。

- **Claude（デスクトップアプリ・claude.ai）** — カスタマイズ → スキルで zip をアップロードします。
  更新のたびに新しい zip をアップロードし直してください。
- **Claude Code に手で入れる場合** — `unzip mermaid-to-png-vX.Y.Z.zip -d ~/.claude/skills/`

スキルは、エージェントが動く場所にコマンドがあることを前提にしています。
クローンからの `make install-skill` は開発用の手順です。zip を作り、それを `~/.claude/skills/` に展開します
（別の場所は `SKILLS_DEST=<dir>`）。

## 使い方

```bash
mermaid-to-png diagram.mmd                 # → diagram.png。書いたパスを印字する
mermaid-to-png -o figure.png diagram.mmd
cat diagram.mmd | mermaid-to-png -o figure.png
mermaid-to-png -o - diagram.mmd > figure.png
mermaid-to-png -scale 1 diagram.mmd        # 小さめの画像
mermaid-to-png -font /path/to/font.ttc -font-name "Face Name" diagram.mmd
```

| フラグ | 意味 |
|---|---|
| `-o path` | 出力の PNG。`-` で標準出力。既定は入力の拡張子を `.png` にしたもの。標準入力から読むときは必須 |
| `-scale n` | 拡大率。(0, 8] の範囲、既定 2（本文の文字が 28 px） |
| `-font path` | 文字のフォントファイル（既定はヒラギノ角ゴシック W3） |
| `-font-name name` | `-font` の中のフェイス。フルネームか PostScript 名 |
| `-bold path`・`-bold-name name` | 太字のフェイス（既定は本文と同じ） |
| `-version` | 版数を印字する |

入力は図のソースだけです（```` ```mermaid ```` の囲みは付けません）。出力が入力そのものになる指定は拒みます。
通常のファイルへは、出力先の隣に書いてから名前を付け替えるので、失敗しても書きかけのファイルは残りません。

### 終了コード

| コード | 意味 |
|---|---|
| 0 | PNG を書いた |
| 1 | エンジンが図を拒んだ — `mermaid-to-png: <入力>:<行>: <種類>: <説明>` |
| 2 | フラグや引数の誤り（続けて使い方の行を出す） |
| 3 | 入力・出力・フォントを読めない、書けない、読み込めない |

`<種類>` はエンジンのもので、`syntax error`・`unsupported diagram type`・
`unsupported construct`（上限超えと、どのフォントにも無い文字もここ）・`layout fault` のどれかです。

## 描けるもの

mermaid-render が描く 7 種類です: flowchart（`flowchart`・`graph`）、シーケンス図、ER 図、状態遷移図、ガントチャート、
円グラフ、マインドマップ。白地のカードに描きます。見た目の指定（`style`・`classDef`・テーマ）は読み捨てます。
エンジンが何をなぜ拒むかは、エンジンの [README](https://github.com/nlink-jp/mermaid-render#readme) にあります。

## スキル（同梱）: mermaid-to-png

このリポジトリには **mermaid-to-png** スキルが付いています。中身は次のとおりです。
このコマンドで図を描く場面、終了コードごとの対処、エンジンが構文を拒んだときの書き換え方
（ラベルの中の HTML、入れ子の subgraph、フォントに無い文字）、そして report-studio やページ自身の mermaid.js に
任せる場面。入れ方は[入れ方](#スキル)にあります。

`make test` は組織のスキル検証も走らせます。`cmd/skill_test.go` はスキルをコマンドに縛ります。スキルが名指しする
フラグ・終了コード・エンジンの種類・図の種類はすべて実在しなければならず、スキルが勧める書き換えは、スキルのとおりに
拒まれ、受け入れられなければなりません。

## ビルド

```bash
make build      # dist/mermaid-to-png
make test       # 単体テスト＋スキルの検証
make package    # 全プラットフォーム、darwin は公証、スキルの zip も
```

## 設計

設計と経緯は RFP にあります:
[docs/ja/mermaid-to-png-rfp.ja.md](docs/ja/mermaid-to-png-rfp.ja.md)
（[English](docs/en/mermaid-to-png-rfp.md)）。

## ライセンス

MIT — [LICENSE](LICENSE) を参照。
