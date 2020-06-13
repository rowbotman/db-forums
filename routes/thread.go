package routes

import (
	"context"
	"encoding/json"
	"fmt"
	db2 "forum/db"
	"forum/models"
	"forum/utils"
	"github.com/dimfeld/httptreemux"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

func SetThreadRouter(router *httptreemux.TreeMux) {
	router.POST("/api/thread/:slug/create", threadCreateHandler)
	router.GET("/api/thread/:slug/details", getDetailsHandler)
	router.POST("/api/thread/:slug/details", postDetailsHandler)
	router.GET("/api/thread/:slug/posts", getPosts)
	router.POST("/api/thread/:slug/vote", threadVoteHandler)
}

func getDetailsHandler(writer http.ResponseWriter, request *http.Request, ps map[string]string) {
	slug := ps["slug"]
	// fmt.Println("getDetHandler")
	// start := time.Now()
	id, _ := strconv.Atoi(slug)

	db := db2.GetDB()

	row := db.QueryRow(context.Background(), "SELECT * " +
		"FROM threads WHERE slug = $1 OR id = $2;", slug, id)

	var thread models.Thread
	err := row.Scan(&thread.Author, &thread.Created, &thread.ForumName, &thread.Id,
		&thread.Message, &thread.Slug, &thread.Title, &thread.Votes)
	if err != nil {
		msg, _ := json.Marshal(map[string]string{"message": "Thread not found"})
		utils.WriteData(writer, 404, msg)
		// fmt.Println("getDH", time.Now().Sub(start))
		return
	}

	data, err := json.Marshal(thread)
	if err != nil {
		http.Error(writer, err.Error(), 500)
	}
	// fmt.Println(data)
	utils.WriteData(writer, 200, data)
	// fmt.Println("getDH", time.Now().Sub(start))
}

func postDetailsHandler(writer http.ResponseWriter, request *http.Request, ps map[string]string) {
	// fmt.Println("postDH")
	// start := time.Now()
	slug := ps["slug"]
	id, _ := strconv.Atoi(slug)

	db := db2.GetDB()

	body, err := ioutil.ReadAll(request.Body)
	defer request.Body.Close()
	if err != nil {
		http.Error(writer, err.Error(), 500)
		return
	}
	// parse input
	var input models.ThreadUpdate
	err = json.Unmarshal(body, &input)
	if err != nil {
		http.Error(writer, err.Error(), 500)
		return
	}
	// form query
	var query string
	if input.Message == "" && input.Title == "" {
		query = `SELECT * FROM threads WHERE id = $1 OR slug = $2`
	} else if input.Message == "" {
		query = fmt.Sprintf(`UPDATE threads 
				SET title = '%s' WHERE id = $1 OR slug = $2 RETURNING *`, input.Title)
	} else if input.Title == "" {
		query = fmt.Sprintf(`UPDATE threads 
				SET message = '%s' WHERE id = $1 OR slug = $2 RETURNING *`, input.Message)
	} else {
		query = fmt.Sprintf(`UPDATE threads 
				SET message = '%s', title = '%s' WHERE id = $1 OR slug = $2 RETURNING *`,
			input.Message, input.Title)
	}

	row := db.QueryRow(context.Background(), query, id, slug)

	var thread models.Thread
	err = row.Scan(&thread.Author, &thread.Created, &thread.ForumName, &thread.Id,
		&thread.Message, &thread.Slug, &thread.Title, &thread.Votes)
	if err != nil {
		// fmt.Println(err)
		// fmt.Println(3)
		msg, _ := json.Marshal(map[string]string{"message": "Thread not found"})
		utils.WriteData(writer, 404, msg)
		// fmt.Println("postDH", time.Now().Sub(start))
		return
	}

	data, err := json.Marshal(thread)
	if err != nil {
		http.Error(writer, err.Error(), 500)
	}
	// fmt.Println(thread)
	utils.WriteData(writer, 200, data)
	// fmt.Println("postDH", time.Now().Sub(start))
}

func threadCreateHandler(writer http.ResponseWriter, request *http.Request, ps map[string]string) {
	body, err := ioutil.ReadAll(request.Body)
	//fmt.Println("threadCreateHandler")
	//start := time.Now()
	defer request.Body.Close()
	if err != nil {
		http.Error(writer, err.Error(), 500)
		return
	}
	// parse input
	var posts []models.Post
	//fmt.Println(body)
	err = json.Unmarshal(body, &posts)
	if err != nil {
		http.Error(writer, err.Error(), 500)
		return
	}

	slug := ps["slug"]
	tid, _ := strconv.Atoi(slug)
	var post models.Post
	db := db2.GetDB()
	sqlTime := `SELECT current_timestamp(3);`
	curTime := time.Time{}
	err = db.QueryRow(context.Background(), sqlTime).Scan(&curTime)
	if err != nil {
		http.Error(writer, err.Error(), 500)
		return
	}
	// check thread and get forum
	err = db.QueryRow(context.Background(), "SELECT id, forum " +
		"FROM threads WHERE slug = $1 OR id = $2", slug, tid).Scan(&post.Tid, &post.ForumName)
	if err != nil {
		msg, _ := json.Marshal(map[string]string{"message": "Thread not found"})
		utils.WriteData(writer, 404, msg)
		//fmt.Println("thrCreateH", time.Now().Sub(start))
		return
	}

	if len(posts) != 0 {
		if posts[0].Parent != 0 {
			var parentTId int
			_ = db.QueryRow(context.Background(), "SELECT tid FROM posts WHERE id = $1", posts[0].Parent).Scan(&parentTId)
			if parentTId != post.Tid {
				msg, _ := json.Marshal(map[string]string{"message": "Parent in another thread"})
				utils.WriteData(writer, 409, msg)
				return
			}
		}
	}

	for i := 0; i < len(posts); i++ {
		posts[i].Tid = post.Tid
		posts[i].ForumName = post.ForumName
		posts[i].Created = curTime
		err = createPost(&posts[i], writer, request)
		if err != nil {
			break
		}
	}

	if err == nil {
		data, err := json.Marshal(posts)
		if err != nil {
			http.Error(writer, err.Error(), 500)
			return
		}

		utils.WriteData(writer, 201, data)
		//fmt.Println("thrCrH", time.Now().Sub(start))
	}
}

func createPost(post *models.Post, writer http.ResponseWriter, request *http.Request) error {
	db := db2.GetDB()
	// fmt.Println(post.Created)
	// check user
	err := db.QueryRow(context.Background(), "SELECT nickname "+
		"FROM users WHERE nickname = $1", post.Author).Scan(&post.Author)
	if err != nil {
		msg, _ := json.Marshal(map[string]string{"message": "User not found"})
		utils.WriteData(writer, 404, msg)
		return err
	}

	query := "INSERT INTO posts (author, forum, message, parent, tid, created, slug, rootId) " +
		"VALUES ($1, $2, $3, $4, $5, $6, " +
		"(SELECT slug FROM posts WHERE id = $4) || (SELECT currval('posts_id_seq')::integer), "
	if post.Parent == 0 {
		query += "(SELECT currval('posts_id_seq')::integer)) RETURNING id"
	} else {
		query += "(SELECT rootId FROM posts WHERE id = $4)) RETURNING id"
	}
	err = db.QueryRow(context.Background(), query,
		post.Author, post.ForumName, post.Message,
		post.Parent, post.Tid, post.Created).Scan(&post.Id)

	if err != nil {
		msg, _ := json.Marshal(map[string]string{"message": err.Error()})
		utils.WriteData(writer, 409, msg)
		return err
	}

	return nil
}

func threadVoteHandler(writer http.ResponseWriter, request *http.Request, ps map[string]string) {
	slug := ps["slug"]
	id, _ := strconv.Atoi(slug)

	body, err := ioutil.ReadAll(request.Body)
	defer request.Body.Close()
	if err != nil {
		http.Error(writer, err.Error(), 500)
		return
	}
	// parse input
	var vote models.Vote
	err = json.Unmarshal(body, &vote)
	if err != nil {
		http.Error(writer, err.Error(), 500)
		return
	}

	var thread models.Thread
	db := db2.GetDB()
	transaction, _ := db.Begin(context.Background())
	// check thread and get forum
	err = transaction.QueryRow(context.Background(), "SELECT author, created, id, forum, message, slug::text, title, votes " +
		"FROM threads WHERE slug = $1 OR id = $2", slug, id).Scan(
			&thread.Author, &thread.Created, &thread.Id, &thread.ForumName,
			&thread.Message, &thread.Slug, &thread.Title, &thread.Votes)
	if err != nil {
		_ = transaction.Rollback(context.Background())
		msg, _ := json.Marshal(map[string]string{"message": "Thread not found"})
		utils.WriteData(writer, 404, msg)
		return
	}
	// create/update vote
	r, err := transaction.Exec(context.Background(), "UPDATE votes SET voice=$1 " +
		"WHERE tid=$2 AND nickname=$3;", vote.Voice, thread.Id, vote.Nickname)
	if count:= r.RowsAffected(); count == 0 {
		_, err := transaction.Exec(context.Background(), "INSERT INTO votes (nickname, tid, voice)" +
			"VALUES ($1, $2, $3);", vote.Nickname, thread.Id, vote.Voice)
		if err != nil {
			_ = transaction.Rollback(context.Background())
			msg, _ := json.Marshal(map[string]string{"message": "User not found"})
			utils.WriteData(writer, 404, msg)
			return
		}
	}
	// get new votes
	err = transaction.QueryRow(context.Background(), "SELECT votes FROM threads " +
		"WHERE id = $1", thread.Id).Scan(&thread.Votes)
	if err != nil {
		_ = transaction.Rollback(context.Background())
		http.Error(writer, err.Error(), 500)
		return
	}

	data, err := json.Marshal(thread)
	if err != nil {
		_ = transaction.Rollback(context.Background())
		http.Error(writer, err.Error(), 500)
		return
	}
	_ = transaction.Commit(context.Background())
	utils.WriteData(writer, 200, data)
}

type postsInput struct {
	Slug string
	Id int
	ParentId int
	Limit string
	Since string
	Sort string
	Desc bool
}

func getFlatPosts(input postsInput, writer http.ResponseWriter, r *http.Request) error {
	db := db2.GetDB()

	query := "SELECT id, author, created, forum, isEdited, message, parent, tid " +
		"FROM posts WHERE tid = $1 "
	if input.Since != ""{
		if input.Desc {
			query += "AND id < '" + input.Since + "' "
		} else {
			query += "AND id > '" + input.Since + "' "
		}
	}
	query += "ORDER BY id "
	if input.Desc {
		query += "DESC "
	}
	if limit := r.FormValue("limit"); limit != "" {
		query += "LIMIT " + limit + " "
	}

	rows, err := db.Query(context.Background(), query, input.Id)
	defer rows.Close()
	if err != nil {
		http.Error(writer, err.Error(), 500)
		return err
	}

	result := []byte("[ ")
	for rows.Next() {
		if len(result) > 2 {
			result = append(result, ',')
		}

		post := models.Post{}
		err = rows.Scan(&post.Id, &post.Author, &post.Created, &post.ForumName,
			&post.IsEdited, &post.Message, &post.Parent, &post.Tid)
		if err != nil {
			http.Error(writer, err.Error(), 500)
			return err
		}

		data, _ := json.Marshal(post)
		result = append(result, data...)
	}
	result = append(result, ']')
	// fmt.Println(result)
	utils.WriteData(writer, 200, result)
	return nil
}

func getTreePosts(input postsInput, writer http.ResponseWriter, r *http.Request) error {
	db := db2.GetDB()

	query := "SELECT id, author, created, forum, isEdited, message, parent, tid" +
		" FROM posts WHERE tid = $1 "
	if input.Since != ""{
		if input.Desc {
			query += "AND slug < (SELECT slug FROM posts WHERE id = " + input.Since + ") "
		} else {
			query += "AND slug > (SELECT slug FROM posts WHERE id = " + input.Since + ") "
		}
	}
	query += "ORDER BY slug "
	if input.Desc {
		query += "DESC "
	}
	if limit := r.FormValue("limit"); limit != "" {
		query += "LIMIT " + limit + " "
	}

	rows, err := db.Query(context.Background(), query, input.Id)
	defer rows.Close()
	if err != nil {
		http.Error(writer, err.Error(), 500)
		return err
	}

	result := []byte("[ ")
	for rows.Next() {
		if len(result) > 2 {
			result = append(result, ',')
		}

		post := models.Post{}
		//slug := ""
		err = rows.Scan(&post.Id, &post.Author, &post.Created, &post.ForumName,
			&post.IsEdited, &post.Message, &post.Parent, &post.Tid)
		if err != nil {
			http.Error(writer, err.Error(), 500)
			return err
		}

		data, _ := json.Marshal(post)
		result = append(result, data...)
	}
	result = append(result, ']')
	// fmt.Println(result)
	utils.WriteData(writer, 200, result)
	return nil
}

func getParentTreePosts(input postsInput, writer http.ResponseWriter, r *http.Request) error {
	db := db2.GetDB()

	query := "WITH roots AS ( " +
		"SELECT id FROM posts WHERE tid = $1 AND parent = 0 "
	if input.Since != ""{
		if input.Desc {
			query += "AND id < (SELECT rootId FROM posts WHERE id = " + input.Since + ") "
		} else {
			query += "AND id > (SELECT rootId FROM posts WHERE id = " + input.Since + ") "
		}
	}
	query += "ORDER BY id"
	if input.Desc {
		query += " DESC"
	}
	if limit := r.FormValue("limit"); limit != "" {
		query += " LIMIT " + limit + " "
	}
	query += ") SELECT posts.id, author, created, forum, isEdited, message, parent, tid " +
		"FROM posts JOIN roots ON roots.id = rootId "

	query += "ORDER BY "
	if input.Desc {
		query += " rootId DESC, slug"
	} else {
		query += " slug"
	}

	rows, err := db.Query(context.Background(), query, input.Id)
	defer rows.Close()
	if err != nil {
		http.Error(writer, err.Error(), 500)
		return err
	}

	result := []byte("[ ")
	for rows.Next() {
		if len(result) > 2 {
			result = append(result, ',')
		}

		post := models.Post{}
		//slug := ""
		err = rows.Scan(&post.Id, &post.Author, &post.Created, &post.ForumName,
			&post.IsEdited, &post.Message, &post.Parent, &post.Tid)
		if err != nil {
			http.Error(writer, err.Error(), 500)
			return err
		}

		data, _ := json.Marshal(post)
		result = append(result, data...)
	}
	result = append(result, ']')
	// fmt.Println(result)
	utils.WriteData(writer, 200, result)
	return nil
}

// Not neded any more, but let it be here
func getChilPosts(input postsInput, writer http.ResponseWriter, request *http.Request) ([]byte, error) {
	db := db2.GetDB()


	query := "SELECT id, author, created, forum, isEdited, message, parent, tid " +
		"FROM posts WHERE tid = $1 AND parent = $2 "
	if input.Since != ""{
		if input.Desc {
			query += "AND id < '" + input.Since + "' "
		} else {
			query += "AND id > '" + input.Since + "' "
		}
	}
	query += "ORDER BY id "
	if input.Desc {
		query += "DESC "
	}
	if input.Limit != "" {
		query += "LIMIT " + input.Limit + " "
	}

	rows, err := db.Query(context.Background(), query, input.Id, input.ParentId)
	defer rows.Close()
	if err != nil {
		http.Error(writer, err.Error(), 500)
		return []byte{}, err
	}

	var result []byte
	for rows.Next() {
		result = append(result, ',')

		post := models.Post{}
		err = rows.Scan(&post.Id, &post.Author, &post.Created, &post.ForumName,
			&post.IsEdited, &post.Message, &post.Parent, &post.Tid)
		if err != nil {
			http.Error(writer, err.Error(), 500)
			return []byte{}, err
		}

		data, _ := json.Marshal(post)
		result = append(result, data...)
	}

	return result, nil
}

func getPosts(writer http.ResponseWriter, r *http.Request, ps map[string]string) {
	if r.Method == "GET" {
		//fmt.Println("getPosts")
		//start := time.Now()
		var input postsInput
		input.Slug = ps["slug"]
		input.Id, _ = strconv.Atoi(input.Slug)
		// check if forum exist
		db := db2.GetDB()
		err := db.QueryRow(context.Background(), "SELECT id FROM threads " +
			"WHERE slug = $1 OR id = $2", input.Slug, input.Id).Scan(&input.Id)
		if err != nil {
			msg, _ := json.Marshal(map[string]string{"message": "Thread not found"})
			utils.WriteData(writer, 404, msg)
			//fmt.Println("getPosts", time.Now().Sub(start))
			return
		}

		// get params
		input.Since = r.FormValue("since")
		input.Desc = r.FormValue("desc") == "true"
		input.Sort = r.FormValue("sort")
		input.Limit = r.FormValue("limit")
		switch input.Sort {
		case "flat", "":
			err = getFlatPosts(input, writer, r)
			//fmt.Print("flat ")
		case "tree":
			err = getTreePosts(input, writer, r)
			//fmt.Print("tree ")
		case "parent_tree":
			err = getParentTreePosts(input, writer, r)
			//fmt.Print("ptree ")
		}
		//fmt.Println("getPosts", time.Now().Sub(start), input)
	}
}
