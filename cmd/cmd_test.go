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

	t.Run("save to directory", func(t *testing.T) {
		dir := t.TempDir()
		cmd := NewRootCmd()
		out := &bytes.Buffer{}
		cmd.SetOut(out)
		cmd.SetArgs([]string{"get", "docs", "123", "--output", dir})

		if err := cmd.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		expectedFile := filepath.Join(dir, "123.md")
		stdout := out.String()
		if !strings.Contains(stdout, "Saved: "+expectedFile) {
			t.Errorf("stdout = %q, missing Saved with auto-named path", stdout)
		}

		content, err := os.ReadFile(expectedFile)
		if err != nil {
			t.Fatalf("failed to read output file: %v", err)
		}
		if !strings.Contains(string(content), "title: テスト記事") {
			t.Errorf("file content missing title frontmatter")
		}
	})

	t.Run("save to directory with trailing slash", func(t *testing.T) {
		dir := filepath.Join(t.TempDir(), "subdir") + "/"
		cmd := NewRootCmd()
		out := &bytes.Buffer{}
		cmd.SetOut(out)
		cmd.SetArgs([]string{"get", "docs", "123", "--output", dir})

		if err := cmd.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		expectedFile := filepath.Join(dir, "123.md")
		if _, err := os.Stat(expectedFile); err != nil {
			t.Fatalf("expected file %s to exist: %v", expectedFile, err)
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
			Number:    123,
			Name:      "更新された記事",
			URL:       "https://docs.esa.io/posts/123",
			BodyMd:    "更新本文",
			Category:  "dev/tips",
			Tags:      []string{"go"},
			WIP:       false,
			UpdatedAt: "2025-02-01T12:00:00+09:00",
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			t.Fatal(err)
		}
	}))

	t.Setenv("ESA_ACCESS_TOKEN", "test-token")
	t.Setenv("ESA_API_BASE_URL", server)

	inputFile := filepath.Join(t.TempDir(), "input.md")
	content := "---\ntitle: 更新された記事\nupdated_at: 2025-01-01T00:00:00+09:00\n---\n\n更新本文\n"
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

	// Verify frontmatter was written back with latest server state
	updated, err := os.ReadFile(inputFile)
	if err != nil {
		t.Fatalf("failed to read updated file: %v", err)
	}
	updatedStr := string(updated)
	if !strings.Contains(updatedStr, "number: 123") {
		t.Errorf("updated file missing number, got:\n%s", updatedStr)
	}
	if !strings.Contains(updatedStr, "url: https://docs.esa.io/posts/123") {
		t.Errorf("updated file missing url, got:\n%s", updatedStr)
	}
	if !strings.Contains(updatedStr, "2025-02-01T12:00:00+09:00") {
		t.Errorf("updated file missing refreshed updated_at, got:\n%s", updatedStr)
	}
	if strings.Contains(updatedStr, "2025-01-01T00:00:00") {
		t.Errorf("updated file still contains stale updated_at, got:\n%s", updatedStr)
	}
}

func TestUpdateCmdNoWriteBack(t *testing.T) {
	server := setupTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := api.Post{
			Number:    123,
			Name:      "更新された記事",
			URL:       "https://docs.esa.io/posts/123",
			BodyMd:    "更新本文",
			UpdatedAt: "2025-02-01T12:00:00+09:00",
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			t.Fatal(err)
		}
	}))

	t.Setenv("ESA_ACCESS_TOKEN", "test-token")
	t.Setenv("ESA_API_BASE_URL", server)

	inputFile := filepath.Join(t.TempDir(), "input.md")
	original := "---\ntitle: 更新された記事\nupdated_at: 2025-01-01T00:00:00+09:00\n---\n\n更新本文\n"
	if err := os.WriteFile(inputFile, []byte(original), 0644); err != nil {
		t.Fatal(err)
	}

	cmd := NewRootCmd()
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetArgs([]string{"update", "docs", "123", "--file", inputFile, "--no-write-back"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	after, err := os.ReadFile(inputFile)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}
	if string(after) != original {
		t.Errorf("file was modified despite --no-write-back:\ngot:\n%s\nwant:\n%s", string(after), original)
	}
}

