---
name: esa-mini
description: >
  esa.io の記事を取得・作成・更新する最小限の CLI ツール。
  記事本文をファイルに保存し、コンテキストウィンドウの消費を抑える。
  使うタイミング: (1) esa の記事を読みたい・取得したい,
  (2) esa に記事を投稿・更新したい, (3) esa のチーム一覧を確認したい。
  環境変数 ESA_ACCESS_TOKEN が必要。
---

# esa-mini

## セットアップ

`esa-mini` が PATH に存在しない場合、先にインストールする。

```bash
go install github.com/zenbo/esa-mini@latest
```

`esa-mini --help` を実行して利用可能なコマンドとオプションを確認する。

## コマンド概要

- `esa-mini teams` — 所属チーム一覧を表示
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
