---
name: esa-mini
description: >
  esa.io の記事を取得・作成・更新する最小限の CLI ツール。
  記事本文をファイルに保存し、コンテキストウィンドウの消費を抑える。
  使うタイミング: (1) esa の記事を読みたい・取得したい,
  (2) esa の記事を検索・一括ダウンロードしたい,
  (3) esa に記事を投稿・更新したい,
  (4) esa のチーム一覧を確認したい,
  (5) esa のカテゴリやタグの一覧を確認したい。
---

# esa-mini

## セットアップ

`esa-mini` が PATH に存在しない場合、先にインストールする。

```bash
go install github.com/zenbo/esa-mini@latest
```

`esa-mini --help` を実行して利用可能なコマンドとオプションを確認する。

注意: esa API は1ユーザーあたり15分間に75リクエストのレートリミットがある。大量取得時は --limit で件数を制限すること。
`esa-mini token set` でトークンを保存するか、環境変数 ESA_ACCESS_TOKEN を設定する。

## コマンド概要

- `esa-mini token set` — アクセストークンを `~/.config/esa-mini/token` に保存
- `esa-mini token show` — 保存済みトークンを確認（マスク表示）
- `esa-mini token delete` — 保存済みトークンを削除
- `esa-mini teams` — 所属チーム一覧を表示
- `esa-mini categories <team> [flags]` — カテゴリ一覧を表示（`--match` で部分一致検索、`--prefix` で前方一致、`--top` でトップレベルのみ）
- `esa-mini tags <team> [flags]` — タグ一覧を表示（`--match` で部分一致フィルタ）
- `esa-mini search <team> [flags]` — 記事を検索し一覧表示（`--output` で一括保存）
- `esa-mini get <team> <number> --output <path>` — 記事を frontmatter 付き Markdown として保存（ディレクトリ指定時は `{number}.md` で自動命名）
- `esa-mini create [team] --file <path>` — ファイルから新規記事を作成（team 省略時は frontmatter から取得）
- `esa-mini update [team] [number] --file <path>` — ファイルから既存記事を更新（team / number 省略時は frontmatter から取得）

## 記事ファイルのフォーマット

create / update に渡す Markdown ファイルは以下の形式で作成する。

```markdown
---
team: myteam
number: 123
title: 記事タイトル
category: dev/tips
tags:
  - go
  - cli
wip: true
---

本文をここに書く
```

`get` で取得したファイルにはすべてのフィールドが含まれる。`create` / `update` では frontmatter の `team` / `number` を使うため CLI 引数での指定は省略可能。

## 推奨ワークフロー

### 記事を探すとき

カテゴリ名やタグ名が不明な場合は、search の前に categories / tags で正確な名前を確認する。

```bash
# 1. カテゴリを探す
esa-mini categories myteam --match "設計"
# → dev/設計ドキュメント  (12 posts)

# 2. 見つけたカテゴリで検索
esa-mini search myteam --category "dev/設計ドキュメント" -o ./docs/
```

### 記事を作成するとき

既存のカテゴリ・タグと表記を揃えるために、create の前に確認する。

```bash
# 1. 既存カテゴリを確認
esa-mini categories myteam --prefix "dev/"

# 2. 既存タグを確認
esa-mini tags myteam --match "go"
# → go  (45 posts),  golang  (3 posts)
# → "go" が正式なタグ名だとわかる

# 3. 正しいカテゴリ・タグで記事を作成
esa-mini create myteam --file ./new-post.md
```

## categories コマンド

カテゴリ一覧を表示する。search の --category や create の frontmatter に指定する正確なカテゴリ名を確認するために使う。

```bash
# 全カテゴリパス一覧
esa-mini categories myteam

# トップレベルのみ
esa-mini categories myteam --top

# 前方一致で絞り込み
esa-mini categories myteam --prefix "dev/"

# 部分一致で検索
esa-mini categories myteam --match "設計"
```

### categories フラグ

| フラグ | 説明 |
|---|---|
| `--top` | トップレベルカテゴリのみ表示 |
| `--prefix` | 前方一致フィルタ |
| `--match` | 部分一致フィルタ |

`--top` と `--prefix`/`--match` は同時に使えない。

## tags コマンド

タグ一覧を記事数付きで表示する。search の --tag や create の frontmatter に指定する正確なタグ名を確認するために使う。

```bash
# タグ一覧（記事数の多い順）
esa-mini tags myteam

# 部分一致でフィルタ
esa-mini tags myteam --match "設計"
```

### tags フラグ

| フラグ | 説明 |
|---|---|
| `--match` | 部分一致フィルタ（大文字小文字区別なし） |

## search コマンド

記事を検索し、結果一覧を表示する。`--output` を指定すると検索結果をファイルに一括保存する。

```bash
# カテゴリで検索して一覧表示
esa-mini search myteam --category "dev/tips"

# 自分がウォッチしている記事を一括ダウンロード
esa-mini search myteam --watched-by --output ./docs/

# 自分が作成した記事（--author 値なし = 自分）
esa-mini search myteam --author

# 複合条件
esa-mini search myteam --author --category "日報" --wip false

# 件数制限
esa-mini search myteam --watched-by --limit 50 --output ./docs/
```

### search フラグ

| フラグ | 短縮 | 説明 |
|---|---|---|
| `--query` | `-q` | 生の検索クエリ |
| `--author` | | 作成者（値なし=自分） |
| `--updated-by` | | 最終更新者（値なし=自分） |
| `--watched-by` | | Watch している人（値なし=自分） |
| `--category` | | カテゴリ前方一致 |
| `--tag` | `-t` | タグ |
| `--wip` | | WIP 状態（true/false） |
| `--sort` | `-s` | ソート対象（デフォルト: updated） |
| `--order` | | asc/desc（デフォルト: desc） |
| `--limit` | `-l` | 取得件数上限（デフォルト: 100） |
| `--output` | `-o` | 保存先ディレクトリ |
