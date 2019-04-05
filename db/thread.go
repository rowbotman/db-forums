package db

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/lib/pq"
	"strconv"
)


type Thread struct {
	Uid     int            `json:"uid, omitempty"`
	Title   sql.NullString `json:"title, omitempty"`
	UserId  int            `json:"userId, omitempty"`
	ForumId int            `json:"forumId, omitempty"`
	Message sql.NullString `json:"message, omitempty"`
	Votes   int            `json:"votes, omitempty"`
	Slug    sql.NullString `json:"slug, omitempty"`
	Created pq.NullTime    `json:"created, omitempty"`
}

type ThreadInfo struct {
	Uid     int            `json:"id, omitempty"`
	Title   sql.NullString `json:"title, omitempty"`
	Author  NullString     `json:"author, omitempty"`
	Forum   NullString     `json:"forum, omitempty"`
	Message NullString     `json:"message, omitempty"`
	Votes   int            `json:"votes, omitempty"`
	Slug    NullString     `json:"slug, omitempty"`
	Created NullTime       `json:"created, omitempty"`
}

type VoteInfo struct {
	Nickname string `json:"nickname,omitempty"`
	Voice    int    `json:"voice, omitempty"`
}


func isThreadExist(slugOrId string) bool {
	sqlStatement := `SELECT uid FROM thread WHERE slug = $1`
	row := DB.QueryRow(sqlStatement, slugOrId)
	id := 0
	err := row.Scan(&id)
	if err != nil || id < 0 {
		fmt.Println("Thread didn`t exist")
		return false
	}

	return true
}

func InsertNewPosts(slugOrId string, posts []Post) ([]Post, error) {
	if !isThreadExist(slugOrId) {
		return nil, errors.New("thread did not exist")
	}
	sqlStatement := `UPDATE post p SET p.thread_id = $1 WHERE p.parent_id = $2 AND p.author = $3 AND p.message = $4 RETURNING p.uid`
	sqlCheckStatement := `SELECT t.slug FROM thread t INNER JOIN post p ON (t.uid = p.thread_id) WHERE p.uid = $1`
	updPosts := []Post{}
	for _, post := range posts {
		checkRow := DB.QueryRow(sqlCheckStatement, post.ParentId)
		var parentThread string
		err := checkRow.Scan(&parentThread)
		if err == sql.ErrNoRows {
			fmt.Println("No rows were returned!")
			return nil, nil
		} else if err != nil {
			return nil, err
		}
		if parentThread != slugOrId {
			return nil, errors.New("Post does not exist in the tread\n")
		}
		row := DB.QueryRow(sqlStatement, slugOrId, post.ParentId, post.Author, post.Message)
		updPost := Post{}
		err = row.Scan(&updPost.Uid)
		if err == sql.ErrNoRows {
			fmt.Println("No rows were returned!")
			return nil, nil
		} else if err != nil {
			return nil, err
		}

		sqlForPostData := `
  SELECT  p.parent_id, f.title, p.user_id, p.thread_id, p.message, p.is_edited, p.created FROM post p
  JOIN forum f ON (p.forum_id = f.uid)
  WHERE p.uid = $1`
		row = DB.QueryRow(sqlForPostData, updPost.Uid)
		err = row.Scan(&updPost.ParentId,
			&updPost.Forum,
			&updPost.Message,
			&updPost.Author,
			&updPost.ThreadId,
			&updPost.Forum,
			&updPost.IsEdited,
			&updPost.Created)
		if err == sql.ErrNoRows {
			fmt.Println("No rows were returned!")
			return nil, nil
		} else if err != nil {
			return nil, err
		}

		updPosts = append(updPosts, updPost)
	}
	return updPosts, nil
}

func SelectFromThread(slugOrId string, isId bool) (ThreadInfo, error) {
	sqlStatement := `SELECT t.uid, t.title, u.nickname, f.title, t.message, t.votes, t.created FROM thread t
	JOIN forum f ON (f.uid = t.forum_id)
	JOIN profile u ON (t.user_id = u.uid) WHERE`
	var row *sql.Row
	if isId {
		sqlStatement += ` t.uid = $1;`
		id, err := strconv.ParseInt(slugOrId, 10, 64)
		if err != nil {
			return ThreadInfo{}, err
		}
		row = DB.QueryRow(sqlStatement, id)
	} else {
		sqlStatement += ` t.slug = $1;`
		row = DB.QueryRow(sqlStatement, slugOrId)
	}

	thread := ThreadInfo{}
	err := row.Scan(
		&thread.Uid,
		&thread.Title,
		&thread.Author,
		&thread.Forum,
		&thread.Message,
		&thread.Votes,
		&thread.Created)
	if err == sql.ErrNoRows {
		fmt.Println("No rows were returned!")
		return ThreadInfo{}, err
	} else if err != nil {
		return ThreadInfo{}, err
	}

	return thread, nil
}

