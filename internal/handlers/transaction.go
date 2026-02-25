package handlers

import (
	"maps"
	"math"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/mlhmz/finances/internal/currency"
	"github.com/mlhmz/finances/internal/middleware"
	"github.com/mlhmz/finances/internal/models"
	"github.com/mlhmz/finances/internal/money"
	"github.com/mlhmz/finances/internal/repository"
)

var txValidPageSizes = map[int]bool{5: true, 10: true, 15: true, 20: true, 25: true}

// TransactionsPage handles GET /transactions.
func TransactionsPage(c *fiber.Ctx) error {
	userID := middleware.CurrentUserID(c)
	page := c.QueryInt("page", 1)
	if page < 1 {
		page = 1
	}
	pageSize := c.QueryInt("pageSize", 10)
	if !txValidPageSizes[pageSize] {
		pageSize = 10
	}

	txRepo := repository.NewTransactionRepository(userID)
	transactions, total, err := txRepo.List(page, pageSize)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Failed to load transactions.")
	}

	totalPages := int(math.Ceil(float64(total) / float64(pageSize)))
	if totalPages < 1 {
		totalPages = 1
	}

	userRepo := repository.NewUserRepository(userID)
	user, err := userRepo.Get()
	if err != nil {
		return c.Redirect("/login", fiber.StatusFound)
	}

	currencies := currency.Supported()
	return c.Render("transactions", fiber.Map{
		"Title":        "Transactions",
		"ActivePage":   "transactions",
		"Transactions": transactions,
		"Page":         page,
		"PageSize":     pageSize,
		"TotalPages":   totalPages,
		"UserCurrency": user.Currency,
		"User":         user,
		"Currencies":   currencies,
		"FormData": fiber.Map{
			"IsEdit":       false,
			"Transaction":  nil,
			"Currencies":   currencies,
			"UserCurrency": user.Currency,
			"Errors":       nil,
		},
	}, "layouts/app")
}

// CreateTransaction handles POST /transactions.
func CreateTransaction(c *fiber.Ctx) error {
	userID := middleware.CurrentUserID(c)

	t, errors, err := parseTransactionForm(c, userID, nil)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Server error.")
	}

	if len(errors) > 0 {
		return renderTransactionForm(c, userID, false, nil, errors, formValuesFromCtx(c))
	}

	repo := repository.NewTransactionRepository(userID)
	if err := repo.Create(t); err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Failed to save transaction.")
	}

	c.Set("HX-Trigger", "closeTransactionForm")
	c.Set("HX-Retarget", "#transaction-list")
	c.Set("HX-Reswap", "afterbegin")
	return c.Render("partials/transaction_row", t)
}

// EditTransactionForm handles GET /transactions/:id/edit.
func EditTransactionForm(c *fiber.Ctx) error {
	userID := middleware.CurrentUserID(c)
	id := c.Params("id")

	repo := repository.NewTransactionRepository(userID)
	t, err := repo.GetByID(id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).SendString("Transaction not found.")
	}

	return renderTransactionForm(c, userID, true, t, nil, nil)
}

// UpdateTransaction handles PUT /transactions/:id.
func UpdateTransaction(c *fiber.Ctx) error {
	userID := middleware.CurrentUserID(c)
	id := c.Params("id")

	repo := repository.NewTransactionRepository(userID)
	existing, err := repo.GetByID(id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).SendString("Transaction not found.")
	}

	updated, errors, err := parseTransactionForm(c, userID, existing)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Server error.")
	}

	if len(errors) > 0 {
		return renderTransactionForm(c, userID, true, existing, errors, formValuesFromCtx(c))
	}

	if err := repo.Update(updated); err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Failed to update transaction.")
	}

	// Re-fetch to get persisted state.
	updated, err = repo.GetByID(id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Failed to reload transaction.")
	}

	c.Set("HX-Trigger", "closeTransactionForm")
	c.Set("HX-Retarget", "#tx-"+id)
	c.Set("HX-Reswap", "outerHTML")
	return c.Render("partials/transaction_row", updated)
}

// TransactionRow handles GET /transactions/:id/row — returns the row partial.
// Used by the cancel-delete action to restore the original row.
func TransactionRow(c *fiber.Ctx) error {
	userID := middleware.CurrentUserID(c)
	id := c.Params("id")

	repo := repository.NewTransactionRepository(userID)
	t, err := repo.GetByID(id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).SendString("Transaction not found.")
	}

	return c.Render("partials/transaction_row", t)
}

