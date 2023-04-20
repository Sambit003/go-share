package controllers

import (
	"net/http"
	"encoding/json"
	"github.com/gorilla/mux"

	"go-share/config"
	"go-share/models"
	"go-share/utils"
)

func RegisterFileRoutes() *mux.Router {
	router := mux.NewRouter()
	router.HandleFunc("/files", CreateFiles).Methods("POST")
	router.HandleFunc("/files", GetFiles).Methods("GET")
	router.HandleFunc("/files/{id}", GetFile).Methods("GET")
	router.HandleFunc("/files/{id}", UpdateFile).Methods("PUT")
	router.HandleFunc("/files/{id}", DeleteFile).Methods("DELETE")

	return router
}

func CreateFiles(w http.ResponseWriter, r *http.Request) {
	var file models.File

	json.NewDecoder(r.Body).Decode(&file)

	token := r.Header.Get("Authorization")
	claims, err := utils.VerifyToken(token)
	if err != nil {
        http.Error(w, "Invalid token", http.StatusUnauthorized)
        return
    }

	file.UserID = claims.UserID
	if err := config.DB.Create(&file).Error; err != nil {
		http.Error(w, "Error in creating file", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(file)
}

func GetFiles(w http.ResponseWriter, r *http.Request) {
	var files []models.File

	if err := config.DB.Find(&files).Error; err != nil {
		http.Error(w, "Error in getting files", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(files)
}

func GetFile(w http.ResponseWriter, r *http.Request) {
	var file models.File

	params := mux.Vars(r)
	if err := config.DB.Where("id = ?", params["id"]).First(&file).Error; err != nil {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(file)
}

func UpdateFile(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
    id := params["id"]

    var file models.File
    if err := config.DB.First(&file, id).Error; err != nil {
        http.Error(w, "File not found", http.StatusNotFound)
        return
    }

    token := r.Header.Get("Authorization")
    claims, err := utils.VerifyToken(token)
    if err != nil {
        http.Error(w, "Invalid token", http.StatusUnauthorized)
        return
    }

    if file.UserID != claims.UserID {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }

    var updatedFile models.File
    json.NewDecoder(r.Body).Decode(&updatedFile)

    if updatedFile.Name != "" {
        file.Name = updatedFile.Name
    }
    if updatedFile.ContType != "" {
        file.ContType = updatedFile.ContType
    }
    if updatedFile.Path != "" {
        file.Path = updatedFile.Path
    }

    if err := config.DB.Save(&file).Error; err != nil {
        http.Error(w, "Error updating file", http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(file)
}

func DeleteFile(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id := params["id"]

	var file models.File
	if err := config.DB.First(&file, id).Error; err != nil {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	token := r.Header.Get("Authorization")
	claims, err := utils.VerifyToken(token)
	if err != nil {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	if file.UserID != claims.UserID {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if err := config.DB.Delete(&file).Error; err != nil {
		http.Error(w, "Error deleting file", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(file)
}