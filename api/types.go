package api

type Team struct {
	Name        string `json:"name"`
	URL         string `json:"url"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
}

type TeamsResponse struct {
	Teams []Team `json:"teams"`
}

type Post struct {
	Number    int      `json:"number"`
	Name      string   `json:"name"`
	FullName  string   `json:"full_name"`
	BodyMd    string   `json:"body_md"`
	URL       string   `json:"url"`
	WIP       bool     `json:"wip"`
	Tags      []string `json:"tags"`
	Category  string   `json:"category"`
	UpdatedAt string   `json:"updated_at"`
}

type CreatePostRequest struct {
	Post PostBody `json:"post"`
}

type UpdatePostRequest struct {
	Post UpdatePostBody `json:"post"`
}

type PostBody struct {
	Name     string   `json:"name"`
	BodyMd   string   `json:"body_md"`
	Tags     []string `json:"tags,omitempty"`
	Category string   `json:"category,omitempty"`
	WIP      *bool    `json:"wip,omitempty"`
	Message  string   `json:"message,omitempty"`
}

type UpdatePostBody struct {
	Name     string   `json:"name,omitempty"`
	BodyMd   string   `json:"body_md,omitempty"`
	Tags     []string `json:"tags,omitempty"`
	Category string   `json:"category,omitempty"`
	WIP      *bool    `json:"wip,omitempty"`
	Message  string   `json:"message,omitempty"`
}

type PostsResponse struct {
	Posts      []Post `json:"posts"`
	PrevPage   *int   `json:"prev_page"`
	NextPage   *int   `json:"next_page"`
	TotalCount int    `json:"total_count"`
	Page       int    `json:"page"`
	PerPage    int    `json:"per_page"`
	MaxPerPage int    `json:"max_per_page"`
}

type User struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	ScreenName string `json:"screen_name"`
	Icon       string `json:"icon"`
}

type SearchParams struct {
	Q       string
	Sort    string
	Order   string
	Page    int
	PerPage int
}

type APIError struct {
	StatusCode int
	Status     string
	Body       string
}

func (e *APIError) Error() string {
	return e.Status
}

// GET /v1/teams/:team_name/categories/paths
type CategoryPath struct {
	Path  *string `json:"path"` // null = 未分類
	Posts int     `json:"posts"`
}

type CategoriesPathsResponse struct {
	Categories []CategoryPath `json:"categories"`
	TotalCount int            `json:"total_count"`
	Page       int            `json:"page"`
	PerPage    int            `json:"per_page"`
	PrevPage   *int           `json:"prev_page"`
	NextPage   *int           `json:"next_page"`
}

// GET /v1/teams/:team_name/categories/top
type TopCategory struct {
	Name     string `json:"name"`
	FullName string `json:"full_name"`
	Count    int    `json:"count"`
	HasChild bool   `json:"has_child"`
}

type CategoriesTopResponse struct {
	Categories []TopCategory `json:"categories"`
	TotalCount int           `json:"total_count"`
	Page       int           `json:"page"`
	PerPage    int           `json:"per_page"`
	PrevPage   *int          `json:"prev_page"`
	NextPage   *int          `json:"next_page"`
}

// GET /v1/teams/:team_name/tags
type Tag struct {
	Name       string `json:"name"`
	PostsCount int    `json:"posts_count"`
}

type TagsResponse struct {
	Tags       []Tag `json:"tags"`
	TotalCount int   `json:"total_count"`
	Page       int   `json:"page"`
	PerPage    int   `json:"per_page"`
	PrevPage   *int  `json:"prev_page"`
	NextPage   *int  `json:"next_page"`
}