func TestCreateCmdNoWriteBack(t *testing.T) {
	server := setupTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := api.Post{
			Number:    456,
			Name:      "新しい記事",
			URL:       "https://docs.esa.io/posts/456",
			BodyMd:    "本文",
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
	original := "---\ntitle: 新しい記事\n---\n\n本文\n"
	if err := os.WriteFile(inputFile, []byte(original), 0644); err != nil {
		t.Fatal(err)
	}

	cmd := NewRootCmd()
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetArgs([]string{"create", "docs", "--file", inputFile, "--no-write-back"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	after, err := os.ReadFile(inputFile)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}
	if string(after) != original {
		t.Errorf("file was modified despite --no-write-back:\ngot:\n%s\nwant:\n%s", string(after), original)
	}
}

func TestGetCmdWritesTeamToFrontmatter(t *testing.T) {
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

	outFile := filepath.Join(t.TempDir(), "test.md")
	cmd := NewRootCmd()
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetArgs([]string{"get", "docs", "123", "--output", outFile})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content, err := os.ReadFile(outFile)
	if err != nil {
		t.Fatalf("failed to read output file: %v", err)
	}
	if !strings.Contains(string(content), "team: docs") {
		t.Errorf("file content missing team frontmatter, got:\n%s", string(content))
	}
}

func TestCreateCmdTeamFromFrontmatter(t *testing.T) {
	server := setupTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/teams/docs/") {
			t.Errorf("path = %q, want team 'docs'", r.URL.Path)
		}
		resp := api.Post{
			Number:    456,
			Name:      "新しい記事",
			URL:       "https://docs.esa.io/posts/456",
			BodyMd:    "本文",
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
	content := "---\nteam: docs\ntitle: 新しい記事\n---\n\n本文\n"
	if err := os.WriteFile(inputFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cmd := NewRootCmd()
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetArgs([]string{"create", "--file", inputFile})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	stdout := out.String()
	if !strings.Contains(stdout, "Created: #456") {
		t.Errorf("stdout = %q, missing Created: #456", stdout)
	}
}

func TestCreateCmdTeamMissing(t *testing.T) {
	inputFile := filepath.Join(t.TempDir(), "input.md")
	content := "---\ntitle: チームなし\n---\n\n本文\n"
	if err := os.WriteFile(inputFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	t.Setenv("ESA_ACCESS_TOKEN", "test-token")
	t.Setenv("ESA_TEAM", "")
	t.Setenv("HOME", t.TempDir())

	cmd := NewRootCmd()
	errOut := &bytes.Buffer{}
	cmd.SetErr(errOut)
	cmd.SetArgs([]string{"create", "--file", inputFile})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing team")
	}
	if !strings.Contains(err.Error(), "team is required") {
		t.Errorf("error = %q, should mention team", err.Error())
	}
}

func TestUpdateCmdAllFromFrontmatter(t *testing.T) {
	server := setupTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/teams/docs/") {
			t.Errorf("path = %q, want team 'docs'", r.URL.Path)
		}
		if !strings.HasSuffix(r.URL.Path, "/posts/789") {
			t.Errorf("path = %q, want suffix /posts/789", r.URL.Path)
		}
		resp := api.Post{
			Number: 789,
			Name:   "全frontmatterテスト",
			URL:    "https://docs.esa.io/posts/789",
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			t.Fatal(err)
		}
	}))

	t.Setenv("ESA_ACCESS_TOKEN", "test-token")
	t.Setenv("ESA_API_BASE_URL", server)

	inputFile := filepath.Join(t.TempDir(), "input.md")
	content := "---\nteam: docs\nnumber: 789\ntitle: 全frontmatterテスト\n---\n\n本文\n"
	if err := os.WriteFile(inputFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cmd := NewRootCmd()
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetArgs([]string{"update", "--file", inputFile})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	stdout := out.String()
	if !strings.Contains(stdout, "Updated: #789") {
		t.Errorf("stdout = %q, missing Updated: #789", stdout)
	}
}

