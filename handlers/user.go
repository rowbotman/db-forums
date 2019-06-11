package handlers

import (
	"../db"
	"encoding/json"
	"fmt"
	//"github.com/go-zoo/bone"
	"github.com/naoina/denco"
	"io/ioutil"
	"net/http"
)

func userGet(w http.ResponseWriter, req *http.Request, ps denco.Params) {
	//params := mux.Vars(req)
	//nickname, ok := params["nickname"]
	nickname := ps.Get("nickname")
	if len(nickname) <= 0 {
		http.Error(w, "can't parse nickname", http.StatusBadRequest)
		return
	}
	user, err := db.SelectUser(nickname)
	if err != nil {
		Get404(w, "Can't find user by nickname: " + nickname)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(user)
}

func userCreate(w http.ResponseWriter,req *http.Request, ps denco.Params) {
	//params := mux.Vars(req)
	body, err := ioutil.ReadAll(req.Body)
	defer req.Body.Close()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	data := db.User{}
	//data.Nickname, _ = params["nickname"]
	data.Nickname = ps.Get("nickname")
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

	_, err = w.Write(output)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func userPost(w http.ResponseWriter, req *http.Request, ps denco.Params) {
	//params := mux.Vars(req)
	//nickname, _ := params["nickname"]
	nickname := ps.Get("nickname")
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
	output, err := json.Marshal(user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(output)
}

func UserHandler(router **denco.Mux) []denco.Handler {
	fmt.Println("user handlers initialized")
	//(*router).POST("/api/user/:nickname/create",  userCreate)
	//(*router).GET( "/api/user/:nickname/profile", userGet)
	//(*router).POST("/api/user/:nickname/profile", userPost)
	return []denco.Handler{
		(*router).POST("/api/user/:nickname/create",  userCreate),
		(*router).GET( "/api/user/:nickname/profile", userGet),
		(*router).POST("/api/user/:nickname/profile", userPost),
	}
}
