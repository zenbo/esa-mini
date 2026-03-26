package cmd

import (
	"bufio"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/zenbo/esa-mini/token"
)

func newTokenCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "token",
		Short: "Manage access token",
	}
	cmd.AddCommand(newTokenSetCmd())
	cmd.AddCommand(newTokenShowCmd())
	cmd.AddCommand(newTokenDeleteCmd())
	return cmd
}

func newTokenSetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "set",
		Short: "Save access token to ~/.config/esa-mini/token",
		RunE: func(cmd *cobra.Command, args []string) error {
			_, _ = fmt.Fprint(cmd.ErrOrStderr(), "Token: ")
			scanner := bufio.NewScanner(cmd.InOrStdin())
			if !scanner.Scan() {
				return cliError("esa-mini token set", "failed to read token", "Provide a token via stdin.")
			}
			tok := strings.TrimSpace(scanner.Text())
			if tok == "" {
				return cliError("esa-mini token set", "empty token", "Provide a non-empty token.")
			}
			if err := token.Save(tok); err != nil {
				return cliError("esa-mini token set", err.Error(), "Check file permissions on ~/.config/esa-mini/.")
			}
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Token saved.")
			return nil
		},
	}
}

func newTokenShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show",
		Short: "Show saved access token (masked)",
		RunE: func(cmd *cobra.Command, args []string) error {
			tok, err := token.Load()
			if err != nil {
				return cliError("esa-mini token show", err.Error(), "Check file permissions on ~/.config/esa-mini/.")
			}
			if tok == "" {
				return cliError("esa-mini token show", "no saved token found", "Run 'esa-mini token set' first.")
			}
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), mask(tok))
			return nil
		},
	}
}

func newTokenDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete",
		Short: "Delete saved access token",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := token.Delete(); err != nil {
				return cliError("esa-mini token delete", err.Error(), "Run 'esa-mini token show' to check if a token exists.")
			}
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Token deleted.")
			return nil
		},
	}
}

func mask(tok string) string {
	if len(tok) <= 8 {
		return strings.Repeat("*", len(tok))
	}
	return tok[:4] + strings.Repeat("*", len(tok)-8) + tok[len(tok)-4:]
}
