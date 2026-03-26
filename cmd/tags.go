package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/zenbo/esa-mini/api"
)

func newTagsCmd() *cobra.Command {
	var match string

	cmd := &cobra.Command{
		Use:   "tags <team>",
		Short: "List tags",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			team := args[0]

			client, err := api.NewClient()
			if err != nil {
				return cliError("esa-mini tags", err.Error(), "Run 'esa-mini token set' or set ESA_ACCESS_TOKEN.")
			}

			resp, err := client.GetTags(team, 1, 100)
			if err != nil {
				return cliError("esa-mini tags", formatAPIError(err), "Check the team name and your token.")
			}

			matchLower := strings.ToLower(match)
			displayed := 0
			for _, t := range resp.Tags {
				if match != "" && !strings.Contains(strings.ToLower(t.Name), matchLower) {
					continue
				}
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "%-40s %4d posts\n", t.Name, t.PostsCount)
				displayed++
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "-- %d tags (total %d) --\n", displayed, resp.TotalCount)
			return nil
		},
	}

	cmd.Flags().StringVar(&match, "match", "", "Substring match filter (case-insensitive)")

	return cmd
}
