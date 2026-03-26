package token

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func dir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "esa-mini"), nil
}

func path() (string, error) {
	d, err := dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(d, "token"), nil
}

// Save はトークンを ~/.config/esa-mini/token に保存する。
func Save(tok string) error {
	d, err := dir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(d, 0o700); err != nil {
		return err
	}
	p := filepath.Join(d, "token")
	return os.WriteFile(p, []byte(tok+"\n"), 0o600)
}

// Load は保存済みトークンを読み込む。
// 見つからない場合は空文字列と nil を返す。
func Load() (string, error) {
	p, err := path()
	if err != nil {
		return "", err
	}
	data, err := os.ReadFile(p)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}

// Delete は保存済みトークンファイルを削除する。
func Delete() error {
	p, err := path()
	if err != nil {
		return err
	}
	err = os.Remove(p)
	if err != nil && os.IsNotExist(err) {
		return fmt.Errorf("no saved token found")
	}
	return err
}

// Resolve は環境変数 ESA_ACCESS_TOKEN > 設定ファイルの優先順位でトークンを解決する。
// どちらにもない場合は空文字列と nil を返す。
func Resolve() (string, error) {
	if tok := os.Getenv("ESA_ACCESS_TOKEN"); tok != "" {
		return tok, nil
	}
	return Load()
}
