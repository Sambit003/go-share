package utils

import (
	"encoding/json"
	"net/http"
)

func JsonResponse(w http.ResponseWriter, statusCode int, data interface{}) error {
	if data == nil {
		return nil
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	return json.NewEncoder(w).Encode(data)
}

func ErrorJsonResponse(w http.ResponseWriter, message string, statusCode int) error {
	return JsonResponse(w, statusCode, map[string]string{"error": message})
}