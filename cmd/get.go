package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/zenbo/esa-mini/api"
	"github.com/zenbo/esa-mini/frontmatter"
)

func newGetCmd() *cobra.Command {
	var output string

	cmd := &cobra.Command{
		Use:   "get <team> <number>",
		Short: "Get a post and save as frontmatter Markdown",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			team := args[0]
			number, err := strconv.Atoi(args[1])
			if err != nil {
				return cliError("esa-mini get", fmt.Sprintf("invalid post number: %s", args[1]), "Post number must be an integer.")
			}

			client, err := api.NewClient()
			if err != nil {
				return cliError("esa-mini get", err.Error(), "Run 'esa-mini token set' or set ESA_ACCESS_TOKEN.")
			}

			post, err := client.GetPost(team, number)
			if err != nil {
				return cliError("esa-mini get", formatAPIError(err), fmt.Sprintf("Check the post number. Verify the post exists at https://%s.esa.io/posts/%d", team, number))
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
				return cliError("esa-mini get", err.Error(), "This is an internal error. Please report it.")
			}

			if output == "-" {
				_, _ = fmt.Fprint(cmd.OutOrStdout(), content)
				return nil
			}

			// ディレクトリ or 末尾 / → savePost に委譲
			isDirectory := false
			if info, statErr := os.Stat(output); statErr == nil && info.IsDir() {
				isDirectory = true
			} else if output[len(output)-1] == filepath.Separator || output[len(output)-1] == '/' {
				isDirectory = true
			}

			var outPath string
			if isDirectory {
				p, err := savePost(output, team, post)
				if err != nil {
					return cliError("esa-mini get", err.Error(), "Check the output directory is writable.")
				}
				outPath = p
			} else {
				if err := os.WriteFile(output, []byte(content), 0644); err != nil {
					return cliError("esa-mini get", err.Error(), "Check the output path is writable.")
				}
				outPath = output
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Saved: %s\nTitle: %s\nURL:   %s\n", outPath, post.Name, post.URL)
			return nil
		},
	}

	cmd.Flags().StringVar(&output, "output", "", "Output file path, directory, or - for stdout")
	_ = cmd.MarkFlagRequired("output")

	return cmd
}
