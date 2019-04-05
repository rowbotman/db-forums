package handlers

import (
	"../db"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
)

func postChangeInfo(w http.ResponseWriter, req *http.Request) {
	var data db.DataForUpdPost
	_= json.NewDecoder(req.Body).Decode(&data)
	forum, err := db.UpdatePost(data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_ = json.NewEncoder(w).Encode(forum)
}

func PostGetInfo(w http.ResponseWriter,req *http.Request) {
	fmt.Println("search post...")
	params := mux.Vars(req)
	id := int64(0)
	if postId, ok := params["id"]; !ok {
		http.Error(w, "Can't parse id", http.StatusBadRequest)
	} else {
		id, _ = strconv.ParseInt(postId, 10, 64)
	}

	var err error
	_ = req.ParseForm() // parses request body and query and stores result in r.Form
	var array []string
	for i := 0;; i++ {
		key := fmt.Sprintf("related[%d]", i)
		values := req.Form[key] // form values are a []string
		if len(values) == 0 {
			// no more values
			break
		}
		array = append(array, values[i])
		i++
	}
	details, err := db.GetPostInfo(id, array)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	output, err := json.Marshal(details)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("content-type", "application/json")
	_, _ = w.Write(output)
}

func PostHandler(router **mux.Router) {
	fmt.Println("post handler")
	(*router).HandleFunc("/api/post/{id}/details/", postChangeInfo).Methods("POST")
	(*router).HandleFunc("/api/post/{id}/details/", PostGetInfo).Methods("GET")
}