package frontmatter

import (
	"strings"
	"testing"
)

func boolPtr(b bool) *bool { return &b }

func TestParse(t *testing.T) {
	t.Run("full frontmatter with body", func(t *testing.T) {
		input := `---
number: 123
title: テスト記事
url: https://docs.esa.io/posts/123
category: dev/tips
tags:
  - go
  - cli
wip: false
updated_at: "2025-07-01T12:00:00+09:00"
revision_number: 5
---

本文がここに続く
`
		doc, err := Parse(strings.NewReader(input))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		fm := doc.Frontmatter
		if fm.Number != 123 {
			t.Errorf("Number = %d, want 123", fm.Number)
		}
		if fm.Title != "テスト記事" {
			t.Errorf("Title = %q, want %q", fm.Title, "テスト記事")
		}
		if fm.URL != "https://docs.esa.io/posts/123" {
			t.Errorf("URL = %q", fm.URL)
		}
		if fm.Category != "dev/tips" {
			t.Errorf("Category = %q", fm.Category)
		}
		if len(fm.Tags) != 2 || fm.Tags[0] != "go" || fm.Tags[1] != "cli" {
			t.Errorf("Tags = %v, want [go cli]", fm.Tags)
		}
		if fm.WIP == nil || *fm.WIP != false {
			t.Errorf("WIP = %v, want false", fm.WIP)
		}
		if fm.UpdatedAt != "2025-07-01T12:00:00+09:00" {
			t.Errorf("UpdatedAt = %q", fm.UpdatedAt)
		}
		if fm.RevisionNumber != 5 {
			t.Errorf("RevisionNumber = %d, want 5", fm.RevisionNumber)
		}
		if doc.Body != "本文がここに続く" {
			t.Errorf("Body = %q, want %q", doc.Body, "本文がここに続く")
		}
	})

	t.Run("minimal frontmatter for create", func(t *testing.T) {
		input := `---
title: 新しい記事
wip: true
---

Hello World
`
		doc, err := Parse(strings.NewReader(input))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if doc.Frontmatter.Title != "新しい記事" {
			t.Errorf("Title = %q", doc.Frontmatter.Title)
		}
		if doc.Frontmatter.WIP == nil || *doc.Frontmatter.WIP != true {
			t.Errorf("WIP = %v, want true", doc.Frontmatter.WIP)
		}
		if doc.Body != "Hello World" {
			t.Errorf("Body = %q", doc.Body)
		}
	})

	t.Run("missing opening delimiter", func(t *testing.T) {
		input := `title: foo
---

body
`
		_, err := Parse(strings.NewReader(input))
		if err == nil {
			t.Fatal("expected error for missing opening ---")
		}
	})

	t.Run("missing closing delimiter", func(t *testing.T) {
		input := `---
title: foo
body without closing
`
		_, err := Parse(strings.NewReader(input))
		if err == nil {
			t.Fatal("expected error for missing closing ---")
		}
	})

	t.Run("empty body", func(t *testing.T) {
		input := `---
title: no body
---
`
		doc, err := Parse(strings.NewReader(input))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if doc.Frontmatter.Title != "no body" {
			t.Errorf("Title = %q", doc.Frontmatter.Title)
		}
		if doc.Body != "" {
			t.Errorf("Body = %q, want empty", doc.Body)
		}
	})

	t.Run("multiline body", func(t *testing.T) {
		input := `---
title: multi
---

line1
line2
line3
`
		doc, err := Parse(strings.NewReader(input))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if doc.Body != "line1\nline2\nline3" {
			t.Errorf("Body = %q", doc.Body)
		}
	})
}

func TestFormat(t *testing.T) {
	t.Run("full frontmatter", func(t *testing.T) {
		fm := Frontmatter{
			Number:         123,
			Title:          "テスト記事",
			URL:            "https://docs.esa.io/posts/123",
			Category:       "dev/tips",
			Tags:           []string{"go", "cli"},
			WIP:            boolPtr(false),
			UpdatedAt:      "2025-07-01T12:00:00+09:00",
			RevisionNumber: 5,
		}
		result, err := Format(fm, "本文")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.HasPrefix(result, "---\n") {
			t.Error("should start with ---")
		}
		if !strings.Contains(result, "title: テスト記事") {
			t.Error("should contain title")
		}
		if !strings.Contains(result, "number: 123") {
			t.Error("should contain number")
		}
		if !strings.Contains(result, "本文") {
			t.Error("should contain body")
		}
		if !strings.HasSuffix(result, "\n") {
			t.Error("should end with newline")
		}
	})

	t.Run("roundtrip", func(t *testing.T) {
		fm := Frontmatter{
			Title:    "Roundtrip Test",
			Category: "test",
			Tags:     []string{"a", "b"},
			WIP:      boolPtr(true),
		}
		body := "This is the body."
		formatted, err := Format(fm, body)
		if err != nil {
			t.Fatalf("Format error: %v", err)
		}

		doc, err := Parse(strings.NewReader(formatted))
		if err != nil {
			t.Fatalf("Parse error: %v", err)
		}
		if doc.Frontmatter.Title != fm.Title {
			t.Errorf("Title = %q, want %q", doc.Frontmatter.Title, fm.Title)
		}
		if doc.Frontmatter.Category != fm.Category {
			t.Errorf("Category = %q, want %q", doc.Frontmatter.Category, fm.Category)
		}
		if doc.Body != body {
			t.Errorf("Body = %q, want %q", doc.Body, body)
		}
	})

	t.Run("trailing newline added", func(t *testing.T) {
		fm := Frontmatter{Title: "test"}
		result, err := Format(fm, "no trailing newline")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.HasSuffix(result, "no trailing newline\n") {
			t.Error("should add trailing newline")
		}
	})

	t.Run("body already has trailing newline", func(t *testing.T) {
		fm := Frontmatter{Title: "test"}
		result, err := Format(fm, "has newline\n")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if strings.HasSuffix(result, "\n\n") {
			t.Error("should not double trailing newline")
		}
	})
}