// ConfirmDeleteTransaction handles GET /transactions/:id/confirm-delete.
func ConfirmDeleteTransaction(c *fiber.Ctx) error {
	userID := middleware.CurrentUserID(c)
	id := c.Params("id")

	repo := repository.NewTransactionRepository(userID)
	t, err := repo.GetByID(id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).SendString("Transaction not found.")
	}

	return c.Render("partials/transaction_confirm_delete", t)
}

// DeleteTransaction handles DELETE /transactions/:id.
func DeleteTransaction(c *fiber.Ctx) error {
	userID := middleware.CurrentUserID(c)
	id := c.Params("id")

	repo := repository.NewTransactionRepository(userID)
	if err := repo.Delete(id); err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Failed to delete transaction.")
	}

	return c.SendString("")
}

// ── Helpers ──────────────────────────────────────────────────────────────────

// parseTransactionForm validates form fields and returns a ready-to-persist
// Transaction, a field-error map, and any unexpected error.
// existing is non-nil for updates (preserves ID, CreatedAt).
func parseTransactionForm(c *fiber.Ctx, userID string, existing *models.Transaction) (*models.Transaction, map[string]string, error) {
	txType := strings.TrimSpace(c.FormValue("type"))
	amountStr := strings.TrimSpace(c.FormValue("amount"))
	currencyCode := strings.TrimSpace(c.FormValue("currency"))
	title := strings.TrimSpace(c.FormValue("title"))
	dateStr := strings.TrimSpace(c.FormValue("date"))
	description := strings.TrimSpace(c.FormValue("description"))

	errors := map[string]string{}

	if txType != "income" && txType != "expense" {
		errors["type"] = "Please select Income or Expense."
	}

	cur, okCur := currency.Get(currencyCode)
	if !okCur {
		errors["currency"] = "Invalid currency."
	}

	var minorUnits int64
	if amountStr == "" {
		errors["amount"] = "Amount is required."
	} else if okCur {
		var parseErr error
		minorUnits, parseErr = money.ParseDecimal(amountStr, cur.Exponent)
		if parseErr != nil || minorUnits <= 0 {
			errors["amount"] = "Amount must be a positive number."
		}
	}

	if title == "" {
		errors["title"] = "Title is required."
	}

	var date time.Time
	if dateStr == "" {
		errors["date"] = "Date is required."
	} else {
		var parseErr error
		date, parseErr = parseDateTimeInput(dateStr)
		if parseErr != nil {
			errors["date"] = "Invalid date."
		}
	}

	if len(errors) > 0 {
		return nil, errors, nil
	}

	if txType == "expense" {
		minorUnits = -minorUnits
	}

	t := &models.Transaction{
		UserID:      userID,
		Title:       title,
		Description: description,
		Amount:      money.New(minorUnits, cur),
		Date:        date,
	}
	if existing != nil {
		t.ID = existing.ID
		t.CreatedAt = existing.CreatedAt
	}

	return t, nil, nil
}

// parseDateTimeInput parses "2006-01-02T15:04" or "2006-01-02" date strings.
func parseDateTimeInput(s string) (time.Time, error) {
	if t, err := time.Parse("2006-01-02T15:04", s); err == nil {
		return t, nil
	}
	return time.Parse("2006-01-02", s)
}

// renderTransactionForm renders the form partial with the given state.
func renderTransactionForm(c *fiber.Ctx, userID string, isEdit bool, t *models.Transaction, errors map[string]string, vals fiber.Map) error {
	userRepo := repository.NewUserRepository(userID)
	user, err := userRepo.Get()
	if err != nil {
		return c.Redirect("/login", fiber.StatusFound)
	}

	data := fiber.Map{
		"IsEdit":       isEdit,
		"Transaction":  t,
		"Currencies":   currency.Supported(),
		"UserCurrency": user.Currency,
		"Errors":       errors,
	}
	maps.Copy(data, vals)
	return c.Render("partials/transaction_form", data)
}

// formValuesFromCtx extracts raw form values to re-populate the form on errors.
func formValuesFromCtx(c *fiber.Ctx) fiber.Map {
	return fiber.Map{
		"FormType":        c.FormValue("type"),
		"FormAmount":      c.FormValue("amount"),
		"FormCurrency":    c.FormValue("currency"),
		"FormTitle":       c.FormValue("title"),
		"FormDate":        c.FormValue("date"),
		"FormDescription": c.FormValue("description"),
	}
}
