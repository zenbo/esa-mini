package token

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSaveAndLoadScreenName(t *testing.T) {
	d := setupTestDir(t)

	if err := SaveScreenName("higuchi"); err != nil {
		t.Fatalf("SaveScreenName failed: %v", err)
	}

	p := filepath.Join(d, ".config", "esa-mini", "screen_name")
	if _, err := os.Stat(p); err != nil {
		t.Fatalf("screen_name file not found: %v", err)
	}

	name, err := LoadScreenName()
	if err != nil {
		t.Fatalf("LoadScreenName failed: %v", err)
	}
	if name != "higuchi" {
		t.Errorf("screen_name = %q, want %q", name, "higuchi")
	}
}

func TestLoadScreenNameNoFile(t *testing.T) {
	setupTestDir(t)

	name, err := LoadScreenName()
	if err != nil {
		t.Fatalf("LoadScreenName failed: %v", err)
	}
	if name != "" {
		t.Errorf("screen_name = %q, want empty", name)
	}
}
