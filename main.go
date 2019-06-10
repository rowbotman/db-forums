package main

import (
	"./db"
	"./handlers"
	"github.com/gorilla/mux"
	"gopkg.in/jackc/pgx.v2"
	"log"
	"net/http"
)

const (
	host     = "localhost"
	port     = 5432
	user     = "park"
	password = "admin"
	dbname   = "park_forum"
)

func main() {
	//psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
	//	"password=%s dbname=%s sslmode=disable",
	//	host, port, user, password, dbname)
	pgxConfig := pgx.ConnConfig{
		Host:     host,
		Port:     port,
		Database: dbname,
		User:     user,
		Password: password,
	}
	pgxConnPoolConfig := pgx.ConnPoolConfig{
		ConnConfig: pgxConfig,
		MaxConnections: 3, // \
		AfterConnect: nil, //  - delete it
		AcquireTimeout: 0, // /
	}
	var err error
	db.DB, err = pgx.NewConnPool(pgxConnPoolConfig)
	if err != nil {
		log.Fatal(err)
	}

	//db.DB, err = sql.Open("postgres", psqlInfo)
	//if err != nil {
	//	panic(err)
	//}
	defer db.DB.Close()

	//err = db.DB.Ping()
	//if err != nil {
	//	panic(err)
	//}

	router := mux.NewRouter()
	handlers.UserHandler(&router)
	handlers.ForumHandler(&router)
	handlers.PostHandler(&router)
	handlers.ServiceHandler(&router)
	handlers.ThreadHandler(&router)
	http.Handle("/",router)

	err = http.ListenAndServe(":5000", router)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
