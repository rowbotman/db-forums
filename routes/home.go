package routes

import (
	"github.com/dimfeld/httptreemux"
	"net/http"
)

func SetHomeRouter(router *httptreemux.TreeMux) {
	router.GET("/api", homeHandler)
}

func homeHandler(writer http.ResponseWriter, request *http.Request, ps map[string]string) {
	writer.WriteHeader(200)
}