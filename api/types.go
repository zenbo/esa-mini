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

type APIError struct {
	StatusCode int
	Status     string
	Body       string
}

func (e *APIError) Error() string {
	return e.Status
}
