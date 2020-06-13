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
	"log"
	"net/http"
)

func SetForumRouter(router *httptreemux.TreeMux) {
	router.POST("/api/forum/:slug/create", slugCreateHandler)
	router.GET("/api/forum/:slug/details", getForum)
	router.GET("/api/forum/:slug/threads", getThreads)
	router.GET("/api/forum/:slug/users", getUsers)
	router.POST("/api/forum/create", createHandler)
}

func createHandler(writer http.ResponseWriter, request *http.Request, ps map[string]string) {
	body, err := ioutil.ReadAll(request.Body)
	defer request.Body.Close()
	if err != nil {
		http.Error(writer, err.Error(), 500)
	}
	// parse input
	var input models.ForumInput
	err = json.Unmarshal(body, &input)
	if err != nil {
		http.Error(writer, err.Error(), 500)
	}
	// check user
	db := db2.GetDB()
	err = db.QueryRow(context.Background(),"SELECT nickname " +
		"FROM users WHERE nickname = $1", input.User).Scan(&input.User)

	if err != nil {
		msg, _ := json.Marshal(map[string]string{"message": "User not found"})
		utils.WriteData(writer, 404, msg)
		return
	}

	_, err = db.Exec(context.Background(), "INSERT INTO forums (slug, title, author) " +
		"VALUES ($1, $2, $3)", input.Slug, input.Title, input.User)
	if err == nil {
		data, err := json.Marshal(input)
		if err != nil {
			http.Error(writer, err.Error(), 500)
		}

		utils.WriteData(writer, 201, data)
		return
	} else {
		row := db.QueryRow(context.Background(),"SELECT * FROM forums WHERE slug = $1", input.Slug)

		var f models.Forum
		var id int
		err = row.Scan(&id, &f.Posts, &f.Slug, &f.Threads, &f.Title, &f.User)
		if err != nil {
			http.Error(writer, err.Error(), 500)
		}

		data, _ := json.Marshal(f)
		utils.WriteData(writer, 409, data)
	}
}

func slugCreateHandler(writer http.ResponseWriter, request *http.Request, ps map[string]string) {
	if request.Method == "POST" {
		slug := ps["slug"]

		body, err := ioutil.ReadAll(request.Body)
		defer request.Body.Close()
		if err != nil {
			http.Error(writer, err.Error(), 500)
		}
		// parse input
		var thread models.Thread
		err = json.Unmarshal(body, &thread)
		if err != nil {
			http.Error(writer, err.Error(), 500)
		}
		// check user
		db := db2.GetDB()
		row := db.QueryRow(context.Background(), "SELECT nickname " +
			"FROM users WHERE nickname = $1", thread.Author)
		err = row.Scan(&thread.Author)
		if err != nil {
			msg, _ := json.Marshal(map[string]string{"message": "User not found"})
			utils.WriteData(writer, 404, msg)
			return
		}
		// Check forum
		err = db.QueryRow(context.Background(), "SELECT slug "+
			"FROM forums WHERE slug = $1", slug).Scan(&thread.ForumName)
		if err != nil {
			msg, _ := json.Marshal(map[string]string{"message": "Forum not found"})
			utils.WriteData(writer, 404, msg)
			return
		}

		if thread.Slug != nil {
			err = db.QueryRow(context.Background(), "INSERT INTO threads (author, created, forum, message, title, slug) "+
				"VALUES ($1, $2, $3, $4, $5, $6) RETURNING id", thread.Author, thread.Created,
				thread.ForumName, thread.Message, thread.Title, thread.Slug).Scan(&thread.Id)
		} else {
			err = db.QueryRow(context.Background(), "INSERT INTO threads (author, created, forum, message, title) "+
				"VALUES ($1, $2, $3, $4, $5) RETURNING id", thread.Author, thread.Created,
				thread.ForumName, thread.Message, thread.Title).Scan(&thread.Id)
		}
		log.Println(context.Background(), "INSERT INTO threads (author, created, forum, message, title, slug) "+
			"VALUES ($1, $2, $3, $4, $5, $6) RETURNING id", thread.Author, thread.Created,
			thread.ForumName, thread.Message, thread.Title, thread.Slug)

		fmt.Println()
		if err == nil {
			data, err := json.Marshal(thread)
			if err != nil {
				http.Error(writer, err.Error(), 500)
				return
			}

			utils.WriteData(writer, 201, data)
			return
		} else {
			row := db.QueryRow(context.Background(), "SELECT * FROM threads WHERE slug = $1", thread.Slug)

			var thr models.Thread
			err = row.Scan(&thr.Author, &thr.Created, &thr.ForumName, &thr.Id,
				&thr.Message, &thr.Slug, &thr.Title, &thr.Votes)
			if err != nil {
				http.Error(writer, err.Error(), 500)
				return
			}

			data, _ := json.Marshal(thr)
			utils.WriteData(writer, 409, data)
		}
	}
}

