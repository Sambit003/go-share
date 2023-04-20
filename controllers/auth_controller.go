package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"

	"go-share/config"
	"go-share/models"
	"go-share/utils"
)

func RegisterRoutes() *mux.Router {
	router := mux.NewRouter()
	router.HandleFunc("/register", Register).Methods("POST")
	router.HandleFunc("/login", Login).Methods("POST")

	return router
}

func Register(w http.ResponseWriter, r *http.Request) {
	var user models.User

	json.NewDecoder(r.Body).Decode(&user)

	hashedPassword, err := utils.HashPassword(user.Password)
	if err != nil {
		http.Error(w, "Error in password hashing", http.StatusInternalServerError)
		return
	}
	user.Password = string(hashedPassword)

	if err := config.DB.Create(&user).Error; err != nil {
		http.Error(w, "Error in creating user", http.StatusInternalServerError)
		return
	}

	token, err := utils.GenerateToken(user.ID)
	if err != nil {
		http.Error(w, "Error in generating token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(token)

}

func Login(w http.ResponseWriter, r *http.Request) {
	var user models.User

	json.NewDecoder(r.Body).Decode(&user)

	var foundUser models.User

	if err := config.DB.Where("email = ?", user.Email).First(&foundUser).Error; err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	if err := utils.ComparePassword(foundUser.Password, user.Password); err {
		http.Error(w, "Incorrect password", http.StatusUnauthorized)
		return
	}

	token, err := utils.GenerateToken(foundUser.ID)
	if err != nil {
		http.Error(w, "Error in generating token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(token)
}
