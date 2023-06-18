package main

import (
	"context"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Note struct {
	ID           string `json:"id"`
	Title        string `json:"title" validate:"required"`
	Body         string `json:"body" validate:"required"`
	LastModified int    `json:"lastModified" bson:"lastModified" validate:"required"`
	ItemKey      string `json:"itemKey" bson:"itemKey" validate:"required"`
}
type NoteForUpdate struct {
	ID           string `json:"id"`
	Title        string `json:"title"`
	Body         string `json:"body"`
	LastModified int    `json:"lastModified" bson:"lastModified" validate:"required"`
	HasTitle     bool   `json:"hasTitle" validate:"required"`
	HasBody      bool   `json:"hasBody" validate:"required"`
}

type UserOnlyNote struct {
	Notes []Note `json:"notes"`
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
			"status":  false,
			"message": "Make sure the request has a title, body and lastModified",
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
			"status":  false,
		})
	}

	return c.Status(fiber.StatusAccepted).JSON(fiber.Map{
		"status":  true,
		"message": "Note Added",
		"note": fiber.Map{
			"title":        note.Title,
			"body":         note.Body,
			"id":           note.ID,
			"lastModified": note.LastModified,
		},
	})
}

func getNotes(c *fiber.Ctx) error {
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
			"status":  false,
		})
	}

	result.Decode(&user)

	notes := user.Notes

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status": true,
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
			"status":  false,
			"message": "Make sure the request has a note ID",
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
			"status":  false,
			"message": "Note Doesn't Exist",
		})
	}

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "something went wrong: " + err.Error(),
			"status":  false,
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  true,
		"message": "Note Deleted",
	})
}

func updateNote(c *fiber.Ctx) error {
	userID := c.Locals("userID")

	var db = client.Database("enotesdb")
	var userCollection = db.Collection("users")

	var note NoteForUpdate
	err := c.BodyParser(&note)

	if err != nil {
		return c.Status(fiber.StatusNotAcceptable).JSON(fiber.Map{
			"status":  false,
			"message": "Make sure the request has a title or body, lastModified, hasTitle and hasBody",
		})
	}

	if note.HasTitle && !note.HasBody {
		_, err = userCollection.UpdateOne(context.Background(), fiber.Map{
			"userID": userID, "notes.id": note.ID,
		}, fiber.Map{
			"$set": fiber.Map{
				"notes.$.title": note.Title, "notes.$.lastModified": note.LastModified,
			},
		})
	} else if !note.HasTitle && note.HasBody{
		_, err = userCollection.UpdateOne(context.Background(), fiber.Map{
			"userID": userID, "notes.id": note.ID,
		}, fiber.Map{
			"$set": fiber.Map{
				"notes.$.body": note.Body, "notes.$.lastModified": note.LastModified,
			},
		})
	} else {
		return c.Status(fiber.StatusNotAcceptable).JSON(fiber.Map{
			"status":  false,
			"message": "Make sure the request has a title or body, lastModified, hasTitle and hasBody",
		})
	}

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Something went wrong: " + err.Error(),
			"status":  false,
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  true,
		"message": "Note Updated",
	})
}

func getNote(c *fiber.Ctx) error {
	userID := c.Locals("userID")
	noteID := c.Params("id")
	var user UserOnlyNote

	if noteID == "" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"status":  false,
			"message": "Note ID parameter is missing",
		})
	}

	var db = client.Database("enotesdb")
	var userCollection = db.Collection("users")

	result := userCollection.FindOne(context.Background(), fiber.Map{
		"userID": userID, "notes.id": noteID,
	}, options.FindOne().SetProjection(fiber.Map{"notes.$": 1}))

	result.Decode(&user)

	if result.Err() != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"status":  false,
			"message": "Note Doesn't Exist",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status": true,
		"note":   user.Notes[0],
	})
}