func getForum(writer http.ResponseWriter, r *http.Request, ps map[string]string) {
	if r.Method == "GET" {
		slug := ps["slug"]
		db := db2.GetDB()
		row := db.QueryRow(context.Background(), "SELECT posts, slug, threads, title, author " +
			"FROM forums WHERE slug = $1", slug)

		var forum models.Forum
		err := row.Scan(&forum.Posts, &forum.Slug, &forum.Threads, &forum.Title, &forum.User)
		if err != nil {
			msg, _ := json.Marshal(map[string]string{"message": "404"})
			utils.WriteData(writer, 404, msg)
			return
		}

		data, err := json.Marshal(forum)
		if err != nil {
			http.Error(writer, err.Error(), 500)
		}
		utils.WriteData(writer, 200, data)
	}
}

func getUsers(writer http.ResponseWriter, r *http.Request, ps map[string]string) {
	if r.Method == "GET" {
		slug := ps["slug"]
		// check if forum exist
		db := db2.GetDB()
		var forum models.Forum
		err := db.QueryRow(context.Background(), "SELECT slug FROM forums WHERE slug = $1", slug).Scan(&forum.Slug)
		if err != nil {
			msg, _ := json.Marshal(map[string]string{"message": "Forum not found"})
			utils.WriteData(writer, 404, msg)
			return
		}
		// TODO use forum's id or not...
		// Form query. It' important to use fUser instead of nickname here, because
		// fUser's collation is overwrite to compare symbols like '-' or '.' correctly
		query := `SELECT about, email, fullname, fUser 
					FROM users JOIN forum_users ON fUser = nickname AND forum = $1 `

		since := r.FormValue("since")
		sort := r.FormValue("desc")
		if since != ""{
			if sort == "true" {
				query += "AND fUser < '" + since + "' "
			} else {
				query += "AND fUser > '" + since + "' "
			}
		}
		query += "ORDER BY fUser "
		if sort != "" && sort != "false" {
			query += "DESC "
		}
		if limit := r.FormValue("limit"); limit != "" {
			query += "LIMIT " + limit + " "
		}

		rows, err := db.Query(context.Background(), query, slug)
		defer rows.Close()
		if err != nil {
			http.Error(writer, err.Error(), 500)
			return
		}

		result := []byte("[ ")
		for rows.Next() {
			if len(result) > 2 {
				result = append(result, ',')
			}

			usr := models.User{}
			err = rows.Scan(&usr.About, &usr.Email, &usr.Fullname, &usr.Nickname)
			if err != nil {
				http.Error(writer, err.Error(), 500)
				return
			}

			data, _ := json.Marshal(usr)
			result = append(result, data...)
		}
		result = append(result, ']')

		utils.WriteData(writer, 200, result)
	}
}

func getThreads(writer http.ResponseWriter, r *http.Request, ps map[string]string) {
	if r.Method == "GET" {
		slug := ps["slug"]
		// check if forum exist
		db := db2.GetDB()
		var forum models.Forum
		err := db.QueryRow(context.Background(), "SELECT slug FROM forums WHERE slug = $1", slug).Scan(&forum.Slug)
		if err != nil {
			msg, _ := json.Marshal(map[string]string{"message": "Forum not found"})
			utils.WriteData(writer, 404, msg)
			return
		}

		// form query
		query := "SELECT author, created, forum, id, message, slug, title, votes FROM threads WHERE forum = $1 "
		since := r.FormValue("since")
		sort := r.FormValue("desc")
		if since != ""{
			if sort == "true" {
				query += "AND created <= '" + since + "' "
			} else {
				query += "AND created >= '" + since + "' "
			}
		}
		query += "ORDER BY created "
		if sort != "" && sort != "false" {
			query += "DESC "
		}
		if limit := r.FormValue("limit"); limit != "" {
			query += "LIMIT " + limit + " "
		}

		rows, err := db.Query(context.Background(), query, slug)
		defer rows.Close()
		if err != nil {
			http.Error(writer, err.Error(), 500)
			return
		}

		result := []byte("[ ")
		for rows.Next() {
			if len(result) > 2 {
				result = append(result, ',')
			}

			thr := models.Thread{}
			err = rows.Scan(&thr.Author, &thr.Created, &thr.ForumName, &thr.Id,
				&thr.Message, &thr.Slug, &thr.Title, &thr.Votes)
			if err != nil {
				http.Error(writer, err.Error(), 500)
				return
			}

			data, _ := json.Marshal(thr)
			result = append(result, data...)
		}
		result = append(result, ']')

		utils.WriteData(writer, 200, result)
	}
}