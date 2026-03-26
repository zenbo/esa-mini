package cmd

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/zenbo/esa-mini/api"
	"github.com/zenbo/esa-mini/token"
)

func newSearchCmd() *cobra.Command {
	var (
		query     string
		author    string
		updatedBy string
		watchedBy string
		category  string
		tag       string
		wip       string
		sort      string
		order     string
		limit     int
		output    string
	)

	cmd := &cobra.Command{
		Use:   "search <team>",
		Short: "Search posts and optionally save to files",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			team := args[0]

			client, err := api.NewClient()
			if err != nil {
				return cliError("esa-mini search", err.Error(), "Run 'esa-mini token set' or set ESA_ACCESS_TOKEN.")
			}

			// @me を screen_name に解決
			screenName, err := resolveScreenName(client, author, updatedBy, watchedBy)
			if err != nil {
				return cliError("esa-mini search", err.Error(), "Failed to resolve screen_name.")
			}
			if author == "@me" {
				author = screenName
			}
			if updatedBy == "@me" {
				updatedBy = screenName
			}
			if watchedBy == "@me" {
				watchedBy = screenName
			}

			q := buildQuery(query, author, updatedBy, watchedBy, category, tag, wip)

			// --output ありで検索条件なしならエラー
			if output != "" && q == "" {
				return cliError("esa-mini search", "--output requires at least one search condition",
					"Add --query, --author, --category, --tag, or other search flags.")
			}

			// ページネーションで全ページ取得
			var allPosts []api.Post
			totalCount := 0
			page := 1

			for {
				params := api.SearchParams{
					Q:       q,
					Sort:    sort,
					Order:   order,
					Page:    page,
					PerPage: 100,
				}
				resp, err := client.SearchPosts(team, params)
				if err != nil {
					var apiErr *api.APIError
					if errors.As(err, &apiErr) && apiErr.StatusCode == 429 {
						_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Rate limit reached. Stopped at %d of %d.\n", len(allPosts), totalCount)
						break
					}
					return cliError("esa-mini search", formatAPIError(err), "Check the team name and search conditions.")
				}

				if page == 1 {
					totalCount = resp.TotalCount
					if output != "" && totalCount > limit {
						_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "%d posts matched. Downloading %d of %d...\n", totalCount, limit, totalCount)
					}
				}

				allPosts = append(allPosts, resp.Posts...)

				if len(allPosts) >= limit || resp.NextPage == nil {
					break
				}
				page = *resp.NextPage
			}

			// limit で切り詰め
			if len(allPosts) > limit {
				allPosts = allPosts[:limit]
			}

			if output != "" {
				return saveSearchResults(cmd, output, team, allPosts)
			}

			return printSearchResults(cmd, allPosts, totalCount)
		},
	}

	cmd.Flags().StringVarP(&query, "query", "q", "", "Raw search query")
	cmd.Flags().StringVar(&author, "author", "", "Author screen_name (empty value = yourself)")
	cmd.Flags().StringVar(&updatedBy, "updated-by", "", "Last updater screen_name (empty value = yourself)")
	cmd.Flags().StringVar(&watchedBy, "watched-by", "", "Watcher screen_name (empty value = yourself)")
	cmd.Flags().StringVar(&category, "category", "", "Category prefix match")
	cmd.Flags().StringVarP(&tag, "tag", "t", "", "Tag")
	cmd.Flags().StringVar(&wip, "wip", "", "WIP status (true/false)")
	cmd.Flags().StringVarP(&sort, "sort", "s", "updated", "Sort field")
	cmd.Flags().StringVar(&order, "order", "desc", "Sort order (asc/desc)")
	cmd.Flags().IntVarP(&limit, "limit", "l", 100, "Maximum number of posts to fetch")
	cmd.Flags().StringVarP(&output, "output", "o", "", "Output directory to save posts")

	// NoOptDefVal: フラグが値なしで指定された場合のデフォルト値
	cmd.Flags().Lookup("author").NoOptDefVal = "@me"
	cmd.Flags().Lookup("updated-by").NoOptDefVal = "@me"
	cmd.Flags().Lookup("watched-by").NoOptDefVal = "@me"

	return cmd
}

func buildQuery(query, author, updatedBy, watchedBy, category, tag, wip string) string {
	var parts []string
	if query != "" {
		parts = append(parts, query)
	}
	if author != "" {
		parts = append(parts, "@"+author)
	}
	if updatedBy != "" {
		parts = append(parts, "updated_by:"+updatedBy)
	}
	if watchedBy != "" {
		parts = append(parts, "watched_by:"+watchedBy)
	}
	if category != "" {
		parts = append(parts, "in:"+category)
	}
	if tag != "" {
		parts = append(parts, "#"+tag)
	}
	if wip != "" {
		parts = append(parts, "wip:"+wip)
	}
	return strings.Join(parts, " ")
}

func resolveScreenName(client *api.Client, flags ...string) (string, error) {
	needResolve := false
	for _, f := range flags {
		if f == "@me" {
			needResolve = true
			break
		}
	}
	if !needResolve {
		return "", nil
	}

	// キャッシュから読み込み
	cached, err := token.LoadScreenName()
	if err != nil {
		return "", err
	}
	if cached != "" {
		return cached, nil
	}

	// API で取得してキャッシュ
	user, err := client.GetUser()
	if err != nil {
		return "", fmt.Errorf("failed to get user info: %w", err)
	}
	if err := token.SaveScreenName(user.ScreenName); err != nil {
		return "", fmt.Errorf("failed to cache screen_name: %w", err)
	}
	return user.ScreenName, nil
}

func printSearchResults(cmd *cobra.Command, posts []api.Post, totalCount int) error {
	for _, p := range posts {
		date := p.UpdatedAt
		if len(date) >= 10 {
			date = date[:10]
		}
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "%-6d %-40s %s\n", p.Number, p.Category+"/"+p.Name, date)
	}

	pages := (totalCount + 99) / 100
	if pages == 0 {
		pages = 1
	}
	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "-- %d posts found (page 1/%d) --\n", totalCount, pages)
	return nil
}

func saveSearchResults(cmd *cobra.Command, outputDir, team string, posts []api.Post) error {
	for _, p := range posts {
		path, err := savePost(outputDir, team, &p)
		if err != nil {
			return err
		}
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Saved: %s\n", path)
	}
	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "-- %d posts saved --\n", len(posts))
	return nil
}
