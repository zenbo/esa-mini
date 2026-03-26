package cmd

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/zenbo/esa-mini/api"
)

func TestBuildQuery(t *testing.T) {
	tests := []struct {
		name      string
		query     string
		author    string
		updatedBy string
		watchedBy string
		category  string
		tag       string
		wip       string
		want      string
	}{
		{
			name:   "author only",
			author: "higuchi",
			want:   "@higuchi",
		},
		{
			name:      "watched-by and category",
			watchedBy: "higuchi",
			category:  "dev/tips",
			want:      "watched_by:higuchi in:dev/tips",
		},
		{
			name:      "all flags",
			query:     "Go入門",
			author:    "higuchi",
			updatedBy: "tanaka",
			watchedBy: "suzuki",
			category:  "dev",
			tag:       "go",
			wip:       "false",
			want:      "Go入門 @higuchi updated_by:tanaka watched_by:suzuki in:dev #go wip:false",
		},
		{
			name: "empty",
			want: "",
		},
		{
			name:  "query only",
			query: "esa-mini",
			want:  "esa-mini",
		},
		{
			name: "tag only",
			tag:  "release",
			want: "#release",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildQuery(tt.query, tt.author, tt.updatedBy, tt.watchedBy, tt.category, tt.tag, tt.wip)
			if got != tt.want {
				t.Errorf("buildQuery() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestSearchCmdList(t *testing.T) {
	server := setupTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/teams/docs/posts" {
			t.Errorf("path = %q", r.URL.Path)
		}
		resp := api.PostsResponse{
			Posts: []api.Post{
				{Number: 123, Name: "Go入門", Category: "dev/tips", UpdatedAt: "2026-03-26T12:00:00+09:00"},
				{Number: 456, Name: "Docker設定", Category: "dev/tips", UpdatedAt: "2026-03-24T12:00:00+09:00"},
			},
			TotalCount: 2,
			Page:       1,
			PerPage:    100,
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			t.Fatal(err)
		}
	}))

	t.Setenv("ESA_ACCESS_TOKEN", "test-token")
	t.Setenv("ESA_API_BASE_URL", server)

	cmd := NewRootCmd()
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetArgs([]string{"search", "docs", "--category", "dev/tips"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "123") {
		t.Errorf("output missing post 123: %s", output)
	}
	if !strings.Contains(output, "456") {
		t.Errorf("output missing post 456: %s", output)
	}
	if !strings.Contains(output, "2 posts found") {
		t.Errorf("output missing summary: %s", output)
	}
}

func TestSearchCmdSave(t *testing.T) {
	server := setupTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := api.PostsResponse{
			Posts: []api.Post{
				{
					Number:    123,
					Name:      "Go入門",
					Category:  "dev/tips",
					BodyMd:    "# Hello",
					URL:       "https://docs.esa.io/posts/123",
					Tags:      []string{"go"},
					WIP:       false,
					UpdatedAt: "2026-03-26T12:00:00+09:00",
				},
			},
			TotalCount: 1,
			Page:       1,
			PerPage:    100,
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			t.Fatal(err)
		}
	}))

	t.Setenv("ESA_ACCESS_TOKEN", "test-token")
	t.Setenv("ESA_API_BASE_URL", server)

	dir := t.TempDir()
	cmd := NewRootCmd()
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetArgs([]string{"search", "docs", "--category", "dev/tips", "--output", dir})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "Saved:") {
		t.Errorf("output missing Saved: %s", output)
	}
	if !strings.Contains(output, "1 posts saved") {
		t.Errorf("output missing summary: %s", output)
	}

	content, err := os.ReadFile(filepath.Join(dir, "123.md"))
	if err != nil {
		t.Fatalf("failed to read saved file: %v", err)
	}
	if !strings.Contains(string(content), "title: Go入門") {
		t.Errorf("saved file missing title frontmatter")
	}
	if !strings.Contains(string(content), "# Hello") {
		t.Errorf("saved file missing body")
	}
}

func TestSearchCmdOutputRequiresCondition(t *testing.T) {
	t.Setenv("ESA_ACCESS_TOKEN", "test-token")

	cmd := NewRootCmd()
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetArgs([]string{"search", "docs", "--output", t.TempDir()})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for --output without search condition")
	}
	if !strings.Contains(err.Error(), "--output requires at least one search condition") {
		t.Errorf("error = %q", err.Error())
	}
}

func TestSearchCmdRateLimit(t *testing.T) {
	server := setupTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = w.Write([]byte(`{"error":"too_many_requests","message":"Rate limit exceeded"}`))
	}))

	t.Setenv("ESA_ACCESS_TOKEN", "test-token")
	t.Setenv("ESA_API_BASE_URL", server)

	cmd := NewRootCmd()
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(errOut)
	cmd.SetArgs([]string{"search", "docs", "--category", "dev"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(errOut.String(), "Rate limit reached") {
		t.Errorf("stderr = %q, missing rate limit message", errOut.String())
	}
}
