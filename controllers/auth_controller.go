package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"go-share/config"
	"go-share/models"
	"go-share/utils"
)

// RegisterAuthRoutes registers the authentication routes with the provided router.
func RegisterAuthRoutes(router *mux.Router) {
	router.HandleFunc("/register", Register).Methods("POST")
	router.HandleFunc("/login", Login).Methods("POST")
}

// Register handles user registration.
func Register(w http.ResponseWriter, r *http.Request) {
	var user models.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		utils.ErrorJsonResponse(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := user.CreateUser(config.DB); err != nil {
		utils.ErrorJsonResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	token, err := utils.GenerateToken(user.ID)
	if err != nil {
		utils.ErrorJsonResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.JsonResponse(w, http.StatusCreated, map[string]string{"token": token})
}

// Login handles user login.
func Login(w http.ResponseWriter, r *http.Request) {
	var user models.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		utils.ErrorJsonResponse(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	foundUser, err := user.ValidateUserCredentials(config.DB) 
	if err != nil {
		utils.ErrorJsonResponse(w, err.Error(), http.StatusUnauthorized) // Use StatusUnauthorized for auth errors
		return
	}

	token, err := utils.GenerateToken(foundUser.ID)
	if err != nil {
		utils.ErrorJsonResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.JsonResponse(w, http.StatusOK, map[string]string{"token": token}) 
}