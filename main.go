package main

import (
	"context"
	"crypto/rand"
	"math/big"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/mongo"
)

var client *mongo.Client

func main() {
	connectDB()
	defer client.Disconnect(context.Background())
	app := fiber.New()

	api := app.Group("/api")

	api.Post("/register", register)
	api.Post("/login", login)

	api.Get("/notes", AuthMiddleWare, getNote)
	api.Post("/notes", AuthMiddleWare, addNote)

	api.Post("/info", AuthMiddleWare, info)

	app.Listen(":3000")
}

func GenerateRandomString(n int) (string, error) {
	const letters = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz-"
	ret := make([]byte, n)
	for i := 0; i < n; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		if err != nil {
			return "", err
		}
		ret[i] = letters[num.Int64()]
	}

	return string(ret), nil
}
