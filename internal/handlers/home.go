package handlers

import "github.com/gofiber/fiber/v2"

func Index(c *fiber.Ctx) error {
	return c.Render("index", fiber.Map{"Title": "Finances"})
}

func Greet(c *fiber.Ctx) error {
	return c.SendString("Hello, World! Welcome to your Finances app.")
}
