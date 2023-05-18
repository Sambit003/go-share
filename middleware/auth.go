package middleware

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

	err = models.CreateUser(&user)
	if err != nil {
		utils.ErrorJsonResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	token, err := utils.GenerateToken(user.ID)
	if err != nil {
		utils.ErrorJsonResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.JsonResponse(w, http.StatusOK, map[string]interface{}{"token": token})
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

	utils.JsonResponse(w, http.StatusOK, map[string]interface{}{"token": token})
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

	utils.JsonResponse(w, http.StatusOK, map[string]interface{}{"user": user})
}
