package handlers

import (
	"../db"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"io/ioutil"
	"net/http"
)

//type User struct{
//	ID        string `json:"id,omitempty"`
//	FirstName string `json:"firstName,omitempty"`
//	LastName  string `json:"lastName,omitempty"`
//}


func userGet(w http.ResponseWriter, req *http.Request){
	params := mux.Vars(req)
	nickname, ok := params["nickname"]
	if !ok {
		http.Error(w, "can't parse slug", http.StatusBadRequest)
		return
	}
	user, err := db.SelectUser(nickname)
	if err != nil {
		http.Error(w, "incorrect slug", http.StatusBadRequest)
		return
	}
	// if not found return empty object with User structure
	_ = json.NewEncoder(w).Encode(user)
}

func userCreate(w http.ResponseWriter,req *http.Request) {
	fmt.Println("user create function")
	params := mux.Vars(req)
	body, err := ioutil.ReadAll(req.Body)
	defer req.Body.Close()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	data := db.User{}
	data.Nickname, _ = params["nickname"]
	fmt.Println("creation", data.Nickname, "is starting...")
	err = json.Unmarshal(body, &data)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	newUser, err := db.InsertIntoUser(data)
	fmt.Println("marshaling is starting...")
	if err != nil {
		fmt.Println("i am here")
		if len(newUser) > 0 {
			output, err := json.Marshal(newUser)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			fmt.Println("aaaaaa")
			w.Header().Set("content-type", "application/json")
			http.Error(w, string(output), http.StatusConflict)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	var output []byte
	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/json")
	output, err = json.Marshal(newUser[0])
	_, _ = w.Write(output)

	//_ = json.NewEncoder(w).Encode(newUser[0])
	////user.Pk = int64(id)
	//w.Header().Set("content-type", "application/json")

}

func userPost(w http.ResponseWriter, req *http.Request) {
	params := mux.Vars(req)
	nickname, _ := params["nickname"]
	body, err := ioutil.ReadAll(req.Body)
	defer req.Body.Close()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	data := db.User{}
	data.Nickname = nickname
	err = json.Unmarshal(body, &data)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	user, err := db.UpdateUser(data)
	output, err := json.Marshal(user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("content-type", "application/json")
	_, _ = w.Write(output)
}


func UserHandler(router **mux.Router) {
	fmt.Println("user handler initialized")
	(*router).HandleFunc("/api/user/{nickname}/create",  userCreate)
	(*router).HandleFunc("/api/user/{nickname}/profile/", userGet).Methods("GET")
	(*router).HandleFunc("/api/user/{nickname}/profile/", userPost).Methods("POST")
}