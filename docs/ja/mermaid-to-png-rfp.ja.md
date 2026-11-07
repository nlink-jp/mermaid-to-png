# RFP: mermaid-to-png

> 作成: 2026-11-07（第 2 稿）
> 状態: 草案

## 1. 問題定義

この機械には、mermaid の図を画像にするコマンドが入っていない。自前の描画エンジン mermaid-render（lib-series）は
ライブラリであってコマンドではなく、唯一のコマンド `mmdpng` はライブラリのデバッグ用で、配布されていない。
mermaid-render の RFP（2026-09-28）が「配布用の CLI は作らない」と決めたためである。道具が無いので、Claude Code は
mermaid を画像にしたい場面（資料・スライド・ノートに図を入れる、書いた図を確かめる）のたびに、ヘッドレス Chrome・
CDN の mermaid.js・npx・mmdc を探したり、図を手で描いたりして、その場で手段を組み立てる。
mermaid-to-png は、エンジンのフロントエンドとなる CLI と、使い方を導くスキルを、1 つのプロジェクトとしてリリースする。
CLI を brew で入れれば手段が PATH にあり、スキルが「mermaid を画像にする場面ではこれを使う」ことと、描けない図への
対処を教える。

**対象にする利用者**: Claude Code（スキルと CLI）。人がシェルやスクリプトから直接使ってもよい。

**範囲の切り分け**: 図の入った Word・PDF を丸ごと作るのは report-studio-mcp の領分。mermaid-to-png は図 1 枚を
画像ファイル 1 つにするだけ。

## 2. 機能仕様

### コマンド / API の面

```
mermaid-to-png [-o out.png] [-scale 2] [-font path [-font-name name]] [-bold path [-bold-name name]] [file.mmd | -]
mermaid-to-png -version
```

- フラグは標準ライブラリの `flag` で読む（1 本ダッシュ。mmdpng と同じ）。
- 入力はファイル、または標準入力（`-`、またはファイル名の省略）。
- 出力の既定は、入力ファイル名の拡張子を `.png` にしたもの。出力が入力と同じパスになるときは拒む。
  標準入力から読むときは `-o` が必須。`-o -` で PNG を標準出力に書く（パイプ用）。
- `-scale` は (0, 8]、既定 2（エンジンの範囲と既定。範囲外はエンジンに渡す前に拒む）。
- フォントのフラグはエンジンの `raster.FontSpec` にそのまま渡す。省略時はヒラギノ角ゴシック W3/W6。
  `-font-name`・`-bold`・`-bold-name` は `-font` が無ければ拒む（mmdpng と同じ）。
- 成功時は書いたファイルのパスを 1 行だけ標準出力に印字する（`-o -` のときは印字しない）。

### 入力 / 出力

- 入力: mermaid のテキスト。上限はエンジンの `MaxTextSize`（UTF-16 の単位で 50,000）。
- 出力: PNG（白地のカード。エンジンの描画そのまま）。
- 失敗の印字（標準エラー）: `mermaid-to-png: <入力>:<行>: <種類>: <説明>`。`<種類>` はエンジンの `ErrorKind` の名前
  （`syntax error`・`unsupported diagram type`・`unsupported construct`・`layout fault`）。行が分からないときは行を省く。
- 終了コード:

| コード | 意味 |
|---|---|
| 0 | 書けた |
| 1 | 図を描けなかった（エンジンが `*mermaidrender.Error` を返した。上限超えと描けない文字もここに入る） |
| 2 | 使い方の誤り（フラグ・引数） |
| 3 | 入出力・フォントの失敗（読めない、書けない、フォントが読めない） |

細部（標準入力のときの `<入力>` の表記、入力が端末のときの扱いなど）は第 1 段階で決め、README とテストに書く。

### スキル

リポジトリの直下に `mermaid-to-png/SKILL.md`（skills-series の形、組織 ADR-027 §2）。中身は CLI の `-h` に書けない
コツだけにする:

- **使う場面**: mermaid の図を画像ファイルにしたいとき（資料・スライド・ノートに入れる、書いた図を目で確かめる）。
  画像ができたら、その PNG を読めば図を見られる。
- **描ける図**: エンジンの 7 種類（flowchart・sequence・ER・state・gantt・pie・mindmap）。
- **終了コード 1 のとき**: 描けない構文を書き換える（実際に要ったもの: ラベルの `<b>`・`<code>` を外す、入れ子の
  subgraph を平らにする、フォントに無い文字を置き換える）か、描けない種類ならソースのまま利用者に伝え、どうするかを尋ねる。
