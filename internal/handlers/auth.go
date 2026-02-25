package handlers

import (
	"fmt"
	"net/url"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/mlhmz/finances/internal/auth"
	"github.com/mlhmz/finances/internal/db"
	"github.com/mlhmz/finances/internal/models"
	"gorm.io/gorm"
)

// AuthConfig holds the JWT configuration consumed by the auth handlers.
// Populate it by calling SetupAuth before registering routes.
type AuthConfig struct {
	JWTSecret  string
	AccessTTL  int
	RefreshTTL int
}

var authCfg AuthConfig

// SetupAuth stores the auth configuration for use in auth handlers.
func SetupAuth(cfg AuthConfig) {
	authCfg = cfg
}

// ─── cookie helpers ──────────────────────────────────────────────────────────

func setAuthCookies(c *fiber.Ctx, userID, email string) error {
	accessToken, err := auth.IssueAccessToken(userID, email, authCfg.JWTSecret, authCfg.AccessTTL)
	if err != nil {
		return err
	}
	refreshToken, err := auth.IssueRefreshToken(userID, authCfg.JWTSecret, authCfg.RefreshTTL)
	if err != nil {
		return err
	}
	c.Cookie(&fiber.Cookie{
		Name:     "access_token",
		Value:    accessToken,
		HTTPOnly: true,
		SameSite: "Lax",
		MaxAge:   authCfg.AccessTTL,
	})
	c.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		HTTPOnly: true,
		SameSite: "Lax",
		MaxAge:   authCfg.RefreshTTL,
	})
	return nil
}

func clearAuthCookies(c *fiber.Ctx) {
	c.Cookie(&fiber.Cookie{Name: "access_token", Value: "", MaxAge: -1, HTTPOnly: true, SameSite: "Lax"})
	c.Cookie(&fiber.Cookie{Name: "refresh_token", Value: "", MaxAge: -1, HTTPOnly: true, SameSite: "Lax"})
}

// ─── OTP helper ──────────────────────────────────────────────────────────────

// lastOTPByEmail stores the most recently issued plaintext OTP per email address.
// Only populated and exposed when TEST_MODE=1; never used in production.
var (
	lastOTPByEmail = map[string]string{}
	lastOTPMu      sync.Mutex
)

func issueOTP(userID, email string) error {
	code, hash, err := auth.GenerateOTP()
	if err != nil {
		return err
	}
	// delete any existing OTPs for this user
	db.DB.Where("user_id = ?", userID).Delete(&models.OTPToken{})
	token := models.OTPToken{
		ID:        uuid.New().String(),
		UserID:    userID,
		CodeHash:  hash,
		ExpiresAt: time.Now().Add(15 * time.Minute),
	}
	if res := db.DB.Create(&token); res.Error != nil {
		return res.Error
	}
	lastOTPMu.Lock()
	lastOTPByEmail[email] = code
	lastOTPMu.Unlock()
	fmt.Printf("[AUTH] OTP for %s: %s (expires in 15 minutes)\n", email, code)
	return nil
}

// TestLastOTP returns the last issued OTP for the given ?email= param as plain text.
// Only available when TEST_MODE=1 — must never be registered in production.
func TestLastOTP(c *fiber.Ctx) error {
	email := c.Query("email")
	if email == "" {
		return c.Status(fiber.StatusBadRequest).SendString("email query param required")
	}
	lastOTPMu.Lock()
	code := lastOTPByEmail[email]
	lastOTPMu.Unlock()
	return c.SendString(code)
}

// ─── Handlers ────────────────────────────────────────────────────────────────

// LoginPage renders GET /login.
func LoginPage(c *fiber.Ctx) error {
	return c.Render("login", fiber.Map{})
}

// RequestOTP handles POST /auth/request.
//
// Known email   → generate OTP, return otp_form fragment.
// Unknown email → 302 redirect to /register?email=<email>.
func RequestOTP(c *fiber.Ctx) error {
	email := c.FormValue("email")
	if email == "" {
		return renderOTPError(c, "Email is required.")
	}

	var user models.User
	result := db.DB.Where("email = ?", email).First(&user)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			c.Set("HX-Redirect", "/register?email="+url.QueryEscape(email))
			return c.SendStatus(fiber.StatusOK)
		}
		return renderOTPError(c, "An error occurred. Please try again.")
	}

	if err := issueOTP(user.ID, email); err != nil {
		return renderOTPError(c, "Failed to generate code. Please try again.")
	}
	return c.Render("partials/otp_form", fiber.Map{"Email": email})
}

