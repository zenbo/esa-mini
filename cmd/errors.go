package cmd

import (
	"errors"
	"fmt"

	"github.com/zenbo/esa-mini/api"
)

type cliErr struct {
	command string
	why     string
	hint    string
}

func (e *cliErr) Error() string {
	return fmt.Sprintf("Error: %s failed\nWhy:   %s\nHint:  %s", e.command, e.why, e.hint)
}

func cliError(command, why, hint string) error {
	return &cliErr{command: command, why: why, hint: hint}
}

func formatAPIError(err error) string {
	var apiErr *api.APIError
	if errors.As(err, &apiErr) {
		return fmt.Sprintf("%d %s", apiErr.StatusCode, apiErr.Body)
	}
	return err.Error()
}
