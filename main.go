package main

import (
	"context"
	"crypto/rand"
	"math/big"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"go.mongodb.org/mongo-driver/mongo"
)

var client *mongo.Client

func main() {
	connectDB()
	defer client.Disconnect(context.Background())
	app := fiber.New()

	app.Use(limiter.New(limiter.Config{
		Max:               20,
		Expiration:        30 * time.Second,
		LimiterMiddleware: limiter.SlidingWindow{},
	}))

	api := app.Group("/api")

	api.Post("/register", register)
	api.Post("/login", login)

	api.Get("/ping", func(c *fiber.Ctx) error {
		return c.Status(200).JSON(fiber.Map{
			"message":"pong",
		})
	})

	api.Get("/notes", AuthMiddleWare, getNote)
	api.Post("/notes", AuthMiddleWare, addNote)
	api.Delete("/notes", AuthMiddleWare, deleteNote)

	api.Post("/info", AuthMiddleWare, info)

	PORT := ":8100"

	if os.Getenv("PORT") != "" {
		PORT = os.Getenv("PORT")
	}

	app.Listen(PORT)
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
