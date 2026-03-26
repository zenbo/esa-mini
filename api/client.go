package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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
