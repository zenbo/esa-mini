package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/zenbo/esa-mini/api"
	"github.com/zenbo/esa-mini/frontmatter"
	"github.com/zenbo/esa-mini/token"
)

func newCreateCmd() *cobra.Command {
	var (
		file        string
		name        string
		tags        string
		category    string
		wip         bool
		message     string
		noWriteBack bool
	)

	cmd := &cobra.Command{
		Use:   "create [team]",
		Short: "Create a new post from a frontmatter Markdown file",
		Long: `Create a new post from a frontmatter Markdown file.
Team can be omitted if the frontmatter contains a 'team' field,
ESA_TEAM is set, or a default team is saved via 'esa-mini team set'.

File format:
  ---
  team: myteam
  title: Post title
  category: dev/tips
  tags:
    - go
    - cli
  wip: true
  ---

  Post body here`,
		Args: cobra.RangeArgs(0, 1),
		RunE: func(cmd *cobra.Command, args []string) error {
			doc, err := readDocument(file)
			if err != nil {
				return cliError("esa-mini create", err.Error(), "Check the file path and format.")
			}

			// CLI arg > frontmatter > env/config
			team := doc.Frontmatter.Team
			if len(args) >= 1 {
				team = args[0]
			}
			if team == "" {
				resolved, err := token.ResolveTeam()
				if err != nil {
					return cliError("esa-mini create", err.Error(), "Check config permissions.")
				}
				team = resolved
			}
			if team == "" {
				return cliError("esa-mini create", "team is required", "Specify as argument, set 'team' in frontmatter, set ESA_TEAM, or run 'esa-mini team set'.")
			}

			body := api.PostBody{
				Name:   doc.Frontmatter.Title,
				BodyMd: frontmatter.NormalizeCRLF(doc.Body),
				Tags:   doc.Frontmatter.Tags,
			}
			if doc.Frontmatter.Category != "" {
				body.Category = doc.Frontmatter.Category
			}
			if doc.Frontmatter.WIP != nil {
				body.WIP = doc.Frontmatter.WIP
			} else {
				t := true
				body.WIP = &t
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
				return cliError("esa-mini create", err.Error(), "Run 'esa-mini token set' or set ESA_ACCESS_TOKEN.")
			}

			post, err := client.CreatePost(team, body)
			if err != nil {
				return cliError("esa-mini create", formatAPIError(err), "Check your input and permissions.")
			}

			if !noWriteBack {
				if err := writeBackPost(file, team, post); err != nil {
					return cliError("esa-mini create", err.Error(), "Check the file path is writable.")
				}
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Created: #%d\nTitle:   %s\nURL:     %s\n", post.Number, post.Name, post.URL)
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
	cmd.Flags().BoolVar(&noWriteBack, "no-write-back", false, "Do not write server response (number, url, updated_at, etc.) back to the input file")

	return cmd
}

func readDocument(path string) (*frontmatter.Document, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()
	return frontmatter.Parse(f)
}
