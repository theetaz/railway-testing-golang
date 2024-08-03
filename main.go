package main

import (
	"os"

	"github.com/gofiber/fiber/v2"
)

func main() {
	// create fiber instance
	app := fiber.New()

	app.Get("/hello", func(c *fiber.Ctx) error {
		return c.SendString("World")
	})

	// start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}
	app.Listen(":" + port)
}