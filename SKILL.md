---
name: esa-mini
description: >
  esa.io の記事を取得・作成・更新する最小限の CLI ツール。
  記事本文をファイルに保存し、コンテキストウィンドウの消費を抑える。
  使うタイミング: (1) esa の記事を読みたい・取得したい,
  (2) esa の記事を検索・一括ダウンロードしたい,
  (3) esa に記事を投稿・更新したい, (4) esa のチーム一覧を確認したい。
  注意: esa API は1ユーザーあたり15分間に75リクエストのレートリミットがある。大量取得時は --limit で件数を制限すること。
  `esa-mini token set` でトークンを保存するか、環境変数 ESA_ACCESS_TOKEN を設定する。
---

# esa-mini

## セットアップ

`esa-mini` が PATH に存在しない場合、先にインストールする。

```bash
go install github.com/zenbo/esa-mini@latest
```

`esa-mini --help` を実行して利用可能なコマンドとオプションを確認する。

## コマンド概要

- `esa-mini token set` — アクセストークンを `~/.config/esa-mini/token` に保存
- `esa-mini token show` — 保存済みトークンを確認（マスク表示）
- `esa-mini token delete` — 保存済みトークンを削除
- `esa-mini teams` — 所属チーム一覧を表示
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
