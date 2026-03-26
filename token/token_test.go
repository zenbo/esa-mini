package token

import (
	"os"
	"path/filepath"
	"testing"
)

func setupTestDir(t *testing.T) string {
	t.Helper()
	d := t.TempDir()
	t.Setenv("HOME", d)
	// XDG_CONFIG_HOME が設定されていると UserHomeDir の結果に影響しないが、
	// テストの独立性のためにクリアしておく
	t.Setenv("XDG_CONFIG_HOME", "")
	return d
}

func TestSaveAndLoad(t *testing.T) {
	d := setupTestDir(t)

	if err := Save("test-token-123"); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// ファイルが正しいパスに作られたか
	p := filepath.Join(d, ".config", "esa-mini", "token")
	if _, err := os.Stat(p); err != nil {
		t.Fatalf("token file not found: %v", err)
	}

	// パーミッション確認
	info, err := os.Stat(p)
	if err != nil {
		t.Fatalf("stat failed: %v", err)
	}
	if perm := info.Mode().Perm(); perm != 0o600 {
		t.Errorf("permission = %o, want 600", perm)
	}

	tok, err := Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if tok != "test-token-123" {
		t.Errorf("token = %q, want %q", tok, "test-token-123")
	}
}

func TestLoadNoFile(t *testing.T) {
	setupTestDir(t)

	tok, err := Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if tok != "" {
		t.Errorf("token = %q, want empty", tok)
	}
}

func TestDelete(t *testing.T) {
	setupTestDir(t)

	if err := Save("to-be-deleted"); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	if err := Delete(); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	tok, err := Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if tok != "" {
		t.Errorf("token = %q, want empty after delete", tok)
	}
}

func TestDeleteNoFile(t *testing.T) {
	setupTestDir(t)

	err := Delete()
	if err == nil {
		t.Fatal("expected error for deleting non-existent token")
	}
}

func TestResolveEnvOverFile(t *testing.T) {
	setupTestDir(t)

	if err := Save("file-token"); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	t.Setenv("ESA_ACCESS_TOKEN", "env-token")

	tok, err := Resolve()
	if err != nil {
		t.Fatalf("Resolve failed: %v", err)
	}
	if tok != "env-token" {
		t.Errorf("token = %q, want %q (env should take precedence)", tok, "env-token")
	}
}

func TestResolveFallbackToFile(t *testing.T) {
	setupTestDir(t)

	if err := Save("file-token"); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	t.Setenv("ESA_ACCESS_TOKEN", "")

	tok, err := Resolve()
	if err != nil {
		t.Fatalf("Resolve failed: %v", err)
	}
	if tok != "file-token" {
		t.Errorf("token = %q, want %q", tok, "file-token")
	}
}

func TestResolveNone(t *testing.T) {
	setupTestDir(t)
	t.Setenv("ESA_ACCESS_TOKEN", "")

	tok, err := Resolve()
	if err != nil {
		t.Fatalf("Resolve failed: %v", err)
	}
	if tok != "" {
		t.Errorf("token = %q, want empty", tok)
	}
}
