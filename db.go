package main

import (
	"context"
	"fmt"
	"log"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)


const MONGO_URL = "mongodb://localhost:27017"


func connectDB() {
	// MongoDB start
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
