package cmd

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/zenbo/esa-mini/api"
)

func setupTestServer(t *testing.T, handler http.Handler) string {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	return server.URL
}

func TestTeamsCmd(t *testing.T) {
	server := setupTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := api.TeamsResponse{
			Teams: []api.Team{
				{Name: "docs", URL: "https://docs.esa.io"},
				{Name: "dev", URL: "https://dev.esa.io"},
			},
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
	cmd.SetArgs([]string{"teams"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "docs\thttps://docs.esa.io") {
		t.Errorf("output = %q, missing docs team", output)
	}
	if !strings.Contains(output, "dev\thttps://dev.esa.io") {
		t.Errorf("output = %q, missing dev team", output)
	}
}

func TestGetCmd(t *testing.T) {
	server := setupTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		post := api.Post{
			Number:    123,
			Name:      "テスト記事",
			BodyMd:    "# Hello",
			URL:       "https://docs.esa.io/posts/123",
			WIP:       false,
			Tags:      []string{"go"},
			Category:  "dev/tips",
			UpdatedAt: "2025-07-01T12:00:00+09:00",
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(post); err != nil {
			t.Fatal(err)
		}
	}))

	t.Setenv("ESA_ACCESS_TOKEN", "test-token")
	t.Setenv("ESA_API_BASE_URL", server)

	t.Run("save to file", func(t *testing.T) {
		outFile := filepath.Join(t.TempDir(), "test.md")
		cmd := NewRootCmd()
		out := &bytes.Buffer{}
		cmd.SetOut(out)
		cmd.SetArgs([]string{"get", "docs", "123", "--output", outFile})

		if err := cmd.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		stdout := out.String()
		if !strings.Contains(stdout, "Saved: "+outFile) {
			t.Errorf("stdout = %q, missing Saved", stdout)
		}
		if !strings.Contains(stdout, "Title: テスト記事") {
			t.Errorf("stdout = %q, missing Title", stdout)
		}

		content, err := os.ReadFile(outFile)
		if err != nil {
			t.Fatalf("failed to read output file: %v", err)
		}
		if !strings.Contains(string(content), "title: テスト記事") {
			t.Errorf("file content missing title frontmatter")
		}
		if !strings.Contains(string(content), "# Hello") {
			t.Errorf("file content missing body")
		}
	})

	t.Run("output to stdout", func(t *testing.T) {
		cmd := NewRootCmd()
		out := &bytes.Buffer{}
		cmd.SetOut(out)
		cmd.SetArgs([]string{"get", "docs", "123", "--output", "-"})

		if err := cmd.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		output := out.String()
		if !strings.Contains(output, "title: テスト記事") {
			t.Errorf("stdout = %q, missing frontmatter", output)
		}
		if !strings.Contains(output, "# Hello") {
			t.Errorf("stdout = %q, missing body", output)
		}
	})
}

func TestCreateCmd(t *testing.T) {
	server := setupTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %q, want POST", r.Method)
		}

		var req api.CreatePostRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode: %v", err)
		}
		if req.Post.Name != "新しい記事" {
			t.Errorf("Name = %q", req.Post.Name)
		}

		resp := api.Post{
			Number:    456,
			Name:      "新しい記事",
			URL:       "https://docs.esa.io/posts/456",
			BodyMd:    "本文",
			Category:  "dev/tips",
			Tags:      []string{"go"},
			WIP:       true,
			UpdatedAt: "2025-01-01T00:00:00+09:00",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			t.Fatal(err)
		}
	}))

	t.Setenv("ESA_ACCESS_TOKEN", "test-token")
	t.Setenv("ESA_API_BASE_URL", server)

	inputFile := filepath.Join(t.TempDir(), "input.md")
	content := "---\ntitle: 新しい記事\ncategory: dev/tips\ntags:\n  - go\nwip: true\n---\n\n本文\n"
	if err := os.WriteFile(inputFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cmd := NewRootCmd()
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetArgs([]string{"create", "docs", "--file", inputFile})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	stdout := out.String()
	if !strings.Contains(stdout, "Created: #456") {
		t.Errorf("stdout = %q, missing Created", stdout)
	}
	if !strings.Contains(stdout, "Title:   新しい記事") {
		t.Errorf("stdout = %q, missing Title", stdout)
	}

	// Verify frontmatter was written back to the file
	updated, err := os.ReadFile(inputFile)
	if err != nil {
		t.Fatalf("failed to read updated file: %v", err)
	}
	updatedStr := string(updated)
	if !strings.Contains(updatedStr, "number: 456") {
		t.Errorf("updated file missing number, got:\n%s", updatedStr)
	}
	if !strings.Contains(updatedStr, "url: https://docs.esa.io/posts/456") {
		t.Errorf("updated file missing url, got:\n%s", updatedStr)
	}
	if !strings.Contains(updatedStr, "updated_at:") {
		t.Errorf("updated file missing updated_at, got:\n%s", updatedStr)
	}
}

