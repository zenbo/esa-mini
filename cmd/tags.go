package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/zenbo/esa-mini/api"
	"github.com/zenbo/esa-mini/token"
)

func newTagsCmd() *cobra.Command {
	var match string

	cmd := &cobra.Command{
		Use:   "tags [team]",
		Short: "List tags",
		Long: `List tags.
Team can be omitted if ESA_TEAM is set or a default team is saved via 'esa-mini team set'.`,
		Args: cobra.RangeArgs(0, 1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var team string
			if len(args) >= 1 {
				team = args[0]
			} else {
				resolved, err := token.ResolveTeam()
				if err != nil {
					return cliError("esa-mini tags", err.Error(), "Check config permissions.")
				}
				if resolved == "" {
					return cliError("esa-mini tags", "team is required", "Specify team as argument, set ESA_TEAM, or run 'esa-mini team set'.")
				}
				team = resolved
			}

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
