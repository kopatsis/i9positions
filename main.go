package main

import (
	"context"
	"encoding/base64"
	"i9-pos/database"
	"i9-pos/platform"
	"log"
	"net/http"
	"os"

	firebase "firebase.google.com/go"
	"go.etcd.io/bbolt"
	"google.golang.org/api/option"

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

	firebaseConfigBase64 := os.Getenv("FIREBASE_CONFIG_BASE64")
	if firebaseConfigBase64 == "" {
		log.Fatal("FIREBASE_CONFIG_BASE64 environment variable is not set.")
	}

	configJSON, err := base64.StdEncoding.DecodeString(firebaseConfigBase64)
	if err != nil {
		log.Fatalf("Error decoding FIREBASE_CONFIG_BASE64: %v", err)
	}

	sa := option.WithCredentialsJSON(configJSON)
	firebase, err := firebase.NewApp(context.Background(), nil, sa)
	if err != nil {
		log.Fatalf("error initializing app: %v\n", err)
	}

	boltDB, err := bbolt.Open("cache.db", 0666, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer boltDB.Close()

	rtr := platform.New(db, firebase, boltDB)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	if err := http.ListenAndServe(":"+port, rtr); err != nil {
		log.Fatalf("There was an error with the http server: %v", err)
	}
}