func UpdateThread(slugOrId string, thread *ThreadInfo) error {
	if !isThreadExist(slugOrId) {
		return errors.New("thread did not exist")
	}
	sqlStatement := `UPDATE thread t SET t.title = $1, t.message = $2 WHERE t.slug = $3`
	_, err := DB.Exec(sqlStatement, thread.Title, thread.Message, slugOrId)
	if err != nil {
		return err
	}

	*thread, err = SelectFromThread(slugOrId, false)
	if err != nil {
		return err
	}

	return nil
}


func SelectThreadPosts(slugOrid string, limit int32, since int64, sort string, desc bool) ([]Post, error) {
	sqlStatement := `SELECT p.uid, p.parent_id, u.nickname, p.message, p.is_edited, f.title, p.thread_id, p.created
	FROM post p JOIN forum ON (f.uid = p.forum_id)
	JOIN profile u ON (u.uid = p.author_id)
	JOIN thread t ON (t.uid = p.thread_id)`

	if since > 0 {
		sqlStatement += ` AND p.path && array[:p.uid]`
	}

	switch sort {
	case "tree": {
		if desc {
			sqlStatement += ` WHERE t.slug = $1 AND p.path < ` // ` AND p.path && array[:$2]`
		} else {
			sqlStatement += ` WHERE t.slug = $1 AND p.path > ` // ` AND p.path && array[$2:]`
		}
		sqlStatement += `(SELECT path FROM post WHERE uid = $2)` // не нужно, если /\ сработает
		if desc {
			sqlStatement += `ORDER BY p.path DESC`
		} else {
			sqlStatement += `ORDER BY p.path ASC`
		}
		sqlStatement += ` LIMIT $3;`
	}
	case "parent_tree": {
		sqlStatement += `WHERE path[1] IN (
		SELECT pst.uid FROM post pst
		JOIN thread td ON (pst.thread_id = td.uid)
		WHERE td.slug = $1 AND pst.parent_id = 0 `
		if desc {
			sqlStatement += ` AND pst.uid < `
		} else {
			sqlStatement += ` AND pst.uid > `
		}
		sqlStatement += `SELECT path[1] FROM post WHERE uid = $2`
		if desc {
			sqlStatement += ` ORDER BY pst.uid DESC LIMIT $3)`
		} else {
			sqlStatement += ` ORDER BY pst.uid ASC  LIMIT $3)`
		}
		if desc {
			sqlStatement += `ORDER BY path[1] DESC, path;`
		} else {
			sqlStatement += `ORDER BY path;`
		}

	}
	default: {
		if desc {
			sqlStatement += ` WHERE t.slug = $1 AND p.uid > $2 ORDER BY p.uid DESC`
		} else {
			sqlStatement += ` WHERE t.slug = $1 AND p.uid < $2 ORDER BY p.uid ASC`
		}
		sqlStatement += ` LIMIT $3;`

	}
	}
	rows, err := DB.Query(sqlStatement, slugOrid, since, limit)
	fmt.Println(sqlStatement)
	if err != nil {
		fmt.Println("thread haven`t posts")
		return nil, err
	}
	defer rows.Close()
	posts := []Post{}
	for rows.Next() {
		newPost := Post{}
		err = rows.Scan(
			&newPost.Uid,
			&newPost.ParentId,
			&newPost.Author,
			&newPost.Message,
			&newPost.IsEdited,
			&newPost.Forum,
			&newPost.Created)
		if err != nil {
			fmt.Println("posts select error")
			return nil, err
		}
		posts = append(posts, newPost)
	}
	// get any error encountered during iteration
	err = rows.Err()
	if err != nil {
		fmt.Println("error occurred in select posts")
		return nil, err
	}
	return posts, nil
}

func UpdateVote(slugOrId string, vote VoteInfo) (ThreadInfo, error) {
	thread, err := SelectFromThread(slugOrId, false)
	if err != nil {
		fmt.Println("thread did not exist")
		return ThreadInfo{}, err
	}
	sqlStatement := `
	UPDATE votes SET user_id   = (SELECT p.uid FROM profile p WHERE p.nickname = $1),
	                 thread_id = (SELECT t.uid FROM thread  t WHERE t.slug     = $2),
	                 value     = $3;`

	_, err  = DB.Exec(sqlStatement, vote.Nickname, slugOrId, vote.Voice)
	if err != nil {
		return ThreadInfo{}, err
	}
	thread.Votes += vote.Voice

	return thread, nil
}

