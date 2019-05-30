package db

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"time"
)


type Thread struct {
	Uid     int64     `json:"uid, omitempty"`
	Title   string    `json:"title, omitempty"`
	UserId  int       `json:"userId, omitempty"`
	ForumId int       `json:"forumId, omitempty"`
	Forum   string    `json:"forum, omitempty"`
	Message string    `json:"message, omitempty"`
	Votes   int       `json:"votes, omitempty"`
	Slug    string    `json:"slug, omitempty"`
	Created time.Time `json:"created, omitempty"`
}

type ThreadInfo struct {
	Uid     int64     `json:"id, omitempty"`
	Title   string    `json:"title, omitempty"`
	Author  string    `json:"author, omitempty"`
	Forum   string    `json:"forum, omitempty"`
	Message string    `json:"message, omitempty"`
	Votes   int       `json:"-"`
	Slug    string    `json:"slug, omitempty"`
	Created time.Time `json:"created, omitempty"`
}

type ThreadInfoMin struct {
	Uid     int64       `json:"id, omitempty"`
	Title   string    `json:"title, omitempty"`
	Author  string    `json:"author, omitempty"`
	Forum   string    `json:"forum, omitempty"`
	Message string    `json:"message, omitempty"`
	Created time.Time `json:"created, omitempty"`
}


type VoteInfo struct {
	Nickname string `json:"nickname,omitempty"`
	Voice    int    `json:"voice, omitempty"`
}


func isThreadExist(slugOrId string) (Thread, bool) {
	reqId, err := strconv.ParseInt(slugOrId, 10, 64)
	sqlStatement := `SELECT uid, title, forum_id, "message", slug, user_id, created FROM thread `
	var row *sql.Row
	thread := Thread{}
	if err != nil {
		sqlStatement += `WHERE slug = $1`
		row = DB.QueryRow(sqlStatement, slugOrId)
	} else {
		sqlStatement += `WHERE uid = $1`
		row = DB.QueryRow(sqlStatement, reqId)
	}
	id := int64(0)
	err = row.Scan(
		&thread.Uid,
		&thread.Title,
		&thread.ForumId,
		&thread.Message,
		&thread.Slug,
		&thread.UserId,
		&thread.Created)
	if err != nil || id < 0 {
		fmt.Println("Thread didn`t exist")
		return Thread{}, false
	}

	return thread, true
}

func InsertNewPosts(slugOrId string, posts []Post) ([]Post, error) {
	sqlTime := `SELECT current_timestamp(3);`
	time := time.Time{}
	err := DB.QueryRow(sqlTime).Scan(&time)
	if err != nil {
		return nil, errors.New("error getting current time")
	}

	thread, ok := isThreadExist(slugOrId)
	if !ok {
		return nil, errors.New("thread did not exist")
	}
	fmt.Println("thread", thread.Uid, "exists")
	sqlStatement := `INSERT INTO post VALUES (default, $6, array[0], $1, $2, $3, $4, default, $5) RETURNING uid`

	updPosts := []Post{}
	for _, post := range posts {
		userData, err := SelectUser(post.Author)
		if err != nil {
			fmt.Println("user", post.Author, "did not exist\n Aborting")
			return nil, err
		}
		row := DB.QueryRow(sqlStatement, thread.ForumId, userData.Pk, thread.Uid, post.Message, time, post.ParentId)
		err = row.Scan(&post.Uid)
		if err == sql.ErrNoRows {
			fmt.Println("No rows were returned!")
			return nil, err
		} else if err != nil {
			return nil, err
		}
		fmt.Println(post.ParentId)
		//post.ParentId = 0
		//post.Created = time
		post.ThreadId = thread.Uid

		sqlForPostData := `
  SELECT f.slug, p.created FROM post p
  JOIN forum f ON (p.forum_id = f.uid)
  WHERE p.uid = $1`
		row = DB.QueryRow(sqlForPostData, post.Uid)
		err = row.Scan(
			&post.Forum,
			&post.Created)
		if err == sql.ErrNoRows {
			fmt.Println("No rows were returned!")
			return nil, nil
		} else if err != nil {
			return nil, err
		}

		updPosts = append(updPosts, post)
	}
	return updPosts, nil
}

