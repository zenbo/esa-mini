package cmd

import (
	"github.com/spf13/cobra"
)

func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "esa-mini",
		Short:         "Minimal CLI for esa.io post operations",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	cmd.AddCommand(newTeamsCmd())
	cmd.AddCommand(newGetCmd())
	cmd.AddCommand(newCreateCmd())
	cmd.AddCommand(newUpdateCmd())

	return cmd
}
