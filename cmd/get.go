package cmd

import (
	"fmt"
	"os"
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
				return cliError("esa-mini get", err.Error(), "Set a valid token in ESA_ACCESS_TOKEN.")
			}

			post, err := client.GetPost(team, number)
			if err != nil {
				return cliError("esa-mini get", formatAPIError(err), fmt.Sprintf("Check the post number. Verify the post exists at https://%s.esa.io/posts/%d", team, number))
			}

			wip := post.WIP
			fm := frontmatter.Frontmatter{
				Number:         post.Number,
				Title:          post.Name,
				URL:            post.URL,
				Category:       post.Category,
				Tags:           post.Tags,
				WIP:            &wip,
				UpdatedAt:      post.UpdatedAt,
				RevisionNumber: post.RevisionNumber,
			}

			content, err := frontmatter.Format(fm, post.BodyMd)
			if err != nil {
				return cliError("esa-mini get", err.Error(), "This is an internal error. Please report it.")
			}

			if output == "-" {
				_, _ = fmt.Fprint(cmd.OutOrStdout(), content)
				return nil
			}

			if err := os.WriteFile(output, []byte(content), 0644); err != nil {
				return cliError("esa-mini get", err.Error(), "Check the output path is writable.")
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Saved: %s\nTitle: %s\nURL:   %s\n", output, post.Name, post.URL)
			return nil
		},
	}

	cmd.Flags().StringVar(&output, "output", "", "Output file path (required, use - for stdout)")
	_ = cmd.MarkFlagRequired("output")

	return cmd
}
