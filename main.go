package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"go-share/config"
	"go-share/controllers"
	"go-share/models" // Keep models for User
	"go-share/pkg/files" // Add new import for files.File
	"github.com/spf13/viper" // Added viper
	"os"                     // Added os
	// "path/filepath"       // Removed as it's no longer directly used here
)

func main() {
	// Set default for storage base path before loading config
	viper.SetDefault("storage.base_path", "./uploads")

	config.LoadConfig()      // Load configuration
	config.ConnectDB()       // Connect to database
	defer config.CloseDB()   // Close database connection

	// Create storage base path directory if it doesn't exist
	storageBasePath := viper.GetString("storage.base_path")
	if err := os.MkdirAll(storageBasePath, os.ModePerm); err != nil {
		log.Fatalf("Error creating storage base path directory: %s", err)
	}


	router := mux.NewRouter()

	// Register routes
	controllers.RegisterAuthRoutes(router)
	controllers.RegisterFileRoutes(router)

	// AutoMigrate database (this should be done only once, usually during initial setup)
	if err := config.DB.AutoMigrate(&models.User{}, &files.File{}); err != nil { // Changed models.File to files.File
		log.Fatalf("Error migrating database: %s", err)
	}

	fmt.Println("Server is running on port 8080")
	log.Fatal(http.ListenAndServe(":8080", router))
} 