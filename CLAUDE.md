# CLAUDE.md

## 仕様・設計

- 詳細仕様: `SPEC.md`
- Agent Skill 定義: `SKILL.md`

## 規約

- エラーは exit 0 + stderr に What/Why/Hint 3行構造（`cmd/errors.go` 参照）
- stdout にはサマリーのみ。本文はファイルに保存しコンテキスト消費を抑える
- CLI オプションは frontmatter の値より優先
- golangci-lint errcheck: 戻り値はすべて明示的にハンドリングする

## 開発コマンド

```bash
go build -o esa-mini .
go test ./...
golangci-lint run ./...
```
