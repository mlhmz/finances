package main

import (
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/html/v2"
	"github.com/mlhmz/finances/internal/config"
	"github.com/mlhmz/finances/internal/db"
	"github.com/mlhmz/finances/internal/handlers"
	"github.com/mlhmz/finances/internal/middleware"
)

func main() {
	cfg := config.Default()
	if _, err := db.Connect(cfg.DBPath); err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	handlers.SetupAuth(handlers.AuthConfig{
		JWTSecret:  cfg.JWTSecret,
		AccessTTL:  cfg.JWTAccessTTL,
		RefreshTTL: cfg.JWTRefreshTTL,
	})

	engine := html.New("./views", ".html")
	app := fiber.New(fiber.Config{Views: engine})

	// Public auth routes
	app.Get("/login", handlers.LoginPage)
	app.Post("/auth/request", handlers.RequestOTP)
	app.Post("/auth/verify", handlers.VerifyOTP)
	app.Get("/register", handlers.RegisterPage)
	app.Post("/register", handlers.RegisterSubmit)
	app.Get("/auth/refresh", handlers.Refresh)

	// Protected routes — auth middleware applied per-handler
	authMw := middleware.AuthMiddleware(cfg.JWTSecret, cfg.JWTAccessTTL)
	app.Get("/", authMw, handlers.Index)
	app.Get("/greet", authMw, handlers.Greet)
	app.Post("/auth/logout", authMw, handlers.Logout)

	// Test-only backdoor — exposes last OTP; only active when TEST_MODE=1
	if os.Getenv("TEST_MODE") == "1" {
		app.Get("/test/last-otp", handlers.TestLastOTP)
	}

	log.Fatal(app.Listen(cfg.Port))
}
