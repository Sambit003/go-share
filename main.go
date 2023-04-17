package main

import (
	"github.com/Sambit003/go-share/controllers"
	"fmt"
	"log"
	"net/http"

	"honnef.co/go/tools/config"
)

func main() {
	router := controllers.RegisterRoutes()
	config.Load()
	defer config.Close()
	fmt.Println("Server is running on port 8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}
