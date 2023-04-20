package utils

import (
	"encoding/json"
	"net/http"
)

func jsonResponse(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func ErrorJsonResponse(w http.ResponseWriter, message string, status int) {
	jsonResponse(w, status, map[string]string{"message": message})
}