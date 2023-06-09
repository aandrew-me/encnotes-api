package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func connectDB() {
	// MongoDB start
	var MONGO_URL = "mongodb://localhost:27017"
	if os.Getenv("MONGO_URL") != "" {
		MONGO_URL = os.Getenv("MONGO_URL")
	}
	var err error
	client, err = mongo.NewClient(options.Client().ApplyURI(MONGO_URL))
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()

	err = client.Connect(ctx)

	if err != nil {
		log.Fatal(err)
	}
	err = client.Ping(ctx, nil)

	if err != nil {
		log.Fatal("Failed to ping MongoDB:", err)
	} else {
		fmt.Println("Successfully pinged MongoDB")
	}
}
