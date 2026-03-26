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
		t.Setenv("HOME", t.TempDir())
		_, err := NewClient()
		if err == nil {
			t.Fatal("expected error for missing token")
		}
	})

	t.Run("valid token from env", func(t *testing.T) {
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

func TestSearchPosts(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/teams/docs/posts" {
			t.Errorf("path = %q", r.URL.Path)
		}
		if r.Method != http.MethodGet {
			t.Errorf("method = %q, want GET", r.Method)
		}
		q := r.URL.Query()
		if q.Get("per_page") != "100" {
			t.Errorf("per_page = %q, want 100", q.Get("per_page"))
		}
		if q.Get("q") != "@higuchi" {
			t.Errorf("q = %q, want @higuchi", q.Get("q"))
		}
		if q.Get("sort") != "updated" {
			t.Errorf("sort = %q, want updated", q.Get("sort"))
		}
		if q.Get("order") != "desc" {
			t.Errorf("order = %q, want desc", q.Get("order"))
		}

		next := 2
		resp := PostsResponse{
			Posts: []Post{
				{Number: 1, Name: "記事1"},
				{Number: 2, Name: "記事2"},
			},
			TotalCount: 5,
			Page:       1,
			PerPage:    100,
			NextPage:   &next,
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			t.Fatal(err)
		}
	})

	client := newTestClient(t, handler)
	resp, err := client.SearchPosts("docs", SearchParams{
		Q:       "@higuchi",
		Sort:    "updated",
		Order:   "desc",
		Page:    1,
		PerPage: 100,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.TotalCount != 5 {
		t.Errorf("TotalCount = %d, want 5", resp.TotalCount)
	}
	if len(resp.Posts) != 2 {
		t.Fatalf("len(Posts) = %d, want 2", len(resp.Posts))
	}
	if resp.NextPage == nil || *resp.NextPage != 2 {
		t.Errorf("NextPage = %v, want 2", resp.NextPage)
	}
}

func TestGetUser(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/user" {
			t.Errorf("path = %q, want /v1/user", r.URL.Path)
		}
		user := User{
			ID:         1,
			Name:       "樋口",
			ScreenName: "higuchi",
			Icon:       "https://example.com/icon.png",
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(user); err != nil {
			t.Fatal(err)
		}
	})

	client := newTestClient(t, handler)
	user, err := client.GetUser()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.ScreenName != "higuchi" {
		t.Errorf("ScreenName = %q, want higuchi", user.ScreenName)
	}
}

func TestGetCategoriesPaths(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/teams/docs/categories/paths" {
			t.Errorf("path = %q", r.URL.Path)
		}
		q := r.URL.Query()
		if q.Get("v") != "2" {
			t.Errorf("v = %q, want 2", q.Get("v"))
		}
		if q.Get("per_page") != "100" {
			t.Errorf("per_page = %q, want 100", q.Get("per_page"))
		}
		if q.Get("prefix") != "dev/" {
			t.Errorf("prefix = %q, want dev/", q.Get("prefix"))
		}

		resp := CategoriesPathsResponse{
			Categories: []CategoryPath{
				{Path: strPtr("dev/tips"), Posts: 12},
				{Path: strPtr("dev/設計"), Posts: 5},
				{Path: nil, Posts: 3},
			},
			TotalCount: 3,
			Page:       1,
			PerPage:    100,
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			t.Fatal(err)
		}
	})

	client := newTestClient(t, handler)
	resp, err := client.GetCategoriesPaths("docs", 1, 100, "dev/", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.TotalCount != 3 {
		t.Errorf("TotalCount = %d, want 3", resp.TotalCount)
	}
	if len(resp.Categories) != 3 {
		t.Fatalf("len(Categories) = %d, want 3", len(resp.Categories))
	}
	if resp.Categories[0].Path == nil || *resp.Categories[0].Path != "dev/tips" {
		t.Errorf("Categories[0].Path = %v", resp.Categories[0].Path)
	}
	if resp.Categories[2].Path != nil {
		t.Errorf("Categories[2].Path = %v, want nil", resp.Categories[2].Path)
	}
}

func TestGetCategoriesTop(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/teams/docs/categories/top" {
			t.Errorf("path = %q", r.URL.Path)
		}

		resp := CategoriesTopResponse{
			Categories: []TopCategory{
				{Name: "dev", FullName: "dev", Count: 15, HasChild: true},
				{Name: "日報", FullName: "日報", Count: 8, HasChild: false},
			},
			TotalCount: 2,
			Page:       1,
			PerPage:    20,
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			t.Fatal(err)
		}
	})

	client := newTestClient(t, handler)
	resp, err := client.GetCategoriesTop("docs")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Categories) != 2 {
		t.Fatalf("len(Categories) = %d, want 2", len(resp.Categories))
	}
	if resp.Categories[0].FullName != "dev" {
		t.Errorf("Categories[0].FullName = %q", resp.Categories[0].FullName)
	}
	if !resp.Categories[0].HasChild {
		t.Errorf("Categories[0].HasChild = false, want true")
	}
	if resp.Categories[1].HasChild {
		t.Errorf("Categories[1].HasChild = true, want false")
	}
}

func TestGetTags(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/teams/docs/tags" {
			t.Errorf("path = %q", r.URL.Path)
		}
		q := r.URL.Query()
		if q.Get("per_page") != "100" {
			t.Errorf("per_page = %q, want 100", q.Get("per_page"))
		}

		resp := TagsResponse{
			Tags: []Tag{
				{Name: "go", PostsCount: 45},
				{Name: "React", PostsCount: 38},
			},
			TotalCount: 2,
			Page:       1,
			PerPage:    100,
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			t.Fatal(err)
		}
	})

	client := newTestClient(t, handler)
	resp, err := client.GetTags("docs", 1, 100)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Tags) != 2 {
		t.Fatalf("len(Tags) = %d, want 2", len(resp.Tags))
	}
	if resp.Tags[0].Name != "go" {
		t.Errorf("Tags[0].Name = %q", resp.Tags[0].Name)
	}
	if resp.Tags[0].PostsCount != 45 {
		t.Errorf("Tags[0].PostsCount = %d", resp.Tags[0].PostsCount)
	}
}

func strPtr(s string) *string {
	return &s
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