// VerifyOTP handles POST /auth/verify.
func VerifyOTP(c *fiber.Ctx) error {
	email := c.FormValue("email")
	code := c.FormValue("code")

	var user models.User
	if res := db.DB.Where("email = ?", email).First(&user); res.Error != nil {
		return renderOTPError(c, "User not found.")
	}

	var otp models.OTPToken
	if res := db.DB.Where("user_id = ?", user.ID).First(&otp); res.Error != nil {
		return renderOTPError(c, "No active code. Please request a new one.")
	}

	if time.Now().After(otp.ExpiresAt) {
		db.DB.Delete(&otp)
		return renderOTPError(c, "Code expired. Request a new one.")
	}

	if otp.AttemptCount >= 3 {
		db.DB.Delete(&otp)
		return renderOTPError(c, "Too many attempts. Request a new code.")
	}

	if auth.HashOTP(code) != otp.CodeHash {
		otp.AttemptCount++
		db.DB.Save(&otp)
		if otp.AttemptCount >= 3 {
			db.DB.Delete(&otp)
			return renderOTPError(c, "Incorrect code. No attempts remaining. Request a new code.")
		}
		remaining := 3 - otp.AttemptCount
		return renderOTPError(c, fmt.Sprintf("Incorrect code. %d attempt(s) remaining.", remaining))
	}

	// Success — delete OTP and issue cookies
	db.DB.Delete(&otp)
	if err := setAuthCookies(c, user.ID, user.Email); err != nil {
		return renderOTPError(c, "Failed to create session. Please try again.")
	}
	c.Set("HX-Redirect", "/")
	return c.SendStatus(fiber.StatusOK)
}

// RegisterPage renders GET /register.
func RegisterPage(c *fiber.Ctx) error {
	email := c.Query("email")
	if email == "" {
		return c.Redirect("/login", fiber.StatusFound)
	}
	return c.Render("register", fiber.Map{
		"Email":      email,
		"Currencies": models.AllowedCurrencies,
	})
}

// RegisterSubmit handles POST /register.
func RegisterSubmit(c *fiber.Ctx) error {
	email := c.FormValue("email")
	fullName := c.FormValue("full_name")
	currency := c.FormValue("currency")

	// Validate email not already taken (race-condition guard)
	var existing models.User
	if res := db.DB.Where("email = ?", email).First(&existing); res.Error == nil {
		return c.Render("register", fiber.Map{
			"Email":      email,
			"Currencies": models.AllowedCurrencies,
			"Error":      "Email is already registered.",
		})
	}

	if fullName == "" {
		return c.Render("register", fiber.Map{
			"Email":      email,
			"Currencies": models.AllowedCurrencies,
			"Error":      "Full name is required.",
		})
	}

	validCurrency := false
	for _, c := range models.AllowedCurrencies {
		if c == currency {
			validCurrency = true
			break
		}
	}
	if !validCurrency {
		return c.Render("register", fiber.Map{
			"Email":      email,
			"Currencies": models.AllowedCurrencies,
			"Error":      "Invalid currency.",
		})
	}

	user := models.User{
		ID:       uuid.New().String(),
		Email:    email,
		FullName: fullName,
		Currency: currency,
		Initials: models.DeriveInitials(fullName),
	}
	if res := db.DB.Create(&user); res.Error != nil {
		return c.Render("register", fiber.Map{
			"Email":      email,
			"Currencies": models.AllowedCurrencies,
			"Error":      "Failed to create account. Please try again.",
		})
	}

	if err := issueOTP(user.ID, email); err != nil {
		return c.Render("register", fiber.Map{
			"Email":      email,
			"Currencies": models.AllowedCurrencies,
			"Error":      "Account created but failed to send code. Please log in.",
		})
	}

	return c.Render("register", fiber.Map{
		"Email":      email,
		"Currencies": models.AllowedCurrencies,
		"ShowOTP":    true,
	})
}

// Logout handles POST /auth/logout.
func Logout(c *fiber.Ctx) error {
	clearAuthCookies(c)
	return c.Redirect("/login", fiber.StatusFound)
}

// Refresh handles GET /auth/refresh — issues a new access token from a valid refresh token.
func Refresh(c *fiber.Ctx) error {
	refreshStr := c.Cookies("refresh_token")
	if refreshStr == "" {
		clearAuthCookies(c)
		return c.Redirect("/login", fiber.StatusFound)
	}
	claims, err := auth.ParseRefreshToken(refreshStr, authCfg.JWTSecret)
	if err != nil {
		clearAuthCookies(c)
		return c.Redirect("/login", fiber.StatusFound)
	}

	var user models.User
	if res := db.DB.First(&user, "id = ?", claims.Subject); res.Error != nil {
		clearAuthCookies(c)
		return c.Redirect("/login", fiber.StatusFound)
	}

	accessToken, err := auth.IssueAccessToken(user.ID, user.Email, authCfg.JWTSecret, authCfg.AccessTTL)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Failed to issue token")
	}
	c.Cookie(&fiber.Cookie{
		Name:     "access_token",
		Value:    accessToken,
		HTTPOnly: true,
		SameSite: "Lax",
		MaxAge:   authCfg.AccessTTL,
	})
	return c.Redirect("/", fiber.StatusFound)
}

// ─── fragment helper ─────────────────────────────────────────────────────────

func renderOTPError(c *fiber.Ctx, message string) error {
	return c.Render("partials/otp_error", fiber.Map{"Message": message})
}
