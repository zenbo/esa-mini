package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/zenbo/esa-mini/api"
	"github.com/zenbo/esa-mini/token"
)

func newCategoriesCmd() *cobra.Command {
	var (
		top    bool
		prefix string
		match  string
	)

	cmd := &cobra.Command{
		Use:   "categories [team]",
		Short: "List categories",
		Long: `List categories.
Team can be omitted if ESA_TEAM is set or a default team is saved via 'esa-mini team set'.`,
		Args: cobra.RangeArgs(0, 1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if top && (prefix != "" || match != "") {
				return cliError("esa-mini categories", "--top cannot be used with --prefix or --match",
					"Use --top alone or use --prefix/--match without --top.")
			}

			var team string
			if len(args) >= 1 {
				team = args[0]
			} else {
				resolved, err := token.ResolveTeam()
				if err != nil {
					return cliError("esa-mini categories", err.Error(), "Check config permissions.")
				}
				if resolved == "" {
					return cliError("esa-mini categories", "team is required", "Specify team as argument, set ESA_TEAM, or run 'esa-mini team set'.")
				}
				team = resolved
			}

			client, err := api.NewClient()
			if err != nil {
				return cliError("esa-mini categories", err.Error(), "Run 'esa-mini token set' or set ESA_ACCESS_TOKEN.")
			}

			if top {
				return printTopCategories(cmd, client, team)
			}
			return printCategoryPaths(cmd, client, team, prefix, match)
		},
	}

	cmd.Flags().BoolVar(&top, "top", false, "Show top-level categories only")
	cmd.Flags().StringVar(&prefix, "prefix", "", "Prefix match filter")
	cmd.Flags().StringVar(&match, "match", "", "Substring match filter")

	return cmd
}

func printTopCategories(cmd *cobra.Command, client *api.Client, team string) error {
	resp, err := client.GetCategoriesTop(team)
	if err != nil {
		return cliError("esa-mini categories", formatAPIError(err), "Check the team name and your token.")
	}

	for _, c := range resp.Categories {
		indicator := ""
		if c.HasChild {
			indicator = " [+]"
		}
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "%-40s %4d posts%s\n", c.FullName, c.Count, indicator)
	}
	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "-- %d categories --\n", len(resp.Categories))
	return nil
}

func printCategoryPaths(cmd *cobra.Command, client *api.Client, team, prefix, match string) error {
	resp, err := client.GetCategoriesPaths(team, 1, 100, prefix, match)
	if err != nil {
		return cliError("esa-mini categories", formatAPIError(err), "Check the team name and your token.")
	}

	for _, c := range resp.Categories {
		path := "(uncategorized)"
		if c.Path != nil {
			path = *c.Path
		}
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "%-40s %4d posts\n", path, c.Posts)
	}
	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "-- %d categories (total %d) --\n", len(resp.Categories), resp.TotalCount)
	return nil
}
