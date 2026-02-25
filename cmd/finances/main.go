package main

import (
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/html/v2"
	"github.com/mlhmz/finances/internal/config"
	"github.com/mlhmz/finances/internal/db"
	"github.com/mlhmz/finances/internal/handlers"
	"github.com/mlhmz/finances/internal/middleware"
	"github.com/mlhmz/finances/internal/money"
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

	// fmtAmountDisplay returns "+ €42.50" (income) or "− €42.50" (expense).
	engine.AddFunc("fmtAmountDisplay", func(m money.Money) string {
		formatted := m.Abs().Format()
		if m.IsNegative() {
			return "− " + formatted
		}
		return "+ " + formatted
	})

	// isIncome returns true when the amount is non-negative.
	engine.AddFunc("isIncome", func(m money.Money) bool {
		return !m.IsNegative()
	})

	// fmtDate formats a time.Time as "02 Jan 2006".
	engine.AddFunc("fmtDate", func(t time.Time) string {
		return t.Format("02 Jan 2006")
	})

	// fmtDateTimeInput formats a time.Time as "2006-01-02T15:04" for datetime-local inputs.
	engine.AddFunc("fmtDateTimeInput", func(t time.Time) string {
		return t.Format("2006-01-02T15:04")
	})

	// absAmountStr returns the absolute decimal string of a Money value (e.g. "42.50").
	// Used to pre-fill the amount input on edit forms (strips the sign).
	engine.AddFunc("absAmountStr", func(m money.Money) string {
		return m.Abs().DecimalString()
	})

	// nowDateTimeInput returns the current local time formatted for a datetime-local input.
	engine.AddFunc("nowDateTimeInput", func() string {
		return time.Now().Format("2006-01-02T15:04")
	})

	// add and sub helpers for pagination arithmetic in templates.
	engine.AddFunc("add", func(a, b int) int { return a + b })
	engine.AddFunc("sub", func(a, b int) int { return a - b })

	app := fiber.New(fiber.Config{Views: engine})

	// Public auth routes
	app.Get("/login", handlers.LoginPage)
	app.Post("/auth/request", handlers.RequestOTP)
	app.Post("/auth/verify", handlers.VerifyOTP)
	app.Get("/register", handlers.RegisterPage)
	app.Post("/register", handlers.RegisterSubmit)
	app.Get("/auth/refresh", handlers.Refresh)

	// Protected routes — all behind a single auth middleware group
	protected := app.Group("", middleware.AuthMiddleware(cfg.JWTSecret, cfg.JWTAccessTTL))
	protected.Get("/", handlers.Index)
	protected.Get("/greet", handlers.Greet)
	protected.Get("/profile", handlers.ProfilePage)
	protected.Post("/profile", handlers.ProfileUpdate)
	protected.Post("/auth/logout", handlers.Logout)

	// Transaction routes
	protected.Get("/transactions", handlers.TransactionsPage)
	protected.Post("/transactions", handlers.CreateTransaction)
	protected.Get("/transactions/:id/edit", handlers.EditTransactionForm)
	protected.Put("/transactions/:id", handlers.UpdateTransaction)
	protected.Get("/transactions/:id/row", handlers.TransactionRow)
	protected.Get("/transactions/:id/confirm-delete", handlers.ConfirmDeleteTransaction)
	protected.Delete("/transactions/:id", handlers.DeleteTransaction)

	log.Fatal(app.Listen(cfg.Port))
}
