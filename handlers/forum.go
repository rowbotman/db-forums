package handlers

import (
	"../db"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
)

func forumCreate(w http.ResponseWriter,req *http.Request) {
	var data db.DataForNewForum
	_= json.NewDecoder(req.Body).Decode(&data)
	forum, err := db.InsertIntoForum(data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_ = json.NewEncoder(w).Encode(forum)
}

func forumGetInfo(w http.ResponseWriter,req *http.Request) {
	fmt.Println("in forum get info")
	params := mux.Vars(req)
	forumSlug, ok := params["slug"]
	if !ok {
		http.Error(w, "incorrect slug", http.StatusBadRequest)
		return
	}
	log.Println(forumSlug)
	forum, err := db.SelectForumInfo(forumSlug, false)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	js, err := json.Marshal(forum)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	_, _ = w.Write(js)
}

func forumGetUsers(w http.ResponseWriter, req *http.Request) {
	// users := select id, firstName, lastName, email from User JOIN forum on (forum.user_id = u.id) where forum.id = req.id;
	params := mux.Vars(req)
	slugOrId, _ := params["slug"]
	var err error
	limit := int64(100)
	if limitStr := req.URL.Query().Get("limit"); len(limitStr) > 0 {
		limit, err = strconv.ParseInt(limitStr, 10, 32)
		if err != nil {
			limit = 100
		}
	}
	since := req.URL.Query().Get("since")
	if len(since) <= 0 {
		since = ""
	}

	desc, err := strconv.ParseBool(req.URL.Query().Get("desc"))
	if err != nil {
		desc = false
	}

	users, err := db.SelectForumUsers(slugOrId, int32(limit), since, desc)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	output, err := json.Marshal(users)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("content-type", "application/json")
	_, _ = w.Write(output)
}

func forumGetThreads(w http.ResponseWriter,req *http.Request) {
	// threads := select id, title, slug, author from thread t JOIN forum f on (forum.user_id = t.id) where forum.id = req.id;
	params := mux.Vars(req)
	slugOrId, _ := params["slug"]
	var err error
	limit := int64(100)
	if limitStr := req.URL.Query().Get("limit"); len(limitStr) > 0 {
		limit, err = strconv.ParseInt(limitStr, 10, 32)
		if err != nil {
			limit = 100
		}
	}
	since := req.URL.Query().Get("since")
	if len(since) <= 0 {
		since = ""
	}

	desc, err := strconv.ParseBool(req.URL.Query().Get("desc"))
	if err != nil {
		desc = false
	}

	users, err := db.SelectForumThreads(slugOrId, int32(limit), since, desc)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	output, err := json.Marshal(users)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("content-type", "application/json")
	_, _ = w.Write(output)
}

func forumCreateThread(w http.ResponseWriter,req *http.Request) {
	params := mux.Vars(req)
	slugOrId, _ := params["slug"]
	data := db.ThreadInfo{}
	body, err := ioutil.ReadAll(req.Body)
	defer req.Body.Close()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	err = json.Unmarshal(body, &data)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	err = db.InsertIntoThread(slugOrId, &data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	output, err := json.Marshal(data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("content-type", "application/json")
	_, _ = w.Write(output)
}


func ForumHandler(router **mux.Router) {
	(*router).HandleFunc("/api/forum/create/",         forumCreate).Methods("POST")
	(*router).HandleFunc("/api/forum/{slug}/details/", forumGetInfo).Methods("GET")
	(*router).HandleFunc("/api/forum/{slug}/create/",  forumCreateThread).Methods("POST")
	(*router).HandleFunc("/api/forum/{slug}/users/",   forumGetUsers).Methods("GET")
	(*router).HandleFunc("/api/forum/{slug}/threads/", forumGetThreads).Methods("GET")
}

