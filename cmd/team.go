package cmd

import (
	"bufio"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/zenbo/esa-mini/api"
	"github.com/zenbo/esa-mini/token"
)

func newTeamCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "team",
		Short: "Manage default team",
	}
	cmd.AddCommand(newTeamSetCmd())
	cmd.AddCommand(newTeamShowCmd())
	cmd.AddCommand(newTeamDeleteCmd())
	return cmd
}

func newTeamSetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "set [team]",
		Short: "Save default team to ~/.config/esa-mini/team",
		Long: `Save default team to ~/.config/esa-mini/team.

If a team name is given as an argument, it is saved directly.
Otherwise, fetches your teams from the API.
If you belong to one team, it is saved automatically.
If you belong to multiple teams, you will be prompted to choose one.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var name string

			if len(args) == 1 {
				name = args[0]
			} else {
				client, err := api.NewClient()
				if err != nil {
					return cliError("esa-mini team set", err.Error(), "Run 'esa-mini token set' or set ESA_ACCESS_TOKEN.")
				}

				teams, err := client.GetTeams()
				if err != nil {
					return cliError("esa-mini team set", formatAPIError(err), "Check your token and network connection.")
				}

				if len(teams) == 0 {
					return cliError("esa-mini team set", "no teams found", "You don't belong to any team.")
				}

				if len(teams) == 1 {
					name = teams[0].Name
				} else {
					name, err = promptTeamSelection(cmd, teams)
					if err != nil {
						return err
					}
				}
			}

			if err := token.SaveTeam(name); err != nil {
				return cliError("esa-mini team set", err.Error(), "Check file permissions on ~/.config/esa-mini/.")
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Team saved: %s (%s.esa.io)\n", name, name)
			return nil
		},
	}
}

func promptTeamSelection(cmd *cobra.Command, teams []api.Team) (string, error) {
	_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "Select a team:")
	for i, t := range teams {
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "  %d) %s (%s.esa.io)\n", i+1, t.Name, t.Name)
	}
	_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Number [1-%d]: ", len(teams))

	scanner := bufio.NewScanner(cmd.InOrStdin())
	if !scanner.Scan() {
		return "", cliError("esa-mini team set", "failed to read input", "Enter a number to select a team.")
	}
	input := strings.TrimSpace(scanner.Text())

	var choice int
	if _, err := fmt.Sscanf(input, "%d", &choice); err != nil || choice < 1 || choice > len(teams) {
		return "", cliError("esa-mini team set", fmt.Sprintf("invalid selection: %s", input), fmt.Sprintf("Enter a number between 1 and %d.", len(teams)))
	}
	return teams[choice-1].Name, nil
}

func newTeamShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show",
		Short: "Show saved default team",
		RunE: func(cmd *cobra.Command, args []string) error {
			name, err := token.LoadTeam()
			if err != nil {
				return cliError("esa-mini team show", err.Error(), "Check file permissions on ~/.config/esa-mini/.")
			}
			if name == "" {
				return cliError("esa-mini team show", "no saved team found", "Run 'esa-mini team set' first.")
			}
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), name)
			return nil
		},
	}
}

func newTeamDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete",
		Short: "Delete saved default team",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := token.DeleteTeam(); err != nil {
				return cliError("esa-mini team delete", err.Error(), "Run 'esa-mini team show' to check if a team is saved.")
			}
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Team deleted.")
			return nil
		},
	}
}
