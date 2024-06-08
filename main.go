package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"go-share/config"
	"go-share/controllers"
	"go-share/models"
)

func main() {
	config.LoadConfig()      // Load configuration
	config.ConnectDB()       // Connect to database
	defer config.CloseDB()   // Close database connection

	router := mux.NewRouter()

	// Register routes
	controllers.RegisterAuthRoutes(router)
	controllers.RegisterFileRoutes(router)

	// AutoMigrate database (this should be done only once, usually during initial setup)
	if err := config.DB.AutoMigrate(&models.User{}, &models.File{}); err != nil {
		log.Fatalf("Error migrating database: %s", err)
	}

	fmt.Println("Server is running on port 8080")
	log.Fatal(http.ListenAndServe(":8080", router))
} 