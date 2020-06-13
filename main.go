package main

import (
	"forum/routes"
	"github.com/dimfeld/httptreemux"
	"log"
	"net/http"
)

func main() {
	router := httptreemux.New()

	routes.SetHomeRouter(router)
	routes.SetForumRouter(router)
	routes.SetServiceRouter(router)
	routes.SetPostRouter(router)
	routes.SetThreadRouter(router)
	routes.SetUserRouter(router)

	server := http.Server{
		Addr:    ":5000",
		Handler: router,
	}

	err := server.ListenAndServe()
	if err != nil {
		log.Fatal(err.Error())
	}
}
