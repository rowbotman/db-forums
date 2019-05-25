package main

import (
	"./db"
	"./handlers"
	"database/sql"
	"fmt"
	"github.com/gorilla/mux"
	"log"      // пакет для логирования
	"net/http" // пакет для поддержки HTTP протокола
)
const (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = "admin"
	dbname   = "park_forum"
)


func main() {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)
	var err error
	db.DB, err = sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}
	defer db.DB.Close()

	err = db.DB.Ping()
	if err != nil {
		panic(err)
	}
	router := mux.NewRouter()
	handlers.UserHandler(&router)
	handlers.ForumHandler(&router)
	handlers.PostHandler(&router)
	handlers.ServiceHandler(&router)
	handlers.ThreadHandler(&router)
	//router.HandleFunc("/api/post/{id}/details/", handlers.PostGetInfo).Methods("GET")
	http.Handle("/",router)


	err = http.ListenAndServe(":5000", router)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}

	//db.InitDB(connectStr)
	//defer db.DB.Close()

}