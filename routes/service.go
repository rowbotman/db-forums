package routes

import (
	"context"
	"encoding/json"
	db2 "forum/db"
	"forum/utils"
	"github.com/dimfeld/httptreemux"
	"net/http"
)

func SetServiceRouter(router *httptreemux.TreeMux) {
	router.POST("/api/service/clear", clearHandler)
	router.GET("/api/service/status", statusHandler)
}

func clearHandler(writer http.ResponseWriter, request *http.Request, ps map[string]string) {
	db := db2.GetDB()
	_, err := db.Exec(context.Background(), `
			TRUNCATE TABLE forum_users, votes, posts, threads, forums, users 
			  RESTART IDENTITY`)
	if err != nil {
		http.Error(writer, err.Error(), 500)
		return
	}

	writer.WriteHeader(200)
}

type dbStats struct {
	Forums int `json:"forum"`
	Posts int `json:"post"`
	Threads int `json:"thread"`
	Users int `json:"user"`
}

func statusHandler(writer http.ResponseWriter, request *http.Request, ps map[string]string) {
	db := db2.GetDB()
	var stats dbStats

	err := db.QueryRow(context.Background(), `SELECT COUNT(*) FROM forums`).Scan(&stats.Forums)
	if err != nil {
		http.Error(writer, err.Error(), 500)
		return
	}
	err = db.QueryRow(context.Background(), `SELECT COUNT(*) FROM threads`).Scan(&stats.Threads)
	if err != nil {
		http.Error(writer, err.Error(), 500)
		return
	}
	err = db.QueryRow(context.Background(), `SELECT COUNT(*) FROM posts`).Scan(&stats.Posts)
	if err != nil {
		http.Error(writer, err.Error(), 500)
		return
	}
	err = db.QueryRow(context.Background(), `SELECT COUNT(*) FROM users`).Scan(&stats.Users)
	if err != nil {
		http.Error(writer, err.Error(), 500)
		return
	}

	data, err := json.Marshal(stats)
	if err != nil {
		http.Error(writer, err.Error(), 500)
		return
	}
	utils.WriteData(writer, 200, data)
}