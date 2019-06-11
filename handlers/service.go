package handlers

import (
	"../db"
	"encoding/json"
	"fmt"
	"github.com/naoina/denco"
	"net/http"
)

func serviceDrop(w http.ResponseWriter, _ *http.Request, _ denco.Params) {
	w.Header().Set("content-type", "text/plain")
	if db.ClearService() {
		_, _ = w.Write([]byte("Отчистка базы успешно завершена"))
		return
	}
	_, _ = w.Write([]byte("error occurred"))
}

func serviceGetInfo(w http.ResponseWriter, _ *http.Request, _ denco.Params) {
	w.Header().Set("content-type", "text/plain")
	status, err := db.ServiceGet()
	if err != nil {
		return
	}

	output, err := json.Marshal(status)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("content-type", "application/json")
	_, _ = w.Write(output)
}

func ServiceHandler(router **denco.Mux) []denco.Handler {
	fmt.Println("services handlers initialized")
	return []denco.Handler{
		(*router).POST("/api/service/clear",  serviceDrop),
		(*router).GET( "/api/service/status", serviceGetInfo)}
}
