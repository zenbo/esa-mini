package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/zenbo/esa-mini/api"
	"github.com/zenbo/esa-mini/frontmatter"
)

// savePost は Post を frontmatter 付き Markdown としてディレクトリに保存する。
// 保存先のファイルパスを返す。
func savePost(outputDir, team string, post *api.Post) (string, error) {
	// ディレクトリが存在しなければ作成
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create output directory: %w", err)
	}

	wip := post.WIP
	fm := frontmatter.Frontmatter{
		Team:      team,
		Number:    post.Number,
		Title:     post.Name,
		URL:       post.URL,
		Category:  post.Category,
		Tags:      post.Tags,
		WIP:       &wip,
		UpdatedAt: post.UpdatedAt,
	}

	content, err := frontmatter.Format(fm, post.BodyMd)
	if err != nil {
		return "", fmt.Errorf("failed to format post %d: %w", post.Number, err)
	}

	outPath := filepath.Join(outputDir, fmt.Sprintf("%d.md", post.Number))
	if err := os.WriteFile(outPath, []byte(content), 0644); err != nil {
		return "", fmt.Errorf("failed to write %s: %w", outPath, err)
	}

	return outPath, nil
}
