package main

import (
	"fmt"
	"log"
	"net/http"

	"go-share/config"
	"go-share/controllers"
	"go-share/models"
	"go-share/pkg/files"
	"os"

	"github.com/gorilla/mux"
	"github.com/spf13/viper"
)

func main() {
	// Set default for storage base path before loading config
	viper.SetDefault("storage.base_path", "./uploads")

	config.LoadConfig()    // Load configuration
	config.ConnectDB()     // Connect to database
	defer config.CloseDB() // Close database connection

	// Create storage base path directory if it doesn't exist
	storageBasePath := viper.GetString("storage.base_path")
	if err := os.MkdirAll(storageBasePath, 0750); err != nil { // Changed from os.ModePerm to 0750
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
