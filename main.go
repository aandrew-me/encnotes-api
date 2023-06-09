package main

import (
	"context"
	"crypto/rand"
	"math/big"
	"os"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/gofiber/storage/redis/v2"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
)

var client *mongo.Client

var REDIS_PORT int
var REDIS_HOST string
var REDIS_USERNAME string
var REDIS_PASSWORD string
var store *session.Store

func init() {
	godotenv.Load()
	REDIS_PORT, _ = strconv.Atoi(os.Getenv("REDIS_PORT"))
	REDIS_HOST = os.Getenv("REDIS_HOST")
	REDIS_USERNAME = os.Getenv("REDIS_USERNAME")
	REDIS_PASSWORD = os.Getenv("REDIS_PASSWORD")

	REDIS := redis.New(redis.Config{
		Host:     REDIS_HOST,
		Port:     REDIS_PORT,
		Username: REDIS_USERNAME,
		Password: REDIS_PASSWORD,
	})

	store = session.New(session.Config{
		Expiration:   time.Hour * 24 * 7,
		CookieSecure: true,
		Storage:      REDIS,
	})
}

func main() {
	connectDB()
	defer client.Disconnect(context.Background())
	app := fiber.New()

	limiterMiddleware := limiter.New(limiter.Config{
		Max:               20,
		Expiration:        30 * time.Second,
		LimiterMiddleware: limiter.SlidingWindow{},
	})

	codeCheckLimiter := limiter.New(limiter.Config{
		Max:        10,
		Expiration: 60 * time.Second,
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(429).JSON(fiber.Map{
				"message": "Too many requests.",
				"status":  "false",
			})
		},
	})

	emailVerificationLimiter := limiter.New(limiter.Config{
		Max:        1,
		Expiration: 60 * time.Second,
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(429).JSON(fiber.Map{
				"message": "Try again after a minute",
				"status":  "false",
			})
		},
	})

	standardLimiter := limiter.New(limiter.Config{
		Max:        200,
		Expiration: 60 * time.Second,
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(429).JSON(fiber.Map{
				"message": "Too many requests.",
				"status":  "false",
			})
		},
	})

	app.Use(cors.New(cors.Config{
		AllowOrigins:  "*",
		ExposeHeaders: "Authorization",
	}))

	app.Use(compress.New(compress.Config{
		Level: compress.LevelBestSpeed,
	}))

	app.Get("/", func(c *fiber.Ctx) error {
		return c.Status(200).SendString("Working fine")
	})
	api := app.Group("/api")

	api.Use("/login", limiterMiddleware)
	api.Use("/register", limiterMiddleware)
	api.Use("/sendEmail", emailVerificationLimiter)
	api.Use("/verify", codeCheckLimiter)
	api.Use("/notes", standardLimiter)
	api.Use("/changePassword", codeCheckLimiter)

	api.Post("/register", register)
	api.Post("/login", login)
	api.Post("/sendEmail", handleSendEmail)
	api.Get("/verify", verifyEmail)
	api.Get("/logout", AuthMiddleWare, logout)
	api.Post("/changePassword", AuthMiddleWare, changePassword)

	api.Get("/ping", func(c *fiber.Ctx) error {
		return c.Status(200).JSON(fiber.Map{
			"message": "pong",
		})
	})

	api.Get("/notes", AuthMiddleWare, getNotes)
	api.Get("/notes/:id", AuthMiddleWare, getNote)
	api.Post("/notes", AuthMiddleWare, addNote)
	api.Delete("/notes", AuthMiddleWare, deleteNote)
	api.Put("/notes", AuthMiddleWare, updateNote)

	api.Get("/info", AuthMiddleWare, info)

	PORT := ":8100"

	if os.Getenv("PORT") != "" {
		PORT = ":" + os.Getenv("PORT")
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
