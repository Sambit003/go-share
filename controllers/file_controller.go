package controllers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"go-share/config"
	"go-share/pkg/files"
	"go-share/utils"
	"io"
	"os"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	"github.com/spf13/viper"
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
	// Parse multipart form (max 10MB)
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		utils.ErrorJsonResponse(w, "Error parsing multipart form", http.StatusBadRequest)
		return
	}

	formFile, handler, err := r.FormFile("file") // "file" is the form field name
	if err != nil {
		utils.ErrorJsonResponse(w, "Error retrieving the file from form", http.StatusBadRequest)
		return
	}
	defer formFile.Close()

	fileName := handler.Filename
	contentType := handler.Header.Get("Content-Type")
	description := r.FormValue("description") // Get description from form value

	token := r.Header.Get("Authorization")
	claims, err := utils.VerifyToken(token)
	if err != nil {
		utils.ErrorJsonResponse(w, "Invalid token", http.StatusUnauthorized)
		return
	}
	userID := claims.UserID

	// Get storagePathBase from config, with a default
	storagePathBase := viper.GetString("storage.base_path")
	if storagePathBase == "" { // Should be set in main.go or config file
		storagePathBase = "./uploads" // Fallback default
	}

	// Ensure the user-specific directory exists (UploadFile will handle this, but good to be aware)
	// For example: userSpecificPath := filepath.Join(storagePathBase, "user_"+strconv.Itoa(int(userID)))
	// os.MkdirAll(userSpecificPath, os.ModePerm)

	// Get encryption key from header
	encryptionKeyHeader := r.Header.Get("X-Encryption-Key")
	var encryptionKey []byte
	if encryptionKeyHeader != "" {
		encryptionKey = []byte(encryptionKeyHeader)
	}

	// Call the new library function
	newFile, err := files.UploadFile(config.DB, formFile, fileName, contentType, description, userID, storagePathBase, encryptionKey)
	if err != nil {
		// Check if the error is a validation error from go-playground/validator
		if _, ok := err.(validator.ValidationErrors); ok {
			utils.ErrorJsonResponse(w, "Validation failed: "+err.Error(), http.StatusBadRequest)
		} else {
			utils.ErrorJsonResponse(w, "Error uploading file: "+err.Error(), http.StatusInternalServerError)
		}
		return
	}

	utils.JsonResponse(w, http.StatusCreated, newFile)
}

// GetFiles returns a list of all files.
// TODO: Add pagination and filtering for production.
func GetFiles(w http.ResponseWriter, r *http.Request) {
	var filesResponse []files.File // Changed to files.File and renamed variable
	if err := config.DB.Find(&filesResponse).Error; err != nil {
		utils.ErrorJsonResponse(w, "Error getting files", http.StatusInternalServerError)
		return
	}

	utils.JsonResponse(w, http.StatusOK, filesResponse)
}

// GetFile retrieves a single file by ID for download.
func GetFile(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	fileID, err := strconv.ParseUint(params["id"], 10, 64)
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
	userID := claims.UserID

	// Get decryption key from header
	decryptionKeyHeader := r.Header.Get("X-Decryption-Key")
	var decryptionKey []byte
	if decryptionKeyHeader != "" {
		decryptionKey = []byte(decryptionKeyHeader)
	}

	openedFile, fileMetadata, err := files.DownloadFile(config.DB, uint(fileID), userID, decryptionKey)
	if err != nil {
		// errMsg := err.Error() // No longer needed for direct string comparison
		switch {
		case errors.Is(err, files.ErrFileNotFound):
			utils.ErrorJsonResponse(w, "File not found", http.StatusNotFound)
		case errors.Is(err, files.ErrUnauthorized):
			utils.ErrorJsonResponse(w, "Forbidden: You don't have permission to access this file", http.StatusForbidden)
		case errors.Is(err, files.ErrDecryptionKeyRequired):
			utils.ErrorJsonResponse(w, "File is encrypted, decryption key required in X-Decryption-Key header", http.StatusBadRequest)
		// Separate cases for decryption failures for more specific client feedback.
		case errors.Is(err, files.ErrInvalidKeyLength): // Assuming ErrInvalidKeyLength is defined in files package
			utils.ErrorJsonResponse(w, "Failed to decrypt file: Invalid decryption key length. Must be 16, 24, or 32 bytes.", http.StatusBadRequest)
		case strings.Contains(err.Error(), "cipher: message authentication failed"): // Specific check for GCM auth failure if not covered by a sentinel error
			utils.ErrorJsonResponse(w, "Failed to decrypt file: Invalid or incorrect decryption key (authentication failed)", http.StatusUnauthorized)
		default:
			// Log the actual error for server-side debugging
			// log.Printf("Error retrieving file %d: %v", fileID, err)
			utils.ErrorJsonResponse(w, "Error retrieving file: "+err.Error(), http.StatusInternalServerError)
		}
		return
	}
	defer openedFile.Close()

	// Set headers for download
	w.Header().Set("Content-Disposition", "attachment; filename=\""+fileMetadata.Name+"\"")
	contentType := fileMetadata.ContentType
	if contentType == "" {
		contentType = "application/octet-stream" // Default content type
	}
	w.Header().Set("Content-Type", contentType)

	// Set Content-Length (optional but good practice)
	// This is more complex now since openedFile is an io.ReadCloser.
	// We can only reliably get the size if it's an *os.File (non-encrypted).
	// For encrypted files, the original size isn't directly available from the stream
	// without reading it or storing it separately.
	if !fileMetadata.IsEncrypted {
		if f, ok := openedFile.(*os.File); ok { // Check if it's an os.File
			fileInfo, err := f.Stat()
			if err == nil {
				w.Header().Set("Content-Length", strconv.FormatInt(fileInfo.Size(), 10))
			}
		}
	}
	// If it's encrypted, Content-Length might not be set, or we'd need to read the
	// decrypted content into a buffer first to get its length, which is less efficient for streaming.

	// Stream the file
	_, err = io.Copy(w, openedFile)
	if err != nil {
		// Log the error, but the headers might have already been sent
		// so we can't send a JSON error response easily.
		http.Error(w, "Error streaming file: "+err.Error(), http.StatusInternalServerError)
	}
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

	var file files.File // Changed to files.File
	if err := config.DB.First(&file, id).Error; err != nil {
		utils.ErrorJsonResponse(w, "File not found", http.StatusNotFound)
		return
	}

	var updatedFile files.File // Changed to files.File
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

	var file files.File // Changed to files.File
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
