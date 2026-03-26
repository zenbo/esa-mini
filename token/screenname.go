package token

import (
	"os"
	"path/filepath"
	"strings"
)

// SaveScreenName は screen_name を ~/.config/esa-mini/screen_name に保存する。
func SaveScreenName(name string) error {
	d, err := dir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(d, 0o700); err != nil {
		return err
	}
	p := filepath.Join(d, "screen_name")
	return os.WriteFile(p, []byte(name+"\n"), 0o600)
}

// LoadScreenName は保存済み screen_name を読み込む。
// 見つからない場合は空文字列と nil を返す。
func LoadScreenName() (string, error) {
	d, err := dir()
	if err != nil {
		return "", err
	}
	p := filepath.Join(d, "screen_name")
	data, err := os.ReadFile(p)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}
