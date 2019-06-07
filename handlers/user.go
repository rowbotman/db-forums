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


func userGet(w http.ResponseWriter, req *http.Request) {
	fmt.Println("getting user profile... \n")
	params := mux.Vars(req)
	nickname, ok := params["nickname"]
	if !ok {
		http.Error(w, "can't parse slug", http.StatusBadRequest)
		return
	}
	user, err := db.SelectUser(nickname)
	if err != nil {
		Get404(w, "Can't find user by nickname: " + nickname)
		return
		//w.Header().Set("Content-Type", "application/json")
		//w.WriteHeader(http.StatusNotFound)
		//
		//http.Error(w, "incorrect slug", http.StatusBadRequest)
		//return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
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
	err = json.Unmarshal(body, &data)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	newUser, err := db.InsertIntoUser(data)
	if err != nil {
		if len(newUser) > 0 {
			output, err := json.Marshal(newUser)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			w.Header().Set("content-type", "application/json")
			w.WriteHeader(http.StatusConflict)
			_, err = w.Write(output)
			if err != nil {
				fmt.Println("writing failed")
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	output, err := json.Marshal(newUser[0])
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Println(string(output) + "\n")
	_, err = w.Write(output)
	if err != nil {
		fmt.Println("writing failed")
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

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
	fmt.Println(data)
	fmt.Println()
	user, err := db.UpdateUser(data)
	if err != nil {
		w.Header().Set("content-type", "application/json")
		if err.Error() == "no rows" {
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(NotFoundPage{"Can't find user by nickname: " + nickname})
			return
		}
		w.WriteHeader(http.StatusConflict)
		_ = json.NewEncoder(w).Encode(NotFoundPage{err.Error()})
		return
	}
	fmt.Println(user)
	output, err := json.Marshal(user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(output)
}


func UserHandler(router **mux.Router) {
	fmt.Println("user handler initialized")
	(*router).HandleFunc("/api/user/{nickname}/create",  userCreate)
	(*router).HandleFunc("/api/user/{nickname}/profile", userGet).Methods("GET")
	(*router).HandleFunc("/api/user/{nickname}/profile", userPost).Methods("POST")
}