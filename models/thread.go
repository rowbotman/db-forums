package models

import "time"

type Thread struct {
	Author string `json:"author"`
	Slug *string `json:"slug"`
	Votes int `json:"votes"`
	Title string `json:"title"`
	Created time.Time `json:"created"`
	ForumName string `json:"forum"`
	Id int `json:"id"`
	Message string `json:"message"`
}

type ThreadResult struct {
	Author string `json:"author"`
	Title string `json:"title"`
	Created time.Time `json:"created"`
	ForumName string `json:"forum"`
	Id int `json:"id"`
	Message string `json:"message"`
}

type ThreadUpdate struct {
	Message string `json:"message"`
	Title string `json:"title"`
}