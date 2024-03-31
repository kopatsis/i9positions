package main

import (
	"i9-pos/database"
	"i9-pos/platform"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

func main() {

	if err := godotenv.Load(); err != nil {
		if os.Getenv("APP_ENV") != "production" {
			log.Fatalf("Failed to load the env vars: %v", err)
		}
	}

	client, db, err := database.ConnectDB()
	if err != nil {
		log.Fatalf("Error while connecting to mongoDB: %s.\nExiting.", err)
	}
	defer database.DisConnectDB(client)

	rtr := platform.New(db)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	if err := http.ListenAndServe(":"+port, rtr); err != nil {
		log.Fatalf("There was an error with the http server: %v", err)
	}
}
