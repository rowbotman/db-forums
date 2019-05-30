package handlers

import (
	"../db"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"io/ioutil"
	"net/http"
	"strconv"
)

func threadChangeInfo(w http.ResponseWriter,req *http.Request) {
	fmt.Println("change thread info is startingq")
	params := mux.Vars(req)
	slugOrId, _ := params["slug_or_id"]
	thread := db.ThreadInfo{}
	body, err := ioutil.ReadAll(req.Body)
	defer req.Body.Close()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = json.Unmarshal(body, &thread)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = db.UpdateThread(slugOrId, &thread)
	output, err := json.Marshal(thread)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("content-type", "application/json")
	_, _ = w.Write(output)
}

func threadCreate(w http.ResponseWriter,req *http.Request) {
	fmt.Println("\n new post creation is starting...")
	params := mux.Vars(req)
	slugOrId, _ := params["slug_or_id"]
	data := []db.Post{}
	body, err := ioutil.ReadAll(req.Body)
	defer req.Body.Close()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	err = json.Unmarshal([]byte(body), &data)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	//fmt.Println("\tits post data: ", data)
	forum, err := db.InsertNewPosts(slugOrId, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(forum)
}

func threadGetInfo(w http.ResponseWriter,req *http.Request) {
	params := mux.Vars(req)
	slugOrId, _ := params["slug_or_id"]
	_, err := strconv.ParseInt(slugOrId, 10, 64)
	thread := db.ThreadInfo{}
	if err != nil {
		thread, err = db.SelectFromThread(slugOrId, false)
	} else {
		thread, err = db.SelectFromThread(slugOrId, true)
	}
	output, err := json.Marshal(thread)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("content-type", "application/json")
	_, _ = w.Write(output)
}


func threadGetPosts(w http.ResponseWriter, req *http.Request) {
	params := mux.Vars(req)
	slugOrId, _ := params["slug_or_id"]
	var err error
	limit := int64(100)
	if limitStr := req.URL.Query().Get("limit"); len(limitStr) != 0 {
		limit, err = strconv.ParseInt(limitStr, 10, 32)
		if err != nil {
			limit = 100
		}
	}

	since := int64(0)
	if sinceStr := req.URL.Query().Get("since"); len(sinceStr) != 0 {
		since, err = strconv.ParseInt(sinceStr, 10, 64)
		if err != nil {
			since = 0
		}
	}

	sort := req.URL.Query().Get("sort")
	if len(sort) == 0 {
		sort = "flat"
	}
	desc, err := strconv.ParseBool(req.URL.Query().Get("desc"))
	if err != nil {
		desc = false
	}

	posts, err := db.SelectThreadPosts(slugOrId, int32(limit), since, sort, desc)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	output, err := json.Marshal(posts)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("content-type", "application/json")
	_, _ = w.Write(output)
}

func threadVote(w http.ResponseWriter,req *http.Request) {
	fmt.Println("\nthread vote is starting...")
	params := mux.Vars(req)
	slugOrId, _ := params["slug_or_id"]
	body, err := ioutil.ReadAll(req.Body)
	defer req.Body.Close()

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	voteInfo := db.VoteInfo{}
	err = json.Unmarshal(body, &voteInfo)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Println(voteInfo)
	thread, err := db.UpdateVote(slugOrId, voteInfo)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	threadData := map[string]interface{}{
		"author" : thread.Author,
		"created": thread.Created,
		"forum"  : thread.Forum,
		"id"     : thread.Uid,
		"message": thread.Message,
		"slug"   : thread.Slug,
		"title"  : thread.Title,
		"votes"  : thread.Votes,
	}
	fmt.Println(threadData["votes"])

	output, err := json.Marshal(threadData)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("content-type", "application/json")
	_, _ = w.Write(output)
}


func ThreadHandler(router **mux.Router) {
	(*router).HandleFunc("/api/thread/{slug_or_id}/create",  threadCreate).Methods("POST")
	(*router).HandleFunc("/api/thread/{slug_or_id}/details", threadGetInfo).Methods("GET")
	(*router).HandleFunc("/api/thread/{slug_or_id}/details", threadChangeInfo).Methods("POST")
	(*router).HandleFunc("/api/thread/{slug_or_id}/posts",   threadGetPosts).Methods("GET")
	(*router).HandleFunc("/api/thread/{slug_or_id}/vote",    threadVote).Methods("POST")
}
