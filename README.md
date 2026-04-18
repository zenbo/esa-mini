# esa-mini

esa.io の記事操作に特化した最小限の CLI ツール。

[esa-mcp-server](https://github.com/esaio/esa-mcp-server) はツールが多く、MCP 経由だと記事本文が JSON テキストとしてコンテキストウィンドウを圧迫する。esa-mini は記事本文をローカルファイルに保存し、AI エージェントが必要なときだけ読める設計にすることでこの問題を解決する。

## インストール

```bash
go install github.com/zenbo/esa-mini@latest
```

GitHub Releases からバイナリをダウンロードすることもできる。

## 認証

トークンは `https://<team>.esa.io/user/tokens` で発行できる。

### 必要なスコープ

トークン発行時に以下のスコープを付与する。

全コマンドを使う場合: `read:post` `write:post` `read:team` `read:category` `read:tag` `read:user`

閲覧・検索のみの場合: `read:post` `read:team` `read:category` `read:tag` `read:user`

| コマンド | 必要スコープ |
|---|---|
| `teams` | `read:team` |
| `get` | `read:post` |
| `search` | `read:post`（`@me` 指定時は `read:user` も必要） |
| `categories` | `read:category` |
| `tags` | `read:tag` |
| `create` | `write:post` |
| `update` | `write:post` |

### トークンを保存する（推奨）

```bash
esa-mini token set
```

トークンは `~/.config/esa-mini/token` に保存され、どのプロジェクトからでも利用できる。

```bash
# 保存済みトークンの確認（マスク表示）
esa-mini token show

# 保存済みトークンの削除
esa-mini token delete
```

### 環境変数で設定する

```bash
export ESA_ACCESS_TOKEN='your-token-here'
```

環境変数が設定されている場合は保存済みトークンより優先される。

## デフォルトチーム

チーム名を保存しておくと、各コマンドで毎回指定する必要がなくなる。

### チーム名を保存する（推奨）

```bash
esa-mini team set
```

API からチーム一覧を取得し、1チームなら自動保存、複数チームなら選択肢を表示する。
チーム名を直接指定することもできる。

```bash
esa-mini team set myteam
```

チーム名は `~/.config/esa-mini/team` に保存される。

```bash
# 保存済みチーム名の確認
esa-mini team show

# 保存済みチーム名の削除
esa-mini team delete
```

### 環境変数で設定する

```bash
export ESA_TEAM='myteam'
```

環境変数が設定されている場合は保存済みチーム名より優先される。direnv 等でプロジェクトごとに切り替えると便利。

### 優先順位

CLI 引数 > 環境変数 `ESA_TEAM` > 設定ファイル（> frontmatter（create/update のみ））

## 使い方

```bash
# 所属チーム一覧
esa-mini teams

# カテゴリ一覧（デフォルトチーム設定済みなら team 省略可）
esa-mini categories myteam

# トップレベルカテゴリのみ
esa-mini categories myteam --top

# カテゴリを部分一致で探す
esa-mini categories myteam --match "設計"

# タグ一覧
esa-mini tags myteam

# タグを絞り込み
esa-mini tags myteam --match "go"

# 記事を検索して一覧表示
esa-mini search myteam --category "dev/tips"

# 検索結果を一括ダウンロード
esa-mini search myteam --watched-by --output ./posts/

# 記事を取得してファイルに保存
esa-mini get myteam 123 --output ./posts/123.md

# ディレクトリを指定すると {number}.md で自動命名
esa-mini get myteam 123 --output ./posts/

# ファイルから新規記事を作成
esa-mini create myteam --file ./posts/new.md

# ファイルから既存記事を更新
esa-mini update myteam 123 --file ./posts/123.md

# create / update はデフォルトでサーバの最新状態（number, url, updated_at 等）を
# 入力ファイルの frontmatter に書き戻す。書き戻したくない場合は --no-write-back。
esa-mini update myteam 123 --file ./posts/123.md --no-write-back

# frontmatter に team / number があれば引数を省略できる
esa-mini create --file ./posts/new.md
esa-mini update --file ./posts/123.md

# デフォルトチーム設定済みなら team 引数を省略できる
esa-mini get 123 --output ./posts/123.md
esa-mini search --category "dev/tips"
esa-mini categories --match "設計"
esa-mini tags --match "go"
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
