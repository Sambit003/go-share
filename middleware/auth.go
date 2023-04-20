package controllers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"go-share/models"
	"go-share/utils"

	"github.com/gorilla/mux"
)

type AuthController struct{}

func (ac AuthController) Register(w http.ResponseWriter, r *http.Request) {
	var user models.User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		utils.ErrorJsonResponse(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err = models.User(&user)
	if err != nil {
		utils.ErrorJsonResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	token, err := utils.GenerateToken(user.ID)
	if err != nil {
		utils.ErrorJsonResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.jsonResponse(w, http.StatusOK, map[string]string{"token": token})
}

func (ac AuthController) Login(w http.ResponseWriter, r *http.Request) {
	var user models.User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		utils.ErrorJsonResponse(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	validUser, err := models.ValidateUser(&user)
	if err != nil {
		utils.ErrorJsonResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	token, err := utils.GenerateToken(validUser.ID)
	if err != nil {
		utils.ErrorJsonResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.jsonResponse(w, map[string]string{"token": token}, http.StatusOK)
}

func (ac AuthController) GetUser(w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.ParseUint(mux.Vars(r)["id"], 10, 32)
	if err != nil {
		utils.ErrorJsonResponse(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	user, err := models.GetUser(uint(userID))
	if err != nil {
		utils.ErrorJsonResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	utils.jsonResponse(w, user, http.StatusOK)
}