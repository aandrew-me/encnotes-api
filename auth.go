package main

import (
	"context"
	"time"

	"github.com/go-playground/validator"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
)

type UserRegister struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6,max=32"`
}

type User struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6,max=32"`
	UserID   string `json:"userID" validate:"required" bson:"userID"`
	Notes    []Note `json:"notes"`
}

// Register
func register(c *fiber.Ctx) error {
	var db = client.Database("enotesdb")
	var userCollection = db.Collection("users")
	c.Accepts("application/json")
	var user UserRegister

	err := c.BodyParser(&user)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "something went wrong: " + err.Error(),
			"status":  "false",
		})
	}

	err = validate.Struct(&user)

	if err != nil {
		return c.Status(fiber.StatusNotAcceptable).JSON(fiber.Map{
			"status":  "false",
			"message": "Make sure password and email are following correct rules",
		})
	}

	// Checking if email already exists

	result := userCollection.FindOne(c.Context(), fiber.Map{"email": user.Email})
	var userResult UserRegister
	result.Decode(&userResult)
	if result.Err() == nil {
		return c.Status(fiber.StatusNotAcceptable).JSON(fiber.Map{
			"status":  "false",
			"message": "Email already in use",
		})

	}
	randomString, err := GenerateRandomString(20)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "something went wrong: " + err.Error(),
			"status":  "false",
		})
	}
	finalUser := User{
		Email:    user.Email,
		Password: user.Password,
		Notes:    []Note{},
		UserID:   randomString,
	}

	_, err = userCollection.InsertOne(context.Background(), finalUser)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "something went wrong: " + err.Error(),
			"status":  "false",
		})
	}
	sess, err := store.Get(c)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "something went wrong: " + err.Error(),
			"status":  "false",
		})
	}

	sess.Set("userID", finalUser.UserID)

	if err := sess.Save(); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to set session cookie" + err.Error(),
			"status":  "false",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  "true",
		"message": "Account Created",
	})

}

// Auth
var validate = validator.New()
var store = session.New(session.Config{
	CookieHTTPOnly: true,
	Expiration:     time.Hour * 24 * 7,
})

func AuthMiddleWare(c *fiber.Ctx) error {
	sess, err := store.Get(c)

	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "not authorized",
			"status":  "false",
		})
	}

	userID := sess.Get("userID")

	if userID == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "not authorized",
			"status":  "false",
		})
	}

	c.Locals("userID", userID)
	return c.Next()
}

func info(c *fiber.Ctx) error {
	userID := c.Locals("userID")

	if userID == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "not authorized",
			"status":  "false",
		})
	}

	var db = client.Database("enotesdb")
	var userCollection = db.Collection("users")

	result := userCollection.FindOne(context.Background(), fiber.Map{
		"userID": userID,
	})

	if result.Err() != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "something went wrong: ",
			"status":  "false",
		})
	}

	var user UserRegister
	result.Decode(&user)

	return c.Status(200).JSON(user)
}

func login(c *fiber.Ctx) error {
	var db = client.Database("enotesdb")
	var userCollection = db.Collection("users")
	c.Accepts("application/json")

	sess, _ := store.Get(c)
	userID := sess.Get("userID")

	if userID != nil {
		return c.Status(fiber.StatusAccepted).JSON(fiber.Map{
			"message": "Already logged in",
			"status":  "true",
		})
	}

	var user UserRegister

	err := c.BodyParser(&user)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "something went wrong: " + err.Error(),
			"status":  "false",
		})
	}

	err = validate.Struct(&user)

	if err != nil {
		return c.Status(fiber.StatusNotAcceptable).JSON(fiber.Map{
			"status":  "false",
			"message": "Make sure password and email are following correct rules",
		})
	}

	result := userCollection.FindOne(context.Background(), fiber.Map{
		"email": user.Email,
	})

	if result.Err() != nil {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"status":  "false",
			"message": "The account does not exist",
		})
	}
	var userResult User
	result.Decode(&userResult)

	if userResult.Password != user.Password {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"status":  "false",
			"message": "Incorrect password",
		})
	}

	sess, error := store.Get(c)

	if error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "something went wrong: " + err.Error(),
			"status":  "false",
		})
	}
	sess.Set("userID", userResult.UserID)
	sess.Save()
	c.Locals("userID", userResult.UserID)
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Successfully logged in. Redirect to home page",
		"status":  "true",
	})

}
