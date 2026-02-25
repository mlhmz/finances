package handlers

import (
	"maps"

	"github.com/gofiber/fiber/v2"
	"github.com/mlhmz/finances/internal/currency"
	"github.com/mlhmz/finances/internal/middleware"
	"github.com/mlhmz/finances/internal/repository"
)

// ProfilePage handles GET /profile.
func ProfilePage(c *fiber.Ctx) error {
	repo := repository.NewUserRepository(middleware.CurrentUserID(c))
	user, err := repo.Get()
	if err != nil {
		return c.Redirect("/login", fiber.StatusFound)
	}
	return c.Render("profile", fiber.Map{
		"Title":      "Profile",
		"ActivePage": "profile",
		"User":       user,
		"Currencies": currency.Supported(),
	}, "layouts/app")
}

// ProfileUpdate handles POST /profile.
func ProfileUpdate(c *fiber.Ctx) error {
	fullName := c.FormValue("full_name")
	currencyCode := c.FormValue("currency")
	repo := repository.NewUserRepository(middleware.CurrentUserID(c))

	user, err := repo.Get()
	if err != nil {
		return c.Redirect("/login", fiber.StatusFound)
	}

	render := func(extra fiber.Map) error {
		data := fiber.Map{
			"Title":      "Profile",
			"ActivePage": "profile",
			"User":       user,
			"Currencies": currency.Supported(),
		}
		maps.Copy(data, extra)
		return c.Render("profile", data, "layouts/app")
	}

	if fullName == "" {
		return render(fiber.Map{"Error": "Full name is required."})
	}
	if _, ok := currency.Get(currencyCode); !ok {
		return render(fiber.Map{"Error": "Invalid currency."})
	}
	if err := repo.Update(fullName, currencyCode); err != nil {
		return render(fiber.Map{"Error": "Failed to update profile. Please try again."})
	}

	// Re-fetch to reflect persisted changes (e.g. derived initials).
	user, err = repo.Get()
	if err != nil {
		return c.Redirect("/login", fiber.StatusFound)
	}
	return render(fiber.Map{"Success": "Profile updated."})
}
