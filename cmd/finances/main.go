package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/html/v2"
	"github.com/mlhmz/finances/internal/config"
	"github.com/mlhmz/finances/internal/db"
	"github.com/mlhmz/finances/internal/handlers"
)

func main() {
	cfg := config.Default()
	if _, err := db.Connect(cfg.DBPath); err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	engine := html.New("./views", ".html")
	app := fiber.New(fiber.Config{Views: engine})
	app.Get("/", handlers.Index)
	app.Get("/greet", handlers.Greet)
	log.Fatal(app.Listen(cfg.Port))
}
