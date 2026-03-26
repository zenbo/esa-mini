package cmd

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/zenbo/esa-mini/api"
	"github.com/zenbo/esa-mini/frontmatter"
)

func newUpdateCmd() *cobra.Command {
	var (
		file     string
		name     string
		tags     string
		category string
		wip      bool
		message  string
	)

	cmd := &cobra.Command{
		Use:   "update [team] [number]",
		Short: "Update an existing post from a frontmatter Markdown file",
		Long: `Update an existing post from a frontmatter Markdown file.
Team and number can be omitted if the frontmatter contains 'team' and 'number' fields.
CLI arguments override frontmatter values.`,
		Args: cobra.RangeArgs(0, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			doc, err := readDocument(file)
			if err != nil {
				return cliError("esa-mini update", err.Error(), "Check the file path and format.")
			}

			// CLI args > frontmatter
			team := doc.Frontmatter.Team
			number := doc.Frontmatter.Number
			if len(args) >= 1 {
				team = args[0]
			}
			if len(args) >= 2 {
				number, err = strconv.Atoi(args[1])
				if err != nil {
					return cliError("esa-mini update", fmt.Sprintf("invalid post number: %s", args[1]), "Post number must be an integer.")
				}
			}
			if team == "" {
				return cliError("esa-mini update", "team is required", "Specify as argument or set 'team' in frontmatter.")
			}
			if number == 0 {
				return cliError("esa-mini update", "post number is required", "Specify as argument or set 'number' in frontmatter.")
			}

			body := api.UpdatePostBody{
				Name:   doc.Frontmatter.Title,
				BodyMd: frontmatter.NormalizeCRLF(doc.Body),
				Tags:   doc.Frontmatter.Tags,
			}
			if doc.Frontmatter.Category != "" {
				body.Category = doc.Frontmatter.Category
			}
			if doc.Frontmatter.WIP != nil {
				body.WIP = doc.Frontmatter.WIP
			}
			// CLI flags override frontmatter
			if cmd.Flags().Changed("name") {
				body.Name = name
			}
			if cmd.Flags().Changed("tags") {
				body.Tags = strings.Split(tags, ",")
			}
			if cmd.Flags().Changed("category") {
				body.Category = category
			}
			if cmd.Flags().Changed("wip") {
				body.WIP = &wip
			}
			if cmd.Flags().Changed("message") {
				body.Message = message
			}

			client, err := api.NewClient()
			if err != nil {
				return cliError("esa-mini update", err.Error(), "Run 'esa-mini token set' or set ESA_ACCESS_TOKEN.")
			}

			post, err := client.UpdatePost(team, number, body)
			if err != nil {
				return cliError("esa-mini update", formatAPIError(err), "Check your input and permissions.")
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Updated: #%d\nTitle:   %s\nURL:     %s\n", post.Number, post.Name, post.URL)
			return nil
		},
	}

	cmd.Flags().StringVar(&file, "file", "", "Frontmatter Markdown file path (required)")
	_ = cmd.MarkFlagRequired("file")
	cmd.Flags().StringVar(&name, "name", "", "Post title (overrides frontmatter)")
	cmd.Flags().StringVar(&tags, "tags", "", "Comma-separated tags (overrides frontmatter)")
	cmd.Flags().StringVar(&category, "category", "", "Category path (overrides frontmatter)")
	cmd.Flags().BoolVar(&wip, "wip", true, "WIP status (overrides frontmatter)")
	cmd.Flags().StringVar(&message, "message", "", "Commit message")

	return cmd
}