func TestCreateCmdCLIOverrides(t *testing.T) {
	var receivedReq api.CreatePostRequest

	server := setupTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&receivedReq); err != nil {
			t.Fatalf("failed to decode: %v", err)
		}
		resp := api.Post{Number: 456, Name: receivedReq.Post.Name, URL: "https://docs.esa.io/posts/456", BodyMd: "本文", UpdatedAt: "2025-01-01T00:00:00+09:00"}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			t.Fatal(err)
		}
	}))

	t.Setenv("ESA_ACCESS_TOKEN", "test-token")
	t.Setenv("ESA_API_BASE_URL", server)

	inputFile := filepath.Join(t.TempDir(), "input.md")
	content := "---\ntitle: 元のタイトル\ncategory: old/cat\ntags:\n  - old\nwip: true\n---\n\n本文\n"
	if err := os.WriteFile(inputFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cmd := NewRootCmd()
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetArgs([]string{
		"create", "docs",
		"--file", inputFile,
		"--name", "上書きタイトル",
		"--tags", "new1,new2",
		"--category", "new/cat",
		"--wip=false",
		"--message", "テストメッセージ",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if receivedReq.Post.Name != "上書きタイトル" {
		t.Errorf("Name = %q, want %q", receivedReq.Post.Name, "上書きタイトル")
	}
	if receivedReq.Post.Category != "new/cat" {
		t.Errorf("Category = %q, want %q", receivedReq.Post.Category, "new/cat")
	}
	if len(receivedReq.Post.Tags) != 2 || receivedReq.Post.Tags[0] != "new1" || receivedReq.Post.Tags[1] != "new2" {
		t.Errorf("Tags = %v", receivedReq.Post.Tags)
	}
	if receivedReq.Post.WIP == nil || *receivedReq.Post.WIP != false {
		t.Errorf("WIP = %v, want false", receivedReq.Post.WIP)
	}
	if receivedReq.Post.Message != "テストメッセージ" {
		t.Errorf("Message = %q", receivedReq.Post.Message)
	}
}

func TestUpdateCmd(t *testing.T) {
	server := setupTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("method = %q, want PATCH", r.Method)
		}
		resp := api.Post{
			Number: 123,
			Name:   "更新された記事",
			URL:    "https://docs.esa.io/posts/123",
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			t.Fatal(err)
		}
	}))

	t.Setenv("ESA_ACCESS_TOKEN", "test-token")
	t.Setenv("ESA_API_BASE_URL", server)

	inputFile := filepath.Join(t.TempDir(), "input.md")
	content := "---\ntitle: 更新された記事\n---\n\n更新本文\n"
	if err := os.WriteFile(inputFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cmd := NewRootCmd()
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetArgs([]string{"update", "docs", "123", "--file", inputFile})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	stdout := out.String()
	if !strings.Contains(stdout, "Updated: #123") {
		t.Errorf("stdout = %q, missing Updated", stdout)
	}
}

func TestMissingToken(t *testing.T) {
	t.Setenv("ESA_ACCESS_TOKEN", "")

	cmd := NewRootCmd()
	errOut := &bytes.Buffer{}
	cmd.SetErr(errOut)
	cmd.SetArgs([]string{"teams"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing token")
	}
	if !strings.Contains(err.Error(), "ESA_ACCESS_TOKEN") {
		t.Errorf("error = %q, should mention ESA_ACCESS_TOKEN", err.Error())
	}
}
