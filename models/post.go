package models

import "time"

type Post struct {
	Author string `json:"author"`
	Created time.Time `json:"created"`
	ForumName string `json:"forum"`
	Id int `json:"id"`
	IsEdited bool `json:"isEdited"`
	Message string `json:"message"`
	Parent int `json:"parent"`
	Tid int `json:"thread"`
}

type DetailedInfo struct {
	PostInfo Post `json:"post"`
	AuthorInfo *User `json:"author,omitempty"`
	ThreadInfo *Thread `json:"thread,omitempty"`
	ForumInfo *Forum `json:"forum,omitempty"`
}