package cmd

import (
	"os"

	"github.com/zenbo/esa-mini/api"
	"github.com/zenbo/esa-mini/frontmatter"
)

// writeBackPost writes the post metadata and body back to the given file as
// frontmatter Markdown, so that the local file reflects the current server state
// (updated_at, number, url, etc.) after create/update.
func writeBackPost(path string, team string, post *api.Post) error {
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
		return err
	}
	return os.WriteFile(path, []byte(content), 0644)
}
