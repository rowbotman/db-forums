package utils

import "net/http"

func WriteData(writer http.ResponseWriter, status int, data []byte) {
	writer.Header().Set("content-type", "application/json")
	writer.WriteHeader(status)
	writer.Write(data)
}