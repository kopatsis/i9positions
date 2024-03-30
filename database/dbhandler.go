package database

import (
	"context"
	"log"
	"os"

	"go.mongodb.org/mongo-driver/mongo"

	"go.mongodb.org/mongo-driver/mongo/options"
)

func DisConnectDB(client *mongo.Client) {
	err := client.Disconnect(context.Background())
	if err != nil {
		log.Fatal(err)
	}
}

func ConnectDB() (*mongo.Client, *mongo.Database, error) {

	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	connectStr := os.Getenv("MONGOSTRING")
	clientOptions := options.Client().ApplyURI(connectStr).SetServerAPIOptions(serverAPI)

	// Connect to MongoDB
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		log.Fatal(err)
		return nil, nil, err
	}

	// Check the connection
	err = client.Ping(context.Background(), nil)
	if err != nil {
		log.Fatal(err)
		return nil, nil, err
	}

	// Specify the database and collection
	database := client.Database("i9pos")

	return client, database, nil
}
