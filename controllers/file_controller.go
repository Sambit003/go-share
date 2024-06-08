package controllers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"go-share/config"
	"go-share/models"
	"go-share/utils"
)

// RegisterFileRoutes registers the file-related API routes.
func RegisterFileRoutes(router *mux.Router) {
	// Apply authentication middleware to all file-related routes
	fileRouter := router.PathPrefix("/files").Subrouter()
	fileRouter.Use(utils.AuthMiddleware)

	fileRouter.HandleFunc("", CreateFile).Methods("POST")
	fileRouter.HandleFunc("", GetFiles).Methods("GET")
	fileRouter.HandleFunc("/{id}", GetFile).Methods("GET")
	fileRouter.HandleFunc("/{id}", UpdateFile).Methods("PUT")
	fileRouter.HandleFunc("/{id}", DeleteFile).Methods("DELETE")
}

// CreateFile handles file creation.
func CreateFile(w http.ResponseWriter, r *http.Request) {
	var file models.File
	if err := json.NewDecoder(r.Body).Decode(&file); err != nil {
		utils.ErrorJsonResponse(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Get the user ID from the request context
	//userID := r.Context().Value("user_id").(uint)

	token := r.Header.Get("Authorization")
	claims, err := utils.VerifyToken(token)
	if err != nil {
		utils.ErrorJsonResponse(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	file.UserID = claims.UserID
	if err := file.CreateFile(config.DB); err != nil {
		utils.ErrorJsonResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.JsonResponse(w, http.StatusCreated, file)
}

// GetFiles returns a list of all files.
// TODO: Add pagination and filtering for production.
func GetFiles(w http.ResponseWriter, r *http.Request) {
	var files []models.File
	if err := config.DB.Find(&files).Error; err != nil {
		utils.ErrorJsonResponse(w, "Error getting files", http.StatusInternalServerError)
		return
	}

	utils.JsonResponse(w, http.StatusOK, files)
}

// GetFile retrieves a single file by ID.
func GetFile(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, err := strconv.ParseUint(params["id"], 10, 64)
	if err != nil {
		utils.ErrorJsonResponse(w, "Invalid file ID", http.StatusBadRequest)
		return
	}

	var file models.File
	if err := config.DB.First(&file, id).Error; err != nil {
		utils.ErrorJsonResponse(w, "File not found", http.StatusNotFound)
		return
	}

	utils.JsonResponse(w, http.StatusOK, file)
}

// UpdateFile updates a file.
func UpdateFile(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, err := strconv.ParseUint(params["id"], 10, 64)
	if err != nil {
		utils.ErrorJsonResponse(w, "Invalid file ID", http.StatusBadRequest)
		return
	}

	token := r.Header.Get("Authorization")
	claims, err := utils.VerifyToken(token)
	if err != nil {
		utils.ErrorJsonResponse(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	var file models.File
	if err := config.DB.First(&file, id).Error; err != nil {
		utils.ErrorJsonResponse(w, "File not found", http.StatusNotFound)
		return
	}

	var updatedFile models.File
	if err := json.NewDecoder(r.Body).Decode(&updatedFile); err != nil {
		utils.ErrorJsonResponse(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := file.UpdateFile(config.DB, claims.UserID, &updatedFile); err != nil { 
		utils.ErrorJsonResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.JsonResponse(w, http.StatusOK, file)
}

// DeleteFile deletes a file.
func DeleteFile(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, err := strconv.ParseUint(params["id"], 10, 64)
	if err != nil {
		utils.ErrorJsonResponse(w, "Invalid file ID", http.StatusBadRequest)
		return
	}

	token := r.Header.Get("Authorization")
	claims, err := utils.VerifyToken(token)
	if err != nil {
		utils.ErrorJsonResponse(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	var file models.File
	if err := config.DB.First(&file, id).Error; err != nil {
		utils.ErrorJsonResponse(w, "File not found", http.StatusNotFound)
		return
	}

	if err := file.DeleteFile(config.DB, claims.UserID); err != nil {
		utils.ErrorJsonResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.JsonResponse(w, http.StatusOK, file) 
}