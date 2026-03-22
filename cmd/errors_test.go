package cmd

import (
	"testing"

	"github.com/zenbo/esa-mini/api"
)

func TestCliError(t *testing.T) {
	err := cliError("esa-mini get", "404 Not Found", "Check the post number.")
	expected := "Error: esa-mini get failed\nWhy:   404 Not Found\nHint:  Check the post number."
	if err.Error() != expected {
		t.Errorf("got:\n%s\nwant:\n%s", err.Error(), expected)
	}
}

func TestFormatAPIError(t *testing.T) {
	t.Run("api error", func(t *testing.T) {
		err := &api.APIError{
			StatusCode: 404,
			Status:     "404 Not Found",
			Body:       `{"error":"not_found"}`,
		}
		result := formatAPIError(err)
		if result != `404 {"error":"not_found"}` {
			t.Errorf("result = %q", result)
		}
	})

	t.Run("generic error", func(t *testing.T) {
		err := &net_error{msg: "connection refused"}
		result := formatAPIError(err)
		if result != "connection refused" {
			t.Errorf("result = %q", result)
		}
	})
}

type net_error struct{ msg string }

func (e *net_error) Error() string { return e.msg }
