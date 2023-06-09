package main

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/go-playground/validator"
	"github.com/gofiber/fiber/v2"
)

type UserRegister struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6,max=128"`
	Captcha  string `json:"captcha" validate:"required"`
}

type UserLogin struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6,max=128"`
}

type User struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6,max=128"`
	UserID   string `json:"userID" validate:"required" bson:"userID"`
	Notes    []Note `json:"notes"`
	Verified bool   `json:"verified"`
}

type PasswordChange struct {
	CurrentPassword string `json:"currentPassword"`
	NewPassword     string `json:"newPassword"`
}

// Register
func register(c *fiber.Ctx) error {
	var db = client.Database("enotesdb")
	var userCollection = db.Collection("users")
	c.Accepts("application/json")
	var user UserRegister

	err := c.BodyParser(&user)

	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Error parsing JSON. Make sure you are sending proper JSON.",
			"status":  "false",
		})
	}

	err = validate.Struct(&user)

	if err != nil {
		return c.Status(fiber.StatusNotAcceptable).JSON(fiber.Map{
			"status":  "false",
			"message": "Make sure password and email are following correct rules and captcha is present.",
		})
	}

	captchaCorrect := verifyCaptcha(user.Captcha)

	if !captchaCorrect {
		return c.Status(fiber.StatusNotAcceptable).JSON(fiber.Map{
			"status":  "false",
			"message": "Incorrect Captcha",
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
		Email:    strings.ToLower(user.Email),
		Password: user.Password,
		Notes:    []Note{},
		UserID:   randomString,
		Verified: false,
	}

	_, err = userCollection.InsertOne(context.Background(), finalUser)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "something went wrong: " + err.Error(),
			"status":  "false",
		})
	}
	var codeCollection = db.Collection("codes")

	code := fmt.Sprint(rand.Uint64())

	go codeCollection.InsertOne(context.Background(), fiber.Map{
		"email":     strings.ToLower(user.Email),
		"code":      code,
		"createdAt": time.Now(),
	})

	go sendEmail(strings.ToLower(user.Email), code)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  "true",
		"message": "Account Created. A verification link has been sent to your email.",
	})

}

// Auth

var validate = validator.New()

func AuthMiddleWare(c *fiber.Ctx) error {
	// Set cookie header from authorization header
	auth_header := c.Get("Authorization")
	c.Request().Header.Set("Cookie", auth_header)

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
			"message": "Something went wrong",
			"status":  "false",
		})
	}

	var user User
	result.Decode(&user)

	return c.Status(200).JSON(user)
}

func login(c *fiber.Ctx) error {
	var db = client.Database("enotesdb")
	var userCollection = db.Collection("users")
	c.Accepts("application/json")

	// Set cookie header from authorization header
	auth_header := c.Get("Authorization")
	c.Request().Header.Set("Cookie", auth_header)

	sess, _ := store.Get(c)
	userID := sess.Get("userID")

	if userID != nil {
		return c.Status(fiber.StatusAccepted).JSON(fiber.Map{
			"message": "Already logged in",
			"status":  "true",
		})
	}

	var user UserLogin

	err := c.BodyParser(&user)

	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
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
			"message": "Your password is incorrect",
		})
	}

	if !userResult.Verified {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"status":  "false",
			"message": "You need to verify your Email",
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

	session_id := c.GetRespHeader("Set-Cookie")
	c.Response().Header.Set("Authorization", session_id)
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Successfully logged in. Redirect to home page",
		"status":  "true",
	})

}

func changePassword(c *fiber.Ctx) error {
	var info PasswordChange

	err := c.BodyParser(&info)

	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Error parsing JSON. Make sure you are sending proper JSON.",
			"status":  "false",
		})
	}
	if len(info.CurrentPassword) < 6 || len(info.NewPassword) < 6 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Provide currentPassword and newPassword in correct format",
			"status":  "false",
		})
	}

	if info.CurrentPassword == info.NewPassword {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Current Password and New Password cannot be same",
			"status":  "false",
		})
	}

	userID := c.Locals("userID")

	if userID == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "You are not not authorized",
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
			"message": "Something went wrong",
			"status":  "false",
		})
	}

	var user User
	result.Decode(&user)

	if user.Password != info.CurrentPassword {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Incorrect Password",
			"status":  "false",
		})
	}

	userCollection.UpdateOne(context.Background(), fiber.Map{
		"userID": user.UserID}, fiber.Map{
		"$set": fiber.Map{
			"password": info.NewPassword,
		},
	},
	)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Password changed",
		"status":  "true",
	})
}

func verifyEmail(c *fiber.Ctx) error {
	email := strings.ToLower(c.Query("email"))
	code := c.Query("code")

	if email == "" || code == "" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"status":  "false",
			"message": "You need to provide email and code as queries.",
		})
	}
	var db = client.Database("enotesdb")
	var userCollection = db.Collection("users")
	var codeCollection = db.Collection("codes")

	result := codeCollection.FindOne(context.Background(), fiber.Map{
		"code":  code,
		"email": email,
	})

	if result.Err() != nil {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"status":  "false",
			"message": "Request is invalid.",
		})
	}

	userCollection.UpdateOne(context.Background(), fiber.Map{
		"email": email,
	}, fiber.Map{
		"$set": fiber.Map{
			"verified": true,
		},
	})

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Email Verified. You can now login.",
		"status":  "true",
	})
}

func handleSendEmail(c *fiber.Ctx) error {
	type Body struct {
		Email string
	}
	var body Body

	err := c.BodyParser(&body)

	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "something went wrong: " + err.Error(),
			"status":  "false",
		})
	}

	var db = client.Database("enotesdb")
	var codeCollection = db.Collection("codes")

	code := fmt.Sprint(rand.Uint64())

	codeCollection.InsertOne(context.Background(), fiber.Map{
		"email":     strings.ToLower(body.Email),
		"code":      code,
		"createdAt": time.Now(),
	})

	sendEmail(strings.ToLower(body.Email), code)

	return c.Status(200).JSON(fiber.Map{
		"status":  true,
		"message": "Verification Email Sent. Check your spam folder too.",
	})

}

func logout(c *fiber.Ctx) error {
	// Set cookie header from authorization header
	auth_header := c.Get("Authorization")
	c.Request().Header.Set("Cookie", auth_header)

	sess, _ := store.Get(c)
	sess.Destroy()

	return c.Status(200).JSON(fiber.Map{
		"status":  true,
		"message": "Logged out",
	})
}
