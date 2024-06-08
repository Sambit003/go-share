package utils

import (
	"encoding/json"
	"net/http"
)

// JsonResponse sends a JSON response with the provided status code and data.
func JsonResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		// Handle the encoding error, possibly logging it. 
		// For now, we'll just write a basic error response.
		http.Error(w, "Error encoding JSON", http.StatusInternalServerError)
	}
}

// ErrorJsonResponse sends a JSON error response.
func ErrorJsonResponse(w http.ResponseWriter, message string, statusCode int) {
	JsonResponse(w, statusCode, map[string]string{"error": message})
}