- **受け渡し**: 図の入った文書を丸ごと作るなら report-studio-mcp。HTML の Artifact の中の mermaid はそのページの
  mermaid.js が描くので、このツールの出番ではない。
- 説明文は肯定の形で書く（「〜を使うな」は書かない。knowledge `llm-integration.md`「「〜をするな」は過剰一般化する」）。
  description は 1,024 字以内。

### 設定

なし。

### 外部への依存

なし（ネットワーク・外部サービス・資格情報を使わない）。Go の依存は `github.com/nlink-jp/mermaid-render` だけ
（`golang.org/x/image`・`golang.org/x/text`・`golang.org/x/sys` はエンジン経由）。

## 3. 設計判断

- **CLI とスキル**: 真因は「道具が入っていない」なので、答えは「入れる」こと。入っているコマンドは見つけてもらえる
  （11-07 に Claude Code は未登録の `report-studio-mcp` を PATH から見つけて使った）。使う場面と描けない図への対処は
  スキルが教える。MCP サーバは作らない（Discussion Log 4）。
- **Go**: エンジンが Go。cgo は使わない。
- **移植元**: CLI の中身は mmdpng（読んで写す。mermaid-render は変えない）。スキルの同梱・zip の作成と検証・
  `~/.claude/skills` への導入・スキル検証の `make test` は report-studio-mcp（`report-manuscript`）。
- **補い合う nlink-jp のツール**:
  - mermaid-render — エンジン。手を入れない。直したいことが出たら mermaid-render に issue を立てる（組織 ADR-025）。
    mmdpng はデバッグ用としてそのまま残る。
  - report-studio-mcp — 文書の組版。スキルから受け渡す。
  - gem-agent・lagent — 画面内の描画はすでにあり、影響しない。
- **mcp-tactics には載せない**: MCP サーバではなく、手元で成果物を作るだけ（組織 ADR-027 §1）。
- **範囲の外**: MCP サーバ、Markdown の中の図をまとめて描く機能、罫線の図、SVG、テーマ・色の指定、7 種類以外の図
  （増やすかはエンジンの領分）、文書の組版。

## 4. 開発計画

### 第 1 段階: 中身

- `_wip/mermaid-to-png` に雛形（CONVENTIONS.md のテンプレート）。
- CLI（§2）。
- スキル（§2）。
- テスト:
  - フラグ・入出力の経路（ファイル・標準入力・`-o -`・入力と同じ出力の拒否）
  - 終了コード（エンジンの各 `ErrorKind` → 1、フラグの誤り → 2、読めない入力・フォントの誤り → 3）
  - 実際のエンジンで 7 種類（キーワード 10 個: flowchart・graph・flowchart-elk・sequenceDiagram・erDiagram・
    stateDiagram・stateDiagram-v2・gantt・pie・mindmap）を描く
  - スキルが名指しするフラグと終了コードが CLI に実在する（組織 ADR-027 §2）、組織のスキル検証
  - `-version`

### 第 2 段階: 実地での確認

- 開発版の CLI とスキルを入れ、実際の作業（図を入れた資料づくりなど）で何度か使い、Chrome などに手を伸ばさずに
  使えるかを見る。スキルを入れるのは利用者の環境を変えることなので、事前に確認し、終わったら元に戻す。
- 気づいたことでスキルの文面を直す。
- 独立検証を 1 本。

### 第 3 段階: リリース

- README.md / README.ja.md（CLI の使い方、スキルの入れ方）、CHANGELOG.md、AGENTS.md。
- 組織の手順どおり: リポジトリの作成、4 プラットフォームのビルド（darwin の署名・公証）、各 zip と別添のスキル zip
  （`mermaid-to-png-vX.Y.Z.zip`、組織のスキル zip の名前）、`verify-release`、Homebrew の formula、util-series への
  submodule とカタログ行、組織のプロフィール、Web サイトのカード、knowledge への還元、check-org。カタログの行には、
  スキルをツールの一部として載せる（「bundled mermaid-to-png skill」、CONVENTIONS §Skill bundled with a tool）。
- README の入れ方の節は、スキルについてアプリでの登録を先に、`~/.claude/skills/` への展開を代わりの手段として書き、
  `make install-skill` は開発用の手順として書く（CONVENTIONS §Skill project scaffold）。
- リリース前に独立検証を 1 本。

**単独でレビューできる区切り**: RFP／第 1 段階の終わり／リリース前。

## 5. 必要な API スコープ / 権限

なし。読むのは入力とフォント、書くのは出力の PNG だけ。

## 6. シリーズの置き場所

