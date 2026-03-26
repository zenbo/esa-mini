# esa-mini

esa.io の記事操作に特化した最小限の CLI ツール。

[esa-mcp-server](https://github.com/esaio/esa-mcp-server) はツールが多く、MCP 経由だと記事本文が JSON テキストとしてコンテキストウィンドウを圧迫する。esa-mini は記事本文をローカルファイルに保存し、AI エージェントが必要なときだけ読める設計にすることでこの問題を解決する。

## インストール

```bash
go install github.com/zenbo/esa-mini@latest
```

GitHub Releases からバイナリをダウンロードすることもできる。

## 認証

環境変数 `ESA_ACCESS_TOKEN` にパーソナルアクセストークンを設定する。

```bash
export ESA_ACCESS_TOKEN='your-token-here'
```

トークンは `https://<team>.esa.io/user/tokens` で発行できる。

## 使い方

```bash
# 所属チーム一覧
esa-mini teams

# 記事を取得してファイルに保存
esa-mini get myteam 123 --output ./posts/123.md

# ディレクトリを指定すると {number}.md で自動命名
esa-mini get myteam 123 --output ./posts/

# ファイルから新規記事を作成
esa-mini create myteam --file ./posts/new.md

# ファイルから既存記事を更新
esa-mini update myteam 123 --file ./posts/123.md

# frontmatter に team / number があれば引数を省略できる
esa-mini create --file ./posts/new.md
esa-mini update --file ./posts/123.md
```

記事ファイルは YAML frontmatter 付き Markdown 形式（`get` で取得したファイルにはすべてのフィールドが含まれる）:

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

## 開発

### 前提

[mise](https://mise.jdx.dev/) でツールバージョンを管理している。

```bash
mise install
```

### ビルド・テスト

```bash
go build -o esa-mini .
go test ./...
golangci-lint run ./...
```

### Git フック

[lefthook](https://github.com/evilmartians/lefthook) で pre-commit 時に gofmt, golangci-lint, go test を実行する。

### リリース

`v*` タグのプッシュで GitHub Actions + [GoReleaser](https://goreleaser.com/) が自動リリースを行う。

## Agent Skill

Claude Code のスキルとしてインストールできる:

```bash
npx skills add zenbo/esa-mini
```
