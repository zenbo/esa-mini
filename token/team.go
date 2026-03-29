package token

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// SaveTeam はチーム名を ~/.config/esa-mini/team に保存する。
func SaveTeam(name string) error {
	d, err := dir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(d, 0o700); err != nil {
		return err
	}
	p := filepath.Join(d, "team")
	return os.WriteFile(p, []byte(name+"\n"), 0o600)
}

// LoadTeam は保存済みチーム名を読み込む。
// 見つからない場合は空文字列と nil を返す。
func LoadTeam() (string, error) {
	d, err := dir()
	if err != nil {
		return "", err
	}
	p := filepath.Join(d, "team")
	data, err := os.ReadFile(p)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}

// DeleteTeam は保存済みチームファイルを削除する。
func DeleteTeam() error {
	d, err := dir()
	if err != nil {
		return err
	}
	p := filepath.Join(d, "team")
	err = os.Remove(p)
	if err != nil && os.IsNotExist(err) {
		return fmt.Errorf("no saved team found")
	}
	return err
}

// ResolveTeam は環境変数 ESA_TEAM > 設定ファイルの優先順位でチーム名を解決する。
// どちらにもない場合は空文字列と nil を返す。
func ResolveTeam() (string, error) {
	if team := os.Getenv("ESA_TEAM"); team != "" {
		return team, nil
	}
	return LoadTeam()
}
