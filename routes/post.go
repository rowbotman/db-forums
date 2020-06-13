package routes

import (
	"context"
	"encoding/json"
	db2 "forum/db"
	"forum/models"
	"forum/utils"
	"github.com/dimfeld/httptreemux"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
)

func SetPostRouter(router *httptreemux.TreeMux) {
	router.GET("/api/post/:id/details", postHandler)
	router.POST("/api/post/:id/details", postPostsHandler)
}

type postInput struct {
	Message string `json:"message"`
}

func getPostAuthor(details *models.DetailedInfo, author string, writer http.ResponseWriter) error {
	db := db2.GetDB()
	row := db.QueryRow(context.Background(), `SELECT about, email, fullname, nickname 
			FROM users WHERE nickname = $1 `, author)

	var user models.User
	err := row.Scan(&user.About, &user.Email, &user.Fullname, &user.Nickname)
	if err != nil {
		msg, _ := json.Marshal(map[string]string{"message": "User not found"})
		utils.WriteData(writer, 404, msg)
		return err
	}

	details.AuthorInfo = &user
	return nil
}

func getPostForum(details *models.DetailedInfo, forumName string, writer http.ResponseWriter) error {
	db := db2.GetDB()
	row := db.QueryRow(context.Background(), `SELECT posts, slug, threads, title, author 
			FROM forums WHERE slug = $1 `, forumName)

	var forum models.Forum
	err := row.Scan(&forum.Posts, &forum.Slug, &forum.Threads,
		&forum.Title, &forum.User)
	if err != nil {
		msg, _ := json.Marshal(map[string]string{"message": "Forum not found"})
		utils.WriteData(writer, 404, msg)
		return err
	}

	details.ForumInfo = &forum
	return nil
}

func getPostThread(details *models.DetailedInfo, tId int, writer http.ResponseWriter) error {
	db := db2.GetDB()
	row := db.QueryRow(context.Background(), `SELECT author, created, forum, id, message, slug, title, votes 
			FROM threads WHERE id = $1 `, tId)

	var thread models.Thread
	err := row.Scan(&thread.Author, &thread.Created, &thread.ForumName, &thread.Id,
		&thread.Message, &thread.Slug, &thread.Title, &thread.Votes)
	if err != nil {
		msg, _ := json.Marshal(map[string]string{"message": "Thread not found"})
		utils.WriteData(writer, 404, msg)
		return err
	}

	details.ThreadInfo = &thread
	return nil
}

func postPostsHandler(writer http.ResponseWriter, request *http.Request, ps map[string]string) {
	id, err := strconv.Atoi(ps["id"])
	if err != nil {
		http.Error(writer, "wrong ID", 400)
		return
	}

	db := db2.GetDB()

	body, err := ioutil.ReadAll(request.Body)
	defer request.Body.Close()
	if err != nil {
		http.Error(writer, err.Error(), 500)
		return
	}
	// parse input
	var input postInput
	err = json.Unmarshal(body, &input)
	if err != nil {
		http.Error(writer, err.Error(), 500)
		return
	}

	var post models.Post
	// check existing and message
	row := db.QueryRow(context.Background(), `
		    SELECT  author, created, forum, id, isEdited, message, parent, tid
			FROM posts WHERE id = $1`, id)
	err = row.Scan(&post.Author, &post.Created, &post.ForumName, &post.Id,
		&post.IsEdited, &post.Message, &post.Parent, &post.Tid)
	if err != nil {
		msg, _ := json.Marshal(map[string]string{"message": "Post not found"})
		utils.WriteData(writer, 404, msg)
		return
	}
	// update if it's needed
	if input.Message != "" && input.Message != post.Message {
		_, err = db.Exec(context.Background(), `UPDATE posts SET message = $1, isEdited = TRUE 
				WHERE id = $2`, input.Message, id)
		if err != nil {
			http.Error(writer, err.Error(), 500)
			return
		}
		post.IsEdited = true
		post.Message = input.Message
	}

	data, err := json.Marshal(post)
	if err != nil {
		http.Error(writer, err.Error(), 500)
		return
	}

	utils.WriteData(writer, 200, data)
}

func postHandler(writer http.ResponseWriter, request *http.Request, ps map[string]string) {
	id, err := strconv.Atoi(ps["id"])
	if err != nil {
		http.Error(writer, "wrong ID", 400)
		return
	}

	db := db2.GetDB()
		var details models.DetailedInfo

	related := strings.Split(request.FormValue("related"), ",")
	row := db.QueryRow(context.Background(), `SELECT author, created, forum, id, message, tid, isEdited, parent 
			FROM posts WHERE id = $1 `, id)

	var post models.Post
	err = row.Scan(&post.Author, &post.Created, &post.ForumName, &post.Id,
		&post.Message, &post.Tid, &post.IsEdited, &post.Parent)
	if err != nil {
		msg, _ := json.Marshal(map[string]string{"message": "Post not found"})
		utils.WriteData(writer, 404, msg)
		return
	}
	details.PostInfo = post

	for i := range related {
		switch related[i] {
		case "user":
			err = getPostAuthor(&details, post.Author, writer)
		case "forum":
			err = getPostForum(&details, post.ForumName, writer)
		case "thread":
			err = getPostThread(&details, post.Tid, writer)
		}

		if err != nil {
			return
		}
	}

	data, err := json.Marshal(details)
	if err != nil {
		http.Error(writer, err.Error(), 500)
	}
	utils.WriteData(writer, 200, data)
}