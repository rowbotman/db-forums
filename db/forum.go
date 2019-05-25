package db

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"
)

type Forum struct {
	Title   NullString `json:"title,omitempty"`
	User    NullString `json:"user,omitempty"`
	Slug    NullString `json:"slug,omitempty"`
	Posts   int        `json:"posts,omitempty"`
	Threads int        `json:"threads,omitempty"`
}

type DataForNewForum struct {
	Title    string `json:"title"`
	Nickname string `json:"user"`
	Slug     string `json:"slug"`
}

func InsertIntoForum(data DataForNewForum) (Forum, error) {
	sqlStatement := `SELECT u.uid FROM profile u WHERE u.nickname = $1`
	row := DB.QueryRow(sqlStatement, data.Nickname)
	authorId := 0
	err := row.Scan(&authorId)
	if err == sql.ErrNoRows {
		fmt.Println("No rows were returned!")
		return Forum{}, err
	} else if err != nil {
		panic(err)
	}

	sqlStatement = `INSERT INTO forum (default, title, author_id, slug) VALUES ($1, $2, $3)`
	_, err = DB.Exec(sqlStatement, data.Title, authorId, data.Slug)
	if err != nil {
		return Forum{}, err
	}

	fmt.Println("New forum slug:", data.Slug)
	return Forum{
		NullString{String: data.Title},
		NullString{String: data.Nickname},
		NullString{String: data.Slug},
		0,
		0,
	}, nil
}

func SelectForumInfo(slug string, isUid bool) (Forum, error) {
	sqlStatement := ``
	var row *sql.Row
	if isUid {
		sqlStatement = `
  SELECT f.title, u.nickname, f.slug, COUNT(p.uid), COUNT(t.uid) FROM forum f
  JOIN profile u ON (f.author_id = u.uid)
  JOIN post    p ON (p.forum_id = f.uid)
  JOIN thread  t ON (t.forum_id = f.uid) WHERE f.uid = $1;`
		id, err := strconv.Atoi(slug)
		if err != nil {
			return Forum{}, err
		}
		row = DB.QueryRow(sqlStatement, id)
	} else {
		sqlStatement = `
  SELECT f.title, u.nickname, f.slug, COUNT(p.uid), COUNT(t.uid) FROM forum f
  JOIN profile u ON (f.author_id = u.uid)
  JOIN post    p ON (p.forum_id = f.uid)
  JOIN thread  t ON (t.forum_id = f.uid) WHERE f.slug = $1;`
		row = DB.QueryRow(sqlStatement, slug)
	}
	var forum Forum
	err := row.Scan(
		&forum.Title,
		&forum.User,
		&forum.Slug,
		&forum.Posts,
		&forum.Threads)

	if err == sql.ErrNoRows {
		fmt.Println("No rows were returned!")
		return Forum{}, nil
	} else if err != nil {
		return Forum{}, err
	}
	return forum, nil
}

func SelectForumUsers(slug string, limit int32, since string, desc bool) ([]User, error) {
	sqlStatement := `
  SELECT u.uid, u.nickname, u.full_name, u.about, u.email FROM profile u
  JOIN post p ON (p.user_id = u.uid)
  JOIN forum f ON (p.forum_id = f.uid)
  WHERE f.title = $1 ORDER BY u.nickname`
	rows, err := DB.Query(sqlStatement, slug)
	fmt.Println(sqlStatement, slug)
	if err != nil {
		fmt.Println("forum haven't users")
		return nil, err
	}
	defer rows.Close()
	users := []User{}
	for rows.Next() {
		newUser := User{}
		err = rows.Scan(
			&newUser.Pk,
			&newUser.Nickname,
			&newUser.Name,
			&newUser.About,
			&newUser.Email)
		if err != nil {
			return nil, err
		}
		users = append(users, newUser)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return users, nil
}

func SelectForumThreads(slug string, limit int32, since string, desc bool) ([]ThreadInfo, error) {
	sqlStatement := `SELECT title FROM forum WHERE uid = $1`
	row := DB.QueryRow(sqlStatement, slug)
	forum := ""
	err := row.Scan(&forum)
	if err == sql.ErrNoRows {
		fmt.Println("No rows were returned!")
		return nil, err
	} else if err != nil {
		return nil, err
	}

	sqlStatement = `
  SELECT t.uid, t.title, u.nickname, f.title, t.message, t.votes, t.slug, t.created
  FROM forum f
  JOIN thread  t ON (t.forum_id = f.uid)
  JOIN profile p ON (t.user_id  = p.uid)
  WHERE f.title = $1 `

	var rows *sql.Rows
	if len(since) > 0 {
		if desc {
			sqlStatement += ` AND t.created <= $2 ORDER BY t.created DESC LIMIT $3;`
		} else {
			sqlStatement += ` AND t.created >= $2 ORDER BY t.created ASC  LIMIT $3;`
		}
		rows, err = DB.Query(sqlStatement, slug, since, limit)
	} else {
		if desc {
			sqlStatement += ` ORDER BY t.created DESC LIMIT $2;`
		} else {
			sqlStatement += ` ORDER BY t.created ASC  LIMIT $2;`
		}
		rows, err = DB.Query(sqlStatement, slug, limit)

	}
	if err != nil {
		fmt.Println("forum haven't users")
		return nil, err
	}
	defer rows.Close()
	threads := []ThreadInfo{}
	for rows.Next() {
		thread := ThreadInfo{}
		err = rows.Scan(
			&thread.Uid,
			&thread.Title,
			&thread.Author,
			&thread.Forum,
			&thread.Message,
			&thread.Votes,
			&thread.Slug,
			&thread.Created)
		if err != nil {
			fmt.Println("errors occurred in get threads from forum")
			return nil, err
		}
		threads = append(threads, thread)
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}
	return threads, nil
}

func InsertIntoThread(slug string, threadData *ThreadInfo) error {
	_ = threadData.Slug.Scan(strings.ToLower(threadData.Title.String))
	sqlStatement := `SELECT p.uid FROM profile p WHERE p.full_name = $1;`
	row := DB.QueryRow(sqlStatement, slug)
	author := int64(0)
	err := row.Scan(&author)
	if err == sql.ErrNoRows {
		fmt.Println("No rows were returned!")
		return err
	} else if err != nil {
		return err
	}

	sqlStatement = `SELECT f.uid, f.title FROM forum f WHERE f.slug = $1;`
	row = DB.QueryRow(sqlStatement, slug)
	forum := int64(0)
	err = row.Scan(&forum, &threadData.Forum)
	if err == sql.ErrNoRows {
		fmt.Println("No rows were returned!")
		return err
	} else if err != nil {
		return err
	}
	sqlStatement = `INSERT INTO thread VALUES(default, $1, $2, $3, $4, $5, $6) RETURNING uid`

	err = DB.QueryRow(
		sqlStatement,
		author,
		forum,
		threadData.Title,
		threadData.Slug,
		threadData.Message,
		threadData.Created).Scan(&threadData.Uid)
	if err != nil {
		return err
	}
	return nil
}