func TestUpdateCmdTeamMissing(t *testing.T) {
	inputFile := filepath.Join(t.TempDir(), "input.md")
	content := "---\nnumber: 123\ntitle: チームなし\n---\n\n本文\n"
	if err := os.WriteFile(inputFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	t.Setenv("ESA_ACCESS_TOKEN", "test-token")
	t.Setenv("ESA_TEAM", "")
	t.Setenv("HOME", t.TempDir())

	cmd := NewRootCmd()
	errOut := &bytes.Buffer{}
	cmd.SetErr(errOut)
	cmd.SetArgs([]string{"update", "--file", inputFile})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing team")
	}
	if !strings.Contains(err.Error(), "team is required") {
		t.Errorf("error = %q, should mention team", err.Error())
	}
}

func TestUpdateCmdNumberFromFrontmatter(t *testing.T) {
	server := setupTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("method = %q, want PATCH", r.Method)
		}
		if !strings.HasSuffix(r.URL.Path, "/posts/456") {
			t.Errorf("path = %q, want suffix /posts/456", r.URL.Path)
		}
		resp := api.Post{
			Number: 456,
			Name:   "frontmatter番号テスト",
			URL:    "https://docs.esa.io/posts/456",
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			t.Fatal(err)
		}
	}))

	t.Setenv("ESA_ACCESS_TOKEN", "test-token")
	t.Setenv("ESA_API_BASE_URL", server)

	inputFile := filepath.Join(t.TempDir(), "input.md")
	content := "---\nnumber: 456\ntitle: frontmatter番号テスト\n---\n\n本文\n"
	if err := os.WriteFile(inputFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cmd := NewRootCmd()
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetArgs([]string{"update", "docs", "--file", inputFile})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	stdout := out.String()
	if !strings.Contains(stdout, "Updated: #456") {
		t.Errorf("stdout = %q, missing Updated: #456", stdout)
	}
}

func TestUpdateCmdNumberMissing(t *testing.T) {
	inputFile := filepath.Join(t.TempDir(), "input.md")
	content := "---\ntitle: 番号なし\n---\n\n本文\n"
	if err := os.WriteFile(inputFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	t.Setenv("ESA_ACCESS_TOKEN", "test-token")

	cmd := NewRootCmd()
	errOut := &bytes.Buffer{}
	cmd.SetErr(errOut)
	cmd.SetArgs([]string{"update", "docs", "--file", inputFile})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing number")
	}
	if !strings.Contains(err.Error(), "post number is required") {
		t.Errorf("error = %q, should mention post number", err.Error())
	}
}

