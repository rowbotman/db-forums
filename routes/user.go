package routes

import (
	"context"
	"database/sql"
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

func SetUserRouter(router *httptreemux.TreeMux) {
	router.POST("/api/user/:nickname/create", userCreateHandler)
	router.GET("/api/user/:nickname/profile", userProfileHandler)
	router.POST("/api/user/:nickname/profile", postProfile)
}

// was used for debugging
func printAll(rows *sql.Rows) {
	for rows.Next() {
		var user models.User
		i := 0
		rows.Scan(&i, &user.About, &user.Email, &user.Fullname, &user.Nickname)
		fmt.Println(user.Email)
	}
	fmt.Println()
}

func userCreateHandler(writer http.ResponseWriter, request *http.Request, ps map[string]string) {
	nickname := ps["nickname"]

	body, err := ioutil.ReadAll(request.Body)
	defer request.Body.Close()
	if err != nil {
		http.Error(writer, err.Error(), 500)
		return
	}

	var user models.User
	err = json.Unmarshal(body, &user)
	if err != nil {
		http.Error(writer, err.Error(), 500)
		return
	}
	user.Nickname = nickname
	db := db2.GetDB()
	_, err = db.Exec(context.Background(), "INSERT INTO users (about, email, fullname, nickname) " +
		"VALUES ($1, $2, $3, $4)",
		user.About, user.Email, user.Fullname, user.Nickname)
	if err == nil {
		data, err := json.Marshal(user)
		if err != nil {
			http.Error(writer, err.Error(), 500)
			return
		}
		writer.Header().Set("content-type", "application/json")
		writer.WriteHeader(201)
		writer.Write(data)
	} else {
		rows, err := db.Query(context.Background(), "SELECT about, email, fullname, nickname " +
			"FROM users WHERE nickname = $1 OR email = $2",
			user.Nickname, user.Email)
		if err != nil {
			log.Println("SELECT about, email, fullname, nickname " +
				"FROM users WHERE nickname = $1 OR email = $2", user.Nickname, user.Email)
			http.Error(writer, err.Error(), 500)
			return
		}
		defer rows.Close()

		conflicts := []byte{'['}
		for rows.Next() {
			if len(conflicts) > 1 {
				conflicts = append(conflicts, ',')
			}
			var u models.User
			_ = rows.Scan(&u.About, &u.Email, &u.Fullname, &u.Nickname)
			data, err := json.Marshal(u)
			if err != nil {
				http.Error(writer, err.Error(), 500)
				return
			}
			conflicts = append(conflicts, data...)
		}
		conflicts = append(conflicts, ']')

		writer.Header().Set("content-type", "application/json")
		writer.WriteHeader(409)
		writer.Write(conflicts)
	}
}

func userProfileHandler(writer http.ResponseWriter, request *http.Request, ps map[string]string) {
	nickname := ps["nickname"]
	db := db2.GetDB()
	row := db.QueryRow(context.Background(), "SELECT about, email, fullname, nickname " +
		"FROM users WHERE nickname = $1", nickname)

	var user models.User
	err := row.Scan(&user.About, &user.Email, &user.Fullname, &user.Nickname)
	if err != nil {
		msg, _ := json.Marshal(map[string]string{"message": "404"})
		utils.WriteData(writer, 404, msg)
	} else {
		data, err := json.Marshal(user)
		if err != nil {
			http.Error(writer, err.Error(), 500)
			return
		}
		utils.WriteData(writer, 200, data)
	}
}

func postProfile(writer http.ResponseWriter, request *http.Request, ps map[string]string) {
	nickname := ps["nickname"]
	db := db2.GetDB()
	// read body
	body, err := ioutil.ReadAll(request.Body)
	defer request.Body.Close()
	if err != nil {
		http.Error(writer, err.Error(), 500)
	}
	// parse body
	var user models.User
	err = json.Unmarshal(body, &user)
	if err != nil {
		http.Error(writer, err.Error(), 500)
	}
	user.Nickname = nickname
	// get current data
	var oldUser models.User
	err = db.QueryRow(context.Background(), "SELECT about, email, fullname, nickname FROM users " +
		"WHERE nickname = $1", user.Nickname).Scan(&oldUser.About, &oldUser.Email, &oldUser.Fullname, &oldUser.Nickname)
	if err != nil {
		msg, _ := json.Marshal(map[string]string{"message": "User not found"})
		utils.WriteData(writer, 404, msg)
		return
	}
	// check empty request
	if user.Email == "" && user.Fullname == "" && user.About == "" {
		data, err := json.Marshal(oldUser)
		if err != nil {
			http.Error(writer, err.Error(), 500)
		}
		utils.WriteData(writer, 200, data)
		return
	}
	// check empty fields
	if user.Fullname == "" {
		user.Fullname = oldUser.Fullname
	}
	if user.Email == "" {
		user.Email = oldUser.Email
	}
	if user.About == "" {
		user.About = oldUser.About
	}

	result, err := db.Exec(context.Background(), "UPDATE users " +
		"SET about = $1, email = $2, fullname = $3 " +
		"WHERE  nickname = $4", user.About, user.Email, user.Fullname, user.Nickname)
	// user with new email already exist
	if err != nil {
		msg, _ := json.Marshal(map[string]string{"message": "conflict"})
		utils.WriteData(writer, 409, msg)
		return
	}

	number := result.RowsAffected()
	if number == 0 {
		msg, _ := json.Marshal(map[string]string{"message": "User not found"})
		utils.WriteData(writer, 404, msg)
		return
	} else {
		data, err := json.Marshal(user)
		if err != nil {
			http.Error(writer, err.Error(), 500)
		}
		utils.WriteData(writer, 200, data)
		return
	}
}
