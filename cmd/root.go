package cmd

import (
	"runtime/debug"

	"github.com/spf13/cobra"
)

var version = "dev"

func resolveVersion() string {
	if version != "dev" {
		return version
	}
	if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" && info.Main.Version != "(devel)" {
		return info.Main.Version
	}
	return version
}

func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "esa-mini",
		Short:         "Minimal CLI for esa.io post operations",
		Version:       resolveVersion(),
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	cmd.AddCommand(newTeamsCmd())
	cmd.AddCommand(newGetCmd())
	cmd.AddCommand(newCreateCmd())
	cmd.AddCommand(newUpdateCmd())

	return cmd
}
