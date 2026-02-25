package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/mlhmz/finances/internal/middleware"
	"github.com/mlhmz/finances/internal/repository"
)

func Index(c *fiber.Ctx) error {
	repo := repository.NewUserRepository(middleware.CurrentUserID(c))
	user, err := repo.Get()
	if err != nil {
		return c.Redirect("/login", fiber.StatusFound)
	}
	return c.Render("index", fiber.Map{
		"Title":      "Home",
		"ActivePage": "home",
		"User":       user,
	}, "layouts/app")
}

func Greet(c *fiber.Ctx) error {
	return c.SendString("Hello, World! Welcome to your Finances app.")
}