func TestMissingToken(t *testing.T) {
	t.Setenv("ESA_ACCESS_TOKEN", "")
	t.Setenv("HOME", t.TempDir())

	cmd := NewRootCmd()
	errOut := &bytes.Buffer{}
	cmd.SetErr(errOut)
	cmd.SetArgs([]string{"teams"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing token")
	}
	if !strings.Contains(err.Error(), "no access token found") {
		t.Errorf("error = %q, should mention no access token found", err.Error())
	}
}

func TestTokenSetAndShow(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	// set
	cmd := NewRootCmd()
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetIn(strings.NewReader("my-secret-token\n"))
	cmd.SetArgs([]string{"token", "set"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("token set failed: %v", err)
	}
	if !strings.Contains(out.String(), "Token saved.") {
		t.Errorf("stdout = %q, missing 'Token saved.'", out.String())
	}

	// show
	cmd = NewRootCmd()
	out = &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetArgs([]string{"token", "show"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("token show failed: %v", err)
	}
	shown := strings.TrimSpace(out.String())
	if strings.Contains(shown, "my-secret-token") {
		t.Errorf("token show should mask the token, got %q", shown)
	}
	if !strings.HasPrefix(shown, "my-s") {
		t.Errorf("token show should show first 4 chars, got %q", shown)
	}
}

func TestTokenDelete(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	// set first
	cmd := NewRootCmd()
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetIn(strings.NewReader("disposable-token\n"))
	cmd.SetArgs([]string{"token", "set"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("token set failed: %v", err)
	}

	// delete
	cmd = NewRootCmd()
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetArgs([]string{"token", "delete"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("token delete failed: %v", err)
	}
	if !strings.Contains(out.String(), "Token deleted.") {
		t.Errorf("stdout = %q, missing 'Token deleted.'", out.String())
	}

	// show should fail
	cmd = NewRootCmd()
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetArgs([]string{"token", "show"})
	if err := cmd.Execute(); err == nil {
		t.Fatal("expected error for show after delete")
	}
}

func TestCategoriesCmd(t *testing.T) {
	server := setupTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := api.CategoriesPathsResponse{
			Categories: []api.CategoryPath{
				{Path: strPtr("dev/tips"), Posts: 12},
				{Path: strPtr("日報"), Posts: 1523},
				{Path: nil, Posts: 5},
			},
			TotalCount: 120,
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
	cmd.SetArgs([]string{"categories", "docs"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "dev/tips") {
		t.Errorf("output missing dev/tips: %q", output)
	}
	if !strings.Contains(output, "1523 posts") {
		t.Errorf("output missing 1523 posts: %q", output)
	}
	if !strings.Contains(output, "(uncategorized)") {
		t.Errorf("output missing (uncategorized): %q", output)
	}
	if !strings.Contains(output, "-- 3 categories (total 120) --") {
		t.Errorf("output missing summary: %q", output)
	}
}

func TestCategoriesCmdTop(t *testing.T) {
	server := setupTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, "/categories/top") {
			t.Errorf("path = %q, want /categories/top", r.URL.Path)
		}
		resp := api.CategoriesTopResponse{
			Categories: []api.TopCategory{
				{Name: "dev", FullName: "dev", Count: 15, HasChild: true},
				{Name: "日報", FullName: "日報", Count: 8, HasChild: false},
			},
			TotalCount: 2,
			Page:       1,
			PerPage:    20,
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
	cmd.SetArgs([]string{"categories", "docs", "--top"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "dev") {
		t.Errorf("output missing dev: %q", output)
	}
	if !strings.Contains(output, "[+]") {
		t.Errorf("output missing [+] indicator: %q", output)
	}
	if !strings.Contains(output, "-- 2 categories --") {
		t.Errorf("output missing summary: %q", output)
	}
}

func TestCategoriesCmdTopWithFilter(t *testing.T) {
	t.Setenv("ESA_ACCESS_TOKEN", "test-token")

	cmd := NewRootCmd()
	cmd.SetArgs([]string{"categories", "docs", "--top", "--prefix", "dev/"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for --top with --prefix")
	}
	if !strings.Contains(err.Error(), "--top cannot be used with --prefix or --match") {
		t.Errorf("error = %q", err.Error())
	}
}

func TestTagsCmd(t *testing.T) {
	server := setupTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := api.TagsResponse{
			Tags: []api.Tag{
				{Name: "go", PostsCount: 45},
				{Name: "React", PostsCount: 38},
				{Name: "設計", PostsCount: 23},
			},
			TotalCount: 150,
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
	cmd.SetArgs([]string{"tags", "docs"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "go") {
		t.Errorf("output missing go: %q", output)
	}
	if !strings.Contains(output, "45 posts") {
		t.Errorf("output missing 45 posts: %q", output)
	}
	if !strings.Contains(output, "-- 3 tags (total 150) --") {
		t.Errorf("output missing summary: %q", output)
	}
}

func TestTagsCmdMatch(t *testing.T) {
	server := setupTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := api.TagsResponse{
			Tags: []api.Tag{
				{Name: "go", PostsCount: 45},
				{Name: "React", PostsCount: 38},
				{Name: "設計", PostsCount: 23},
			},
			TotalCount: 150,
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
	cmd.SetArgs([]string{"tags", "docs", "--match", "react"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "React") {
		t.Errorf("output missing React: %q", output)
	}
	if strings.Contains(output, "go") && !strings.Contains(output, "-- 1 tags") {
		t.Errorf("output should only contain React: %q", output)
	}
	if !strings.Contains(output, "-- 1 tags (total 150) --") {
		t.Errorf("output missing summary: %q", output)
	}
}

func strPtr(s string) *string {
	return &s
}

func TestTokenDeleteNoToken(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	cmd := NewRootCmd()
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetArgs([]string{"token", "delete"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for deleting non-existent token")
	}
}

func TestTeamSetAndShowSingleTeam(t *testing.T) {
	server := setupTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := api.TeamsResponse{
			Teams: []api.Team{
				{Name: "docs", URL: "https://docs.esa.io"},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			t.Fatal(err)
		}
	}))

	t.Setenv("ESA_ACCESS_TOKEN", "test-token")
	t.Setenv("ESA_API_BASE_URL", server)
	t.Setenv("HOME", t.TempDir())

	// set (single team = auto-save)
	cmd := NewRootCmd()
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetArgs([]string{"team", "set"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("team set failed: %v", err)
	}
	if !strings.Contains(out.String(), "Team saved: docs") {
		t.Errorf("stdout = %q, missing 'Team saved: docs'", out.String())
	}

	// show
	cmd = NewRootCmd()
	out = &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetArgs([]string{"team", "show"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("team show failed: %v", err)
	}
	shown := strings.TrimSpace(out.String())
	if shown != "docs" {
		t.Errorf("team show = %q, want %q", shown, "docs")
	}
}

func TestTeamSetMultipleTeams(t *testing.T) {
	server := setupTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := api.TeamsResponse{
			Teams: []api.Team{
				{Name: "docs", URL: "https://docs.esa.io"},
				{Name: "dev", URL: "https://dev.esa.io"},
				{Name: "design", URL: "https://design.esa.io"},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			t.Fatal(err)
		}
	}))

	t.Setenv("ESA_ACCESS_TOKEN", "test-token")
	t.Setenv("ESA_API_BASE_URL", server)
	t.Setenv("HOME", t.TempDir())

	// select 2nd team
	cmd := NewRootCmd()
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetIn(strings.NewReader("2\n"))
	cmd.SetArgs([]string{"team", "set"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("team set failed: %v", err)
	}
	if !strings.Contains(out.String(), "Team saved: dev") {
		t.Errorf("stdout = %q, missing 'Team saved: dev'", out.String())
	}

	// show
	cmd = NewRootCmd()
	out = &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetArgs([]string{"team", "show"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("team show failed: %v", err)
	}
	shown := strings.TrimSpace(out.String())
	if shown != "dev" {
		t.Errorf("team show = %q, want %q", shown, "dev")
	}
}

func TestTeamSetWithArg(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("ESA_ACCESS_TOKEN", "test-token")

	// set with argument (no API call needed)
	cmd := NewRootCmd()
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetArgs([]string{"team", "set", "myteam"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("team set failed: %v", err)
	}
	if !strings.Contains(out.String(), "Team saved: myteam") {
		t.Errorf("stdout = %q, missing 'Team saved: myteam'", out.String())
	}

	// show
	cmd = NewRootCmd()
	out = &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetArgs([]string{"team", "show"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("team show failed: %v", err)
	}
	if strings.TrimSpace(out.String()) != "myteam" {
		t.Errorf("team show = %q, want %q", strings.TrimSpace(out.String()), "myteam")
	}
}

func TestTeamDeleteCmd(t *testing.T) {
	server := setupTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := api.TeamsResponse{
			Teams: []api.Team{{Name: "disposable", URL: "https://disposable.esa.io"}},
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			t.Fatal(err)
		}
	}))

	t.Setenv("ESA_ACCESS_TOKEN", "test-token")
	t.Setenv("ESA_API_BASE_URL", server)
	t.Setenv("HOME", t.TempDir())

	// set first
	cmd := NewRootCmd()
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetArgs([]string{"team", "set"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("team set failed: %v", err)
	}

	// delete
	cmd = NewRootCmd()
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetArgs([]string{"team", "delete"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("team delete failed: %v", err)
	}
	if !strings.Contains(out.String(), "Team deleted.") {
		t.Errorf("stdout = %q, missing 'Team deleted.'", out.String())
	}

	// show should fail
	cmd = NewRootCmd()
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetArgs([]string{"team", "show"})
	if err := cmd.Execute(); err == nil {
		t.Fatal("expected error for show after delete")
	}
}

func TestTeamDeleteNoTeam(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	cmd := NewRootCmd()
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetArgs([]string{"team", "delete"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for deleting non-existent team")
	}
}

func TestGetCmdWithDefaultTeam(t *testing.T) {
	server := setupTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/teams/docs/") {
			t.Errorf("path = %q, want team 'docs'", r.URL.Path)
		}
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
	t.Setenv("ESA_TEAM", "docs")

	outFile := filepath.Join(t.TempDir(), "test.md")
	cmd := NewRootCmd()
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetArgs([]string{"get", "123", "--output", outFile})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out.String(), "Saved:") {
		t.Errorf("stdout = %q, missing Saved", out.String())
	}
}

func TestGetCmdNoTeam(t *testing.T) {
	t.Setenv("ESA_ACCESS_TOKEN", "test-token")
	t.Setenv("ESA_TEAM", "")
	t.Setenv("HOME", t.TempDir())

	outFile := filepath.Join(t.TempDir(), "test.md")
	cmd := NewRootCmd()
	cmd.SetArgs([]string{"get", "123", "--output", outFile})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing team")
	}
	if !strings.Contains(err.Error(), "team is required") {
		t.Errorf("error = %q, should mention team", err.Error())
	}
}

func TestSearchCmdWithDefaultTeam(t *testing.T) {
	server := setupTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/teams/docs/") {
			t.Errorf("path = %q, want team 'docs'", r.URL.Path)
		}
		resp := api.PostsResponse{
			Posts:      []api.Post{},
			TotalCount: 0,
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
	t.Setenv("ESA_TEAM", "docs")

	cmd := NewRootCmd()
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetArgs([]string{"search", "--query", "test"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCategoriesCmdWithDefaultTeam(t *testing.T) {
	server := setupTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/teams/docs/") {
			t.Errorf("path = %q, want team 'docs'", r.URL.Path)
		}
		resp := api.CategoriesPathsResponse{
			Categories: []api.CategoryPath{},
			TotalCount: 0,
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
	t.Setenv("ESA_TEAM", "docs")

	cmd := NewRootCmd()
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetArgs([]string{"categories"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestTagsCmdWithDefaultTeam(t *testing.T) {
	server := setupTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/teams/docs/") {
			t.Errorf("path = %q, want team 'docs'", r.URL.Path)
		}
		resp := api.TagsResponse{
			Tags:       []api.Tag{},
			TotalCount: 0,
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
	t.Setenv("ESA_TEAM", "docs")

	cmd := NewRootCmd()
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetArgs([]string{"tags"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCreateCmdTeamFromConfig(t *testing.T) {
	server := setupTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/teams/docs/") {
			t.Errorf("path = %q, want team 'docs'", r.URL.Path)
		}
		resp := api.Post{
			Number:    456,
			Name:      "新しい記事",
			URL:       "https://docs.esa.io/posts/456",
			BodyMd:    "本文",
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
	t.Setenv("ESA_TEAM", "docs")

	inputFile := filepath.Join(t.TempDir(), "input.md")
	content := "---\ntitle: 新しい記事\n---\n\n本文\n"
	if err := os.WriteFile(inputFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cmd := NewRootCmd()
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetArgs([]string{"create", "--file", inputFile})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out.String(), "Created: #456") {
		t.Errorf("stdout = %q, missing Created: #456", out.String())
	}
}

func TestUpdateCmdTeamFromConfig(t *testing.T) {
	server := setupTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/teams/docs/") {
			t.Errorf("path = %q, want team 'docs'", r.URL.Path)
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
	t.Setenv("ESA_TEAM", "docs")

	inputFile := filepath.Join(t.TempDir(), "input.md")
	content := "---\nnumber: 123\ntitle: 更新された記事\n---\n\n更新本文\n"
	if err := os.WriteFile(inputFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cmd := NewRootCmd()
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetArgs([]string{"update", "--file", inputFile})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out.String(), "Updated: #123") {
		t.Errorf("stdout = %q, missing Updated: #123", out.String())
	}
}
