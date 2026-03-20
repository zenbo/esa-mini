# esa-mini 仕様書

esa.io の記事操作に特化した最小限の CLI ツール。
Claude Code からの利用を想定し、コンテキストウィンドウの消費を最小化する設計。

## 背景

- esa-mcp-server (https://github.com/esaio/esa-mcp-server) はツールが 25 個と多機能すぎる
- MCP 経由だとレスポンスがすべて JSON text でコンテキストに載り、長い記事でトークンを大量消費する
- CLI 版なら本文をファイルに保存し、Claude Code は必要なときだけ Read ツールで読める

## コマンド一覧（4つ）

### `esa-mini teams`

所属チーム一覧を表示する。

**stdout:**

```
docs   https://docs.esa.io
dev    https://dev.esa.io
```

### `esa-mini get <team> <number> --output <path>`

記事を取得し、frontmatter 付き Markdown ファイルとして保存する。

- `--output`（必須）: 保存先ファイルパス
- `--output -` で stdout に出力

**保存されるファイル:**

```markdown
---
number: 123
title: 記事タイトル
url: https://docs.esa.io/posts/123
category: dev/tips
tags:
  - go
  - cli
wip: false
updated_at: 2025-07-01T12:00:00+09:00
revision_number: 5
---

本文 Markdown がここに続く...
```

**stdout:**

```
Saved: ./posts/123.md
Title: 記事タイトル
URL:   https://docs.esa.io/posts/123
```

### `esa-mini create <team> --file <path>`

ローカルの frontmatter 付き Markdown ファイルから新規記事を作成する。

- `--file`（必須）: 投稿するファイルパス
- frontmatter の `title`, `tags`, `category`, `wip` を読み取る
- `--help` にファイルフォーマットの例を表示する
- CLI オプションで明示指定した場合は frontmatter より優先:
  - `--name <title>`
  - `--tags <tag1,tag2>`
  - `--category <path>`
  - `--wip`（フラグ、デフォルト true）
  - `--message <text>`

**stdout:**

```
Created: #456
Title:   新しい記事
URL:     https://docs.esa.io/posts/456
```

### `esa-mini update <team> <number> --file <path>`

ローカルの frontmatter 付き Markdown ファイルから既存記事を更新する。

- `--file`（必須）: 投稿するファイルパス
- frontmatter の `title`, `tags`, `category`, `wip` を読み取る
- CLI オプションで明示指定した場合は frontmatter より優先（create と同様）
- frontmatter に `revision_number` があれば `original_revision` として送信し、コンフリクト検出を行う

**stdout:**

```
Updated: #123
Title:   更新された記事
URL:     https://docs.esa.io/posts/123
```

コンフリクト検出時:

```
Updated: #123 (conflict detected, merged by server)
Title:   更新された記事
URL:     https://docs.esa.io/posts/123
```

## 認証

環境変数 `ESA_ACCESS_TOKEN` でパーソナルアクセストークンを指定する。
トークンは `https://<team>.esa.io/user/tokens` で発行できる。

## エラー設計

すべてのエラーは **exit code 0** で返し、詳細を **stderr** に出力する。
Claude Code がエラー内容を読み取り、自律的にリカバリできることを目指す。

### フォーマット: What / Why / Hint

```
Error: esa-mini get failed
Why:   404 Not Found - post number 999 does not exist in team "docs"
Hint:  Check the post number. Use "esa-mini search" on esa.io to find the correct number.
```

```
Error: esa-mini create failed
Why:   401 Unauthorized - invalid or expired access token
Hint:  Set a valid token in ESA_ACCESS_TOKEN. Generate one at https://docs.esa.io/user/tokens
```

### エラー時の原則

- stdout には何も出さない（パイプ安全）
- stderr の 3 行構造で、エージェントが原因と次のアクションを判断できるようにする
- スタックトレースやデバッグ情報は出さない

## 技術スタック

- Go (>=1.22)
- 標準ライブラリの `net/http` で API 通信（外部 HTTP 依存なし）
- `encoding/json` でレスポンス解析
- CLI 引数パース: 標準ライブラリ or 軽量ライブラリ（cobra 等）
- frontmatter パース: YAML ライブラリ

## 出力設計（コンテキスト節約のポイント）

| コマンド | stdout | ファイル |
|---|---|---|
| teams | チーム一覧（1行1チーム） | なし |
| get | 保存先パス・タイトル・URL | frontmatter 付き Markdown |
| create | 番号・タイトル・URL | なし |
| update | 番号・タイトル・URL | なし |

- JSON 出力ではなくプレーンテキスト（トークン効率が良い）
- 本文はファイルに分離し、Claude Code が Read で必要なときだけ読む

## Frontmatter 仕様

get で保存されるファイルの frontmatter フィールド:

| フィールド | 型 | 説明 |
|---|---|---|
| number | int | 記事番号 |
| title | string | 記事名 |
| url | string | 記事 URL |
| category | string | カテゴリパス（例: `dev/tips`） |
| tags | string[] | タグ一覧 |
| wip | bool | WIP 状態 |
| updated_at | string | 最終更新日時 (ISO8601) |
| revision_number | int | リビジョン番号（update 時のコンフリクト検出に使用） |

create/update 時に読み取るフィールド:

| フィールド | 型 | 用途 |
|---|---|---|
| title | string | 記事名（`--name` で上書き可） |
| category | string | カテゴリ（`--category` で上書き可） |
| tags | string[] | タグ（`--tags` で上書き可） |
| wip | bool | WIP 状態（`--wip` で上書き可） |
| revision_number | int | update 時の `original_revision.number` に使用 |

## esa API リファレンス

Base URL: `https://api.esa.io`
認証ヘッダー: `Authorization: Bearer <token>`

### GET /v1/teams

チーム一覧を取得。レスポンス: `{ teams: [{ name, url, description, icon }] }`

### GET /v1/teams/{team_name}/posts/{post_number}

記事を取得。レスポンスに `body_md`, `name`, `full_name`, `url`, `wip`, `tags`, `category`, `updated_at`, `revision_number` 等が含まれる。

### POST /v1/teams/{team_name}/posts

記事作成。body: `{ post: { name, body_md, tags, category, wip, message } }`

### PATCH /v1/teams/{team_name}/posts/{post_number}

記事更新。body: `{ post: { name, body_md, tags, category, wip, message, original_revision } }`
`original_revision: { number }` を指定するとコンフリクト検出が有効になる。

## Agent Skill 対応

`npx skills add zenbo/esa-mini` でインストール可能な SKILL.md を同梱する。
スキルは「esa の記事操作が必要」というトリガーで esa-mini の存在に気づかせる役割を持ち、
具体的な使い方は `esa-mini --help` に委ねる。

**SKILL.md:**

```markdown
---
name: esa-mini
description: esa.io の記事を取得・作成・更新する CLI ツール。esa の記事操作が必要なときに使う。
---

# esa-mini

esa.io の記事操作に特化した最小限の CLI。

## トリガー

- esa の記事を読みたい / 取得したい
- esa に記事を投稿・更新したい
- esa のチーム一覧を確認したい

## 使い方

\`esa-mini --help\` を実行して利用可能なコマンドとオプションを確認すること。

## 新規記事の作成

以下のフォーマットで Markdown ファイルを作成し、\`esa-mini create <team> --file <path>\` に渡す。

\`\`\`markdown
---
title: 記事タイトル
category: dev/tips
tags:
  - go
  - cli
wip: true
---

本文をここに書く
\`\`\`
```

## ディレクトリ構成

```
esa-mini/
  go.mod
  go.sum
  main.go           # エントリポイント
  SKILL.md           # Agent Skill 定義
  cmd/               # CLI コマンド定義
    root.go
    teams.go
    get.go
    create.go
    update.go
  api/               # esa API クライアント
    client.go
    types.go
  frontmatter/       # frontmatter パース・生成
    frontmatter.go
```

## ビルド・配布

- GitHub Actions でクロスコンパイルし、リリースアーティファクトとして配布
- タグプッシュ（`v*`）をトリガーに自動リリース
- 対象プラットフォーム:
  - `darwin/arm64`
  - `linux/amd64`
  - `linux/arm64`
  - `windows/amd64`
- GoReleaser の利用を検討（CHANGELOG 自動生成、checksums 付与）

### インストール方法

```bash
# GitHub Releases からダウンロード
curl -L https://github.com/zenbo/esa-mini/releases/latest/download/esa-mini_darwin_arm64.tar.gz | tar xz
sudo mv esa-mini /usr/local/bin/

# または go install
go install github.com/zenbo/esa-mini@latest
```

## 参考リポジトリ

- https://github.com/esaio/esa-mcp-server （API 型定義、エンドポイント仕様の参考に）
