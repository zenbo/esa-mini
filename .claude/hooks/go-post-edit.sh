#!/bin/bash
# PostToolUse hook: gofmt + golangci-lint on .go files after Write/Edit

file_path=$(echo "$TOOL_INPUT" | jq -r '.file_path // .filePath // empty')

if [ -z "$file_path" ] || [[ "$file_path" != *.go ]]; then
  exit 0
fi

gofmt -w "$file_path" 2>&1
golangci-lint run "$file_path" 2>&1
