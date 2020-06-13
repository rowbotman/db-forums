package models

type Forum struct {
	Posts   int    `json:"posts"`
	Slug    string `json:"slug"`
	Threads int    `json:"threads"`
	Title   string `json:"title"`
	User    string `json:"user"`
}

type ForumInput struct {
	Slug  string `json:"slug"`
	Title string `json:"title"`
	User  string `json:"user"`
}