シリーズ: util-series
理由: パイプで使える変換の CLI（`csv-to-json`・`pptx-to-markdown` と同じ「X-to-Y」）。lib-series はエンジンの置き場所で、
配布用のツールは置かない。

## 7. 外部プラットフォームの制約

- **プラットフォーム**: util-series の通常どおり 4 つ（darwin/arm64・linux/amd64・linux/arm64・windows/amd64）。
  macOS 以外にはヒラギノが無いので、`-font` を付けないと終了コード 3 になる（README に書く）。Homebrew の tap は arm64 だけ。
- 外部 API のレート制限など: なし。

---

## Discussion Log

1. **発端（2026-11-07）**: 利用者「mermaid 記述を画像に変換する CLI ツールがあると良さそう」。続けて「Claude Code が
   毎回、おかしな方法で mermaid のレンダリングを試みる（Chrome を無理やり経由させようとする）ので、ちゃんとした
   ツールセットとして用意すべき」。
2. **記録で確かめたこと**（`~/.claude/projects/*/*.jsonl` のツール呼び出し）:
   - 09-06・09-12・09-13: mermaid を描く手段（node・npx・mermaid CLI・ヘッドレス Chrome）を探し、使った。
     エンジン（最初のコミット 09-28）ができる前で、手段がそもそも無かった。09-12 と 09-13 は同じ資料で、Chrome は
     PDF 全体の組版にも使っていた。
   - 11-07: Claude Code は report-studio-mcp を PATH から見つけて 4 文書を図入りで仕上げた。利用者が「report-studio を
     使わずに仕上げて（改ページが良くない）」と頼むと、CDN の mermaid.js の取得 → chrome-pilot の読み込みを試み、
     利用者が「chrome 使わないでください」と止めた後、Python で図を SVG として手で描いた。
     単体で図を画像にする道具があれば足りた場面。
   - 10-30（QLMarkdown）はアプリ自体が mermaid.js を使う作業なので数えない。
3. **真因の特定**: PATH に mermaid を画像にするコマンドは無い（`mmdpng`・`mmdc` とも無し）。エンジンはライブラリで、
   `mmdpng` は `lib-series/mermaid-render/dist` にあるデバッグ用。原因は 09-28 の「配布用の CLI は作らない」という
   判断そのもの。当初、私（Claude）は「道具はあるのに見えていない」と見立て、MCP サーバ・名前での発見・説明文での誘導・
   ユーザー単位の登録を積み上げた（第 1 稿）。独立検証 2 本で 11-07 の経緯の誤記が見つかり、その後、利用者に
   「真因をシンプルに特定できていない」と指摘されて見立てを「入っていない」に改めた。
4. **形（利用者の提案）**: 「mmdpng はライブラリのデバッグ用。ライブラリのフロントエンドを CLI で構成し、利用誘導用の
   スキルとセットで独立したプロジェクトとしてリリースすればよい」。MCP サーバは「見えるようにするため」という誤った
   見立てから出たものなので外した。スキルは、使う場面と描けない図への対処という、CLI の `-h` を超えるコツを持つので、
   組織 ADR-027 §2 の同梱スキルの条件を満たす。
5. **前の判断を改める**: mermaid-render の RFP の「配布用の CLI（署名・公証・版管理の負担に見合わない）」は、
   その判断が手段の欠如を生み、Claude Code の手探りという代価を払っていることが分かったので、このツールでは採らない。
   mermaid-render の文書は書き換えない（組織 ADR-025）。mermaid-render に頼むことは無いので issue も立てない。
6. **第 1 稿の検証で出た指摘のうち、第 2 稿に効いたもの**: 「〜を使うな」と書くと第 3 の経路を生む（knowledge と
   gem-agent ADR-0063）→ スキルは肯定の形で書く。上限超えと描けない文字はエンジンでは `UnsupportedConstruct` で、
   区別できない → 終了コードは分けない。エンジンの依存は `x/image` のほかに `x/text`・`x/sys`。`MaxTextSize` の単位は
   UTF-16。HTML の Artifact の中の mermaid はブラウザが描く正規の道。残りの指摘（work_dir の応答の形、ファイル名の
   規則、構造化エラーの綴り、MCP の登録の後片づけ）は MCP サーバと一緒に消えた。
7. **プラットフォーム**: 第 1 稿は darwin/arm64 だけとしたが、比べる相手（report-studio-mcp・agent-report）は構造的に
   macOS に縛られていて理由にならないと指摘された。CLI は純粋な Go で `-font` があれば他の OS でも描けるので、
   util-series の通常の 4 プラットフォームにする。
