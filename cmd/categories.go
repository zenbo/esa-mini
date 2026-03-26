package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/zenbo/esa-mini/api"
)

func newCategoriesCmd() *cobra.Command {
	var (
		top    bool
		prefix string
		match  string
	)

	cmd := &cobra.Command{
		Use:   "categories <team>",
		Short: "List categories",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if top && (prefix != "" || match != "") {
				return cliError("esa-mini categories", "--top cannot be used with --prefix or --match",
					"Use --top alone or use --prefix/--match without --top.")
			}

			team := args[0]

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
