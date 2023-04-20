package main

import (
	"fmt"
	"log"
	"net/http"

	"go-share/config"
	"go-share/controllers"
)

func main() {
	router := controllers.RegisterRoutes()
	config.Load()
	defer config.Close()
	fmt.Println("Server is running on port 8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}
