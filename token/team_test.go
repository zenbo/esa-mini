package token

import (
	"os"
	"path/filepath"
	"testing"
)

func TestTeamSaveAndLoad(t *testing.T) {
	d := setupTestDir(t)

	if err := SaveTeam("myteam"); err != nil {
		t.Fatalf("SaveTeam failed: %v", err)
	}

	// ファイルが正しいパスに作られたか
	p := filepath.Join(d, ".config", "esa-mini", "team")
	if _, err := os.Stat(p); err != nil {
		t.Fatalf("team file not found: %v", err)
	}

	// パーミッション確認
	info, err := os.Stat(p)
	if err != nil {
		t.Fatalf("stat failed: %v", err)
	}
	if perm := info.Mode().Perm(); perm != 0o600 {
		t.Errorf("permission = %o, want 600", perm)
	}

	team, err := LoadTeam()
	if err != nil {
		t.Fatalf("LoadTeam failed: %v", err)
	}
	if team != "myteam" {
		t.Errorf("team = %q, want %q", team, "myteam")
	}
}

func TestTeamLoadNoFile(t *testing.T) {
	setupTestDir(t)

	team, err := LoadTeam()
	if err != nil {
		t.Fatalf("LoadTeam failed: %v", err)
	}
	if team != "" {
		t.Errorf("team = %q, want empty", team)
	}
}

func TestTeamDelete(t *testing.T) {
	setupTestDir(t)

	if err := SaveTeam("to-be-deleted"); err != nil {
		t.Fatalf("SaveTeam failed: %v", err)
	}

	if err := DeleteTeam(); err != nil {
		t.Fatalf("DeleteTeam failed: %v", err)
	}

	team, err := LoadTeam()
	if err != nil {
		t.Fatalf("LoadTeam failed: %v", err)
	}
	if team != "" {
		t.Errorf("team = %q, want empty after delete", team)
	}
}

func TestTeamDeleteNoFile(t *testing.T) {
	setupTestDir(t)

	err := DeleteTeam()
	if err == nil {
		t.Fatal("expected error for deleting non-existent team")
	}
}

func TestTeamResolveEnvOverFile(t *testing.T) {
	setupTestDir(t)

	if err := SaveTeam("file-team"); err != nil {
		t.Fatalf("SaveTeam failed: %v", err)
	}

	t.Setenv("ESA_TEAM", "env-team")

	team, err := ResolveTeam()
	if err != nil {
		t.Fatalf("ResolveTeam failed: %v", err)
	}
	if team != "env-team" {
		t.Errorf("team = %q, want %q (env should take precedence)", team, "env-team")
	}
}

func TestTeamResolveFallbackToFile(t *testing.T) {
	setupTestDir(t)

	if err := SaveTeam("file-team"); err != nil {
		t.Fatalf("SaveTeam failed: %v", err)
	}

	t.Setenv("ESA_TEAM", "")

	team, err := ResolveTeam()
	if err != nil {
		t.Fatalf("ResolveTeam failed: %v", err)
	}
	if team != "file-team" {
		t.Errorf("team = %q, want %q", team, "file-team")
	}
}

func TestTeamResolveNone(t *testing.T) {
	setupTestDir(t)
	t.Setenv("ESA_TEAM", "")

	team, err := ResolveTeam()
	if err != nil {
		t.Fatalf("ResolveTeam failed: %v", err)
	}
	if team != "" {
		t.Errorf("team = %q, want empty", team)
	}
}
