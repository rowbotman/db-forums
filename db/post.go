package db

import (
	"database/sql"
	"errors"
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
	Created  time.Time `json:"created,omitempty"`
}

type DataForUpdPost struct {
	Id      int64  `json:"id,omitempty"`
	Message string `json:"message,omitempty"`
}

func UpdatePost(data DataForUpdPost) (Post, error) {
	sqlStatement := `
  SELECT p.uid, p.parent_id, p.message, p.is_edited, u.nickname, f.slug, p.thread_id, p.created 
  FROM post p JOIN profile u ON (p.user_id = u.uid) JOIN forum f ON (p.forum_id = f.uid)
  WHERE p.uid = $1 GROUP BY p.uid, u.nickname, f.slug;`
	row := DB.QueryRow(sqlStatement, data.Id)
	fmt.Println(sqlStatement, data.Id)
	post := Post{}
	isEdited := false
	err := row.Scan(
		&post.Uid,
		&post.ParentId,
		&post.Message,
		&isEdited,
		&post.Author,
		&post.Forum,
		&post.ThreadId,
		&post.Created)
	if err == sql.ErrNoRows {
		fmt.Println("No rows were returned!")
		ret := Post{}
		ret.Uid = -1
		return ret, errors.New("Can't find post with id: " + strconv.FormatInt(data.Id, 10))
	} else if err != nil {
		return Post{}, err
	}
	if len(data.Message) > 0 {
		if data.Message != post.Message {
			sqlStatement = `UPDATE post SET message = $1, is_edited = TRUE WHERE uid = $2;`
			_, err = DB.Exec(sqlStatement, data.Message, post.Uid)
			if err != nil {
				return Post{}, err
			}
			fmt.Println("New post message")
			post.Message = data.Message
			fmt.Println(isEdited, post.IsEdited)
			post.IsEdited = true
		}
		//if isEdited {
		//	post.IsEdited = true
		//}
	}
	fmt.Println(post)
	return post, nil
}


func GetPostInfo(postId int64, strArray []string) (map[string]interface{}, error) {
	type threadInfo map[string]interface{}
	sqlStatement := `
  SELECT p.uid, p.parent_id, u.nickname, p.message, p.is_edited, f.slug, f.uid, p.thread_id, p.created 
  FROM post p JOIN profile u ON (p.user_id = u.uid) JOIN forum f ON (p.forum_id = f.uid)
  WHERE p.uid = $1 GROUP BY p.uid, p.parent_id, u.nickname, f.slug, f.uid;`
	row := DB.QueryRow(sqlStatement, postId)
	fmt.Println(sqlStatement, postId)
	post := Post{}
	forumId := int(0)
	//isEdited := false
	err := row.Scan(
		&post.Uid,
		&post.ParentId,
		&post.Author,
		&post.Message,
		&post.IsEdited,
		&post.Forum,
		&forumId,
		&post.ThreadId,
		&post.Created)
	fullInfo := threadInfo{}
	if err == sql.ErrNoRows {
		fmt.Println("No rows were returned!")
		fullInfo["err"] = true
		return fullInfo, errors.New("Can't find post with id: " + strconv.FormatInt(postId, 10))
	} else if err != nil {
		return threadInfo{}, err
	}
	//if isEdited {
	//	post.IsEdited = true
	//}
	fmt.Println(post)
	fullInfo["post"] = post
	for _, obj := range strArray {
		switch obj {
		case "user": {
			userData, err := SelectUser(post.Author)
			if err != nil {
				return threadInfo{}, err
			}
			fullInfo["author"] = userData
		}
		case "forum": {
			id := strconv.Itoa(int(forumId))
			fmt.Println(id)
			forumData, err := SelectForumInfo(id, true)
			if err != nil {
				fmt.Println(err.Error())
				return threadInfo{}, err
			}
			fullInfo["forum"] = forumData
		}
		case "thread": {
			id := strconv.FormatInt(post.ThreadId, 10)
			threadData, err := SelectFromThread(id, true)
			if err != nil {
				return threadInfo{}, err
			}
			fullInfo["thread"] = threadData
		}
		}
	}

	return fullInfo, nil
}