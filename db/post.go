package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"
	"time"
)

type Post struct {
	Uid      int64     `json:"id,omitempty"`
	ParentId int       `json:"parent,omitempty"`
	Author   string    `json:"author,omitempty"`
	Message  string    `json:"message,omitempty"`
	Forum    string    `json:"forum,omitempty"`
	ThreadId int64     `json:"thread,omitempty"`
	IsEdited bool      `json:"isEdited,omitempty"`
	Created  time.Time `json:"created, omitempty"`
}


type DataForUpdPost struct {
	Id      int64  `json:"id"`
	Message string `json:"message"`
}

func UpdatePost(data DataForUpdPost) (Post, error) {
	sqlStatement := `
  SELECT p.uid, p.parent_id, u.nickname, f.title, p.thread_id, p.created 
  FROM post p JOIN profile u ON (p.author_id = u.uid) JOIN forum f ON (p.forum_id = f.uid)
  WHERE p.uid = $1 GROUP BY p.uid;`
	row := DB.QueryRow(sqlStatement, data.Id)
	post := Post{}
	err := row.Scan(
		&post.Uid,
		&post.ParentId,
		&post.Author,
		&post.Forum,
		&post.ThreadId,
		&post.Created)
	if err == sql.ErrNoRows {
		fmt.Println("No rows were returned!")
		return Post{}, err
	} else if err != nil {
		return Post{}, err
	}

	sqlStatement = `UPDATE post p SET p.message = $1 WHERE p.uid = $2;`
	_, err = DB.Exec(sqlStatement, post.Message, post.Uid)
	if err != nil {
		return Post{}, err
	}

	fmt.Println("New post message")
	post.Message = data.Message
	post.IsEdited = true
	return post, nil
}


func GetPostInfo(postId int64, strArray []string) (map[string][]byte, error) {
	type threadInfo map[string][]byte
	sqlStatement := `
  SELECT p.uid, p.parent_id, u.nickname, p.message, p.is_edited, f.title, p.thread_id, p.created 
  FROM post p JOIN profile u ON (p.author_id = u.uid) JOIN forum f ON (p.forum_id = f.uid)
  WHERE p.uid = $1 GROUP BY p.uid;`
	row := DB.QueryRow(sqlStatement, postId)
	post := Post{}
	err := row.Scan(
		&post.Uid,
		&post.ParentId,
		&post.Author,
		&post.Message,
		&post.IsEdited,
		&post.Forum,
		&post.ThreadId,
		&post.Created)
	if err == sql.ErrNoRows {
		fmt.Println("No rows were returned!")
		return threadInfo{}, err
	} else if err != nil {
		return threadInfo{}, err
	}
	fullInfo := threadInfo{}
	type JsonData []byte
	for _, obj := range strArray {
		sqlStatement = ``
		var jsonStr JsonData
		switch obj {
		case "user": {
			userData, err := SelectUser(post.Author)
			if err != nil {
				return threadInfo{}, err
			}
			jsonStr, _ = json.Marshal(userData)
		}
		case "forum": {
			sqlStatement = `SELECT `
			forumData, err := SelectForumInfo(strconv.Itoa(int(postId)), true)
			if err != nil {
				return threadInfo{}, err
			}
			jsonStr, _ = json.Marshal(forumData)
		}
		case "thread": {
			id := strconv.FormatInt(post.ThreadId, 64)
			threadData, err := SelectFromThread(id, true)
			if err != nil {
				return threadInfo{}, err
			}
			jsonStr, _ = json.Marshal(threadData)
		}
		}
		fullInfo[obj] = jsonStr
	}

	return fullInfo, nil
}