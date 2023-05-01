package main

import (
	"context"

	"github.com/gofiber/fiber/v2"
)

type Note struct {
	ID    string `json:"id"`
	Title string `json:"title" validate:"required"`
	Body  string `json:"body" validate:"required"`
}

type NoteToDelete struct {
	ID string `json:"id"`
}

func addNote(c *fiber.Ctx) error {
	userID := c.Locals("userID")

	var db = client.Database("enotesdb")
	var userCollection = db.Collection("users")

	var note Note

	err := c.BodyParser(&note)

	if err != nil {
		return c.Status(fiber.StatusNotAcceptable).JSON(fiber.Map{
			"status":  "false",
			"message": "Make sure the request has a title and a body",
		})
	}

	note.ID, _ = GenerateRandomString(20)

	_, err = userCollection.UpdateOne(context.Background(), fiber.Map{
		"userID": userID,
	}, fiber.Map{
		"$push": fiber.Map{
			"notes": note,
		},
	})

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "something went wrong: " + err.Error(),
			"status":  "false",
		})
	}

	return c.Status(fiber.StatusAccepted).JSON(fiber.Map{
		"status":  "true",
		"message": "Note Added",
	})
}

func getNote(c *fiber.Ctx) error {
	userID := c.Locals("userID")

	var db = client.Database("enotesdb")
	var userCollection = db.Collection("users")

	var user User

	result := userCollection.FindOne(context.Background(), fiber.Map{
		"userID": userID,
	})

	if result.Err() != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "something went wrong: " + result.Err().Error(),
			"status":  "false",
		})
	}

	result.Decode(&user)

	notes := user.Notes

	return c.Status(fiber.StatusAccepted).JSON(fiber.Map{
		"status": "true",
		"notes":  notes,
	})
}

func deleteNote(c *fiber.Ctx) error {

	userID := c.Locals("userID")

	var db = client.Database("enotesdb")
	var userCollection = db.Collection("users")

	var note NoteToDelete

	err := c.BodyParser(&note)

	if err != nil {
		return c.Status(fiber.StatusNotAcceptable).JSON(fiber.Map{
			"status":  "false",
			"message": "Make sure the request has a title and a body",
		})
	}

	result, err := userCollection.UpdateOne(context.Background(), fiber.Map{
		"userID": userID,
	}, fiber.Map{
		"$pull": fiber.Map{
			"notes": fiber.Map{
				"id": note.ID,
			},
		},
	})

	if result.ModifiedCount == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"status":  "true",
			"message": "Note Doesn't Exist",
		})
	}

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "something went wrong: " + err.Error(),
			"status":  "false",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  "true",
		"message": "Note Deleted",
	})
}
