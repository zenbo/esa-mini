package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newTestClient(t *testing.T, handler http.Handler) *Client {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	return &Client{
		token:      "test-token",
		httpClient: server.Client(),
		baseURL:    server.URL,
	}
}

func TestNewClient(t *testing.T) {
	t.Run("missing token", func(t *testing.T) {
		t.Setenv("ESA_ACCESS_TOKEN", "")
		_, err := NewClient()
		if err == nil {
			t.Fatal("expected error for missing token")
		}
	})

	t.Run("valid token", func(t *testing.T) {
		t.Setenv("ESA_ACCESS_TOKEN", "my-token")
		c, err := NewClient()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if c.token != "my-token" {
			t.Errorf("token = %q, want %q", c.token, "my-token")
		}
	})
}

func TestGetTeams(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/teams" {
			t.Errorf("path = %q, want /v1/teams", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Errorf("auth header = %q", r.Header.Get("Authorization"))
		}
		resp := TeamsResponse{
			Teams: []Team{
				{Name: "docs", URL: "https://docs.esa.io"},
				{Name: "dev", URL: "https://dev.esa.io"},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			t.Fatal(err)
		}
	})

	client := newTestClient(t, handler)
	teams, err := client.GetTeams()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(teams) != 2 {
		t.Fatalf("len(teams) = %d, want 2", len(teams))
	}
	if teams[0].Name != "docs" {
		t.Errorf("teams[0].Name = %q", teams[0].Name)
	}
	if teams[1].Name != "dev" {
		t.Errorf("teams[1].Name = %q", teams[1].Name)
	}
}

func TestGetPost(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/teams/docs/posts/123" {
			t.Errorf("path = %q", r.URL.Path)
		}
		if r.Method != http.MethodGet {
			t.Errorf("method = %q, want GET", r.Method)
		}
		post := Post{
			Number:    123,
			Name:      "テスト記事",
			BodyMd:    "# Hello",
			URL:       "https://docs.esa.io/posts/123",
			WIP:       false,
			Tags:      []string{"go"},
			Category:  "dev/tips",
			UpdatedAt: "2025-07-01T12:00:00+09:00",
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(post); err != nil {
			t.Fatal(err)
		}
	})

	client := newTestClient(t, handler)
	post, err := client.GetPost("docs", 123)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if post.Number != 123 {
		t.Errorf("Number = %d", post.Number)
	}
	if post.Name != "テスト記事" {
		t.Errorf("Name = %q", post.Name)
	}
	if post.BodyMd != "# Hello" {
		t.Errorf("BodyMd = %q", post.BodyMd)
	}
}

func TestCreatePost(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/teams/docs/posts" {
			t.Errorf("path = %q", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Errorf("method = %q, want POST", r.Method)
		}

		var req CreatePostRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}
		if req.Post.Name != "新しい記事" {
			t.Errorf("Name = %q", req.Post.Name)
		}
		if req.Post.BodyMd != "本文" {
			t.Errorf("BodyMd = %q", req.Post.BodyMd)
		}

		resp := Post{
			Number: 456,
			Name:   "新しい記事",
			URL:    "https://docs.esa.io/posts/456",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			t.Fatal(err)
		}
	})

	wip := true
	client := newTestClient(t, handler)
	post, err := client.CreatePost("docs", PostBody{
		Name:   "新しい記事",
		BodyMd: "本文",
		WIP:    &wip,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if post.Number != 456 {
		t.Errorf("Number = %d", post.Number)
	}
}

func TestUpdatePost(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/teams/docs/posts/123" {
			t.Errorf("path = %q", r.URL.Path)
		}
		if r.Method != http.MethodPatch {
			t.Errorf("method = %q, want PATCH", r.Method)
		}

		resp := Post{
			Number: 123,
			Name:   "更新された記事",
			URL:    "https://docs.esa.io/posts/123",
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			t.Fatal(err)
		}
	})

	client := newTestClient(t, handler)
	post, err := client.UpdatePost("docs", 123, UpdatePostBody{
		Name:   "更新された記事",
		BodyMd: "更新本文",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if post.Name != "更新された記事" {
		t.Errorf("Name = %q", post.Name)
	}
}

func TestAPIError(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error":"not_found","message":"Not Found"}`))
	})

	client := newTestClient(t, handler)
	_, err := client.GetPost("docs", 999)
	if err == nil {
		t.Fatal("expected error")
	}

	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected *APIError, got %T", err)
	}
	if apiErr.StatusCode != 404 {
		t.Errorf("StatusCode = %d, want 404", apiErr.StatusCode)
	}
}
