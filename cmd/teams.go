package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/zenbo/esa-mini/api"
)

func newTeamsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "teams",
		Short: "List teams",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := api.NewClient()
			if err != nil {
				return cliError("esa-mini teams", err.Error(), "Run 'esa-mini token set' or set ESA_ACCESS_TOKEN.")
			}
			teams, err := client.GetTeams()
			if err != nil {
				return cliError("esa-mini teams", formatAPIError(err), "Check your token and network connection.")
			}
			for _, t := range teams {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "%s\t%s\n", t.Name, t.URL)
			}
			return nil
		},
	}
}
