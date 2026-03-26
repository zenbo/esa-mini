package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"

	tkn "github.com/zenbo/esa-mini/token"
)

const defaultBaseURL = "https://api.esa.io"

type Client struct {
	token      string
	httpClient *http.Client
	baseURL    string
}

func NewClient() (*Client, error) {
	tok, err := tkn.Resolve()
	if err != nil {
		return nil, fmt.Errorf("failed to resolve token: %w", err)
	}
	if tok == "" {
		return nil, fmt.Errorf("no access token found")
	}
	base := os.Getenv("ESA_API_BASE_URL")
	if base == "" {
		base = defaultBaseURL
	}
	return &Client{
		token:      tok,
		httpClient: &http.Client{},
		baseURL:    base,
	}, nil
}

func (c *Client) do(method, path string, body io.Reader) ([]byte, error) {
	url := c.baseURL + path
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, &APIError{
			StatusCode: resp.StatusCode,
			Status:     resp.Status,
			Body:       string(data),
		}
	}
	return data, nil
}

func (c *Client) GetTeams() ([]Team, error) {
	data, err := c.do("GET", "/v1/teams", nil)
	if err != nil {
		return nil, err
	}
	var resp TeamsResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}
	return resp.Teams, nil
}

func (c *Client) GetPost(team string, number int) (*Post, error) {
	path := fmt.Sprintf("/v1/teams/%s/posts/%d", team, number)
	data, err := c.do("GET", path, nil)
	if err != nil {
		return nil, err
	}
	var post Post
	if err := json.Unmarshal(data, &post); err != nil {
		return nil, err
	}
	return &post, nil
}

func (c *Client) CreatePost(team string, body PostBody) (*Post, error) {
	reqBody := CreatePostRequest{Post: body}
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}
	path := fmt.Sprintf("/v1/teams/%s/posts", team)
	data, err := c.do("POST", path, bytes.NewReader(jsonData))
	if err != nil {
		return nil, err
	}
	var post Post
	if err := json.Unmarshal(data, &post); err != nil {
		return nil, err
	}
	return &post, nil
}

func (c *Client) SearchPosts(team string, params SearchParams) (*PostsResponse, error) {
	path := fmt.Sprintf("/v1/teams/%s/posts?per_page=%d&page=%d",
		team, params.PerPage, params.Page)
	if params.Q != "" {
		path += "&q=" + url.QueryEscape(params.Q)
	}
	if params.Sort != "" {
		path += "&sort=" + url.QueryEscape(params.Sort)
	}
	if params.Order != "" {
		path += "&order=" + url.QueryEscape(params.Order)
	}
	data, err := c.do("GET", path, nil)
	if err != nil {
		return nil, err
	}
	var resp PostsResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) GetUser() (*User, error) {
	data, err := c.do("GET", "/v1/user", nil)
	if err != nil {
		return nil, err
	}
	var user User
	if err := json.Unmarshal(data, &user); err != nil {
		return nil, err
	}
	return &user, nil
}

func (c *Client) GetCategoriesPaths(team string, page, perPage int, prefix, match string) (*CategoriesPathsResponse, error) {
	path := fmt.Sprintf("/v1/teams/%s/categories/paths?v=2&per_page=%d&page=%d",
		team, perPage, page)
	if prefix != "" {
		path += "&prefix=" + url.QueryEscape(prefix)
	}
	if match != "" {
		path += "&match=" + url.QueryEscape(match)
	}
	data, err := c.do("GET", path, nil)
	if err != nil {
		return nil, err
	}
	var resp CategoriesPathsResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) GetCategoriesTop(team string) (*CategoriesTopResponse, error) {
	path := fmt.Sprintf("/v1/teams/%s/categories/top", team)
	data, err := c.do("GET", path, nil)
	if err != nil {
		return nil, err
	}
	var resp CategoriesTopResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) GetTags(team string, page, perPage int) (*TagsResponse, error) {
	path := fmt.Sprintf("/v1/teams/%s/tags?per_page=%d&page=%d",
		team, perPage, page)
	data, err := c.do("GET", path, nil)
	if err != nil {
		return nil, err
	}
	var resp TagsResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) UpdatePost(team string, number int, body UpdatePostBody) (*Post, error) {
	reqBody := UpdatePostRequest{Post: body}
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}
	path := fmt.Sprintf("/v1/teams/%s/posts/%d", team, number)
	data, err := c.do("PATCH", path, bytes.NewReader(jsonData))
	if err != nil {
		return nil, err
	}
	var post Post
	if err := json.Unmarshal(data, &post); err != nil {
		return nil, err
	}
	return &post, nil
}