func SelectFromThread(slugOrId string, isId bool) (ThreadInfo, error) {
	sqlStatement := `SELECT t.uid, t.title, u.nickname, f.slug, t.message, t.votes, t.created, t.slug FROM thread t
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
		&thread.Created,
		&thread.Slug)
	if err == sql.ErrNoRows {
		fmt.Println("No, rows were returned!")
		return ThreadInfo{}, err
	} else if err != nil {
		return ThreadInfo{}, err
	}

	return thread, nil
}

func UpdateThread(slugOrId string, thread *ThreadInfo) error {
	existThread, ok := isThreadExist(slugOrId)
	if !ok {
		return errors.New("thread did not exist")
	}
	sqlStatement := `UPDATE thread SET title = $1, message = $2 WHERE uid = $3`
	_, err := DB.Exec(sqlStatement, thread.Title, thread.Message, existThread.Uid)
	fmt.Println(sqlStatement, existThread.Uid)
	if err != nil {
		return err
	}

	//thread.Uid = existThread.Uid
	//thread.Created = existThread.Created
	//thread.Author = existThread.
	//existThread.Title = thread.Title
	//existThread.Message = thread.Message

	threadInfo, err := SelectFromThread(existThread.Slug, false)
	fmt.Println(threadInfo)
	if err != nil {
		return err
	}
	fmt.Println(threadInfo)
	(*thread).Author = threadInfo.Author
	(*thread).Created = threadInfo.Created
	(*thread).Uid = threadInfo.Uid
	(*thread).Forum = threadInfo.Forum
	(*thread).Slug = threadInfo.Slug

	return nil
}


func SelectThreadPosts(slugOrid string, limit int32, since int64, sort string, desc bool) ([]Post, error) {
	_, err := strconv.ParseInt(slugOrid, 10, 64)
	isId := true
	if err != nil {
		isId = false
	}
	thread, err := SelectFromThread(slugOrid, isId)
	if err != nil {
		return nil, err
	}
	sqlStatement := `SELECT p.uid, p.parent_id, u.nickname,
       p.message, p.is_edited, f.slug,
       p.thread_id, p.created FROM post p
    JOIN forum   AS f ON (f.uid = p.forum_id)
	JOIN profile AS u ON (u.uid = p.user_id)
	JOIN thread  AS t ON (t.uid = p.thread_id)`

	//if since > 0 {
	//	sqlStatement += ` AND p.path && array[:p.uid]`
	//}

	var rows *sql.Rows
	switch sort {
	case "tree": {
		if desc {
			sqlStatement += ` WHERE t.uid = $1 AND p.path `
			if since == 0 {
				sqlStatement += ` >= `
				since = 1
			} else {
				sqlStatement += ` < `
			}
		} else {
			sqlStatement += ` WHERE t.uid = $1 AND p.path `
			if since == 0 {
				sqlStatement += ` >= `
				since = 1
			} else {
				sqlStatement += ` > `
			}
		}

		sqlStatement += `(SELECT path FROM post WHERE uid = $2)`
		if desc {
			sqlStatement += `ORDER BY p.path DESC, p.created DESC`
		} else {
			sqlStatement += `ORDER BY p.path, p.created ASC`
		}
		sqlStatement += ` LIMIT $3;`
		rows, err = DB.Query(sqlStatement, thread.Uid, since, limit)

	}
	case "parent_tree": {
		strLimit := strconv.FormatInt(int64(limit), 10)

		sqlStatement += `WHERE path[1] IN (
		SELECT pst.uid FROM post AS pst
		JOIN thread AS td ON (pst.thread_id = td.uid)
		WHERE td.uid = $1 AND pst.parent_id = 0 `
		if since > 0 {
			//sqlStatement += ` AND pst.uid `
			if desc {
				sqlStatement += ` AND pst.uid < `
			} else {
				sqlStatement += ` AND pst.uid > `
			}
			sqlStatement += ` (SELECT path[1] FROM post WHERE uid = $2) `
		}

		//if desc {
		//	sqlStatement += ` ORDER BY pst.uid DESC LIMIT ` + strLimit + `)`
		//} else {
		//	sqlStatement += ` ORDER BY pst.uid ASC  LIMIT ` + strLimit + `)`
		//}
		//sqlStatement += `ORDER BY pst.uid LIMIT ` + strLimit + `) `
		if desc {
			sqlStatement += `ORDER BY pst.uid DESC LIMIT ` + strLimit + `) ORDER BY p.path[1] DESC, p.path;`
		} else {
			sqlStatement += `ORDER BY pst.uid LIMIT ` + strLimit + `) ORDER BY path;`
		}
		if since > 0 {
			rows, err = DB.Query(sqlStatement, thread.Uid, since)
		} else {
			rows, err = DB.Query(sqlStatement, thread.Uid)
		}

	}
	default: {
		if desc {
			sqlStatement += ` WHERE t.uid = $1 `
			if since > 0 {
				sqlStatement += ` AND p.uid < $2 ORDER BY p.created DESC, p.uid DESC LIMIT $3`
				rows, err = DB.Query(sqlStatement, thread.Uid, since, limit)
			} else {
				sqlStatement += `ORDER BY p.created DESC, p.uid DESC LIMIT $2`
				rows, err = DB.Query(sqlStatement, thread.Uid, limit)
			}
		} else {
			sqlStatement += ` WHERE t.uid = $1 AND p.uid > $2 ORDER BY p.created, p.uid ASC LIMIT $3;`
			//sqlStatement += ` LIMIT $3;`
			rows, err = DB.Query(sqlStatement, thread.Uid, since, limit)
		}

	}
	}
	fmt.Println(sqlStatement, thread.Uid, since, limit)
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
			&newPost.ThreadId,
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
	_, err := strconv.ParseInt(slugOrId, 10, 64)
	thread := ThreadInfo{}
	if err != nil {
		thread, err = SelectFromThread(slugOrId, false)
	} else {
		thread, err = SelectFromThread(slugOrId, true)
	}
	if err != nil {
		fmt.Println("thread did not exist")
		return ThreadInfo{}, errors.New("Can't find thread by slug: " + slugOrId)
	}

	userId := int64(0)
	sqlGetId := `SELECT p.uid FROM profile p WHERE p.nickname = $1;`
	err = DB.QueryRow(sqlGetId, vote.Nickname).Scan(&userId)
	if err != nil {
		return ThreadInfo{}, errors.New("Can't find user by nickname: " + vote.Nickname)
	}

	sqlGetVotes := `SELECT "value", is_edited FROM vote WHERE user_id = $1 AND thread_id = $2;`
	value := int(0)
	isEdited := false
	err = DB.QueryRow(sqlGetVotes, userId, thread.Uid).Scan(&value, &isEdited)
	fmt.Println("user vote:", value)
	if err == sql.ErrNoRows {
		sqlStatement := `
	INSERT INTO vote (user_id, thread_id, "value") VALUES ($1, $2, $3)`
		fmt.Println(sqlStatement, userId, thread.Uid, vote.Voice)
		_, err  = DB.Exec(sqlStatement, userId, thread.Uid, vote.Voice)
		if err != nil {
			return ThreadInfo{}, err
		}
		thread.Votes += vote.Voice
		fmt.Println("all votes: ", thread.Votes)
		return thread, nil
	} else if err != nil {
		return ThreadInfo{}, err
	}
	if (value > 0 && vote.Voice > 0) ||
		(value < 0 && vote.Voice < 0) {
		fmt.Println("all votes: ", thread.Votes)
		return thread, nil
	}

	sqlVote := `UPDATE vote SET "value" = $1, is_edited = true WHERE user_id = $2 AND thread_id = $3;`
	fmt.Println(sqlVote, vote.Voice, userId, thread.Uid)
	_, err  = DB.Exec(sqlVote, vote.Voice, userId, thread.Uid)
	if err != nil {
		return ThreadInfo{}, err
	}
	thread.Votes += vote.Voice * 2
	fmt.Println("all votes: ", thread.Votes)
	return thread, nil

	//sqlStatement := `
	//INSERT INTO vote (user_id, thread_id, "value")
	//SELECT p.uid, $1, $2 FROM profile p, get_thread WHERE p.nickname = $3;`
}

