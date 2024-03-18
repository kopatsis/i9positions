package main

import (
	"i9-pos/database"
	"i9-pos/platform"
	"log"
	"net/http"

	"github.com/joho/godotenv"
)

func main() {

	if err := godotenv.Load(); err != nil {
		log.Fatalf("Failed to load the env vars: %v", err)
	}

	client, db, err := database.ConnectDB()
	if err != nil {
		log.Fatalf("Error while connecting to mongoDB: %s.\nExiting.", err)
	}
	defer database.DisConnectDB(client)

	rtr := platform.New(db)

	log.Print("Server listening on http://localhost:3500/")
	if err := http.ListenAndServe("0.0.0.0:3500", rtr); err != nil {
		log.Fatalf("There was an error with the http server: %v", err)
	}
}
