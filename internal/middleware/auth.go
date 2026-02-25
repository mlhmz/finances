// Package middleware provides Fiber middleware for the Finances application.
package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/mlhmz/finances/internal/auth"
	"github.com/mlhmz/finances/internal/db"
	"github.com/mlhmz/finances/internal/models"
)

// AuthMiddleware returns a Fiber handler that enforces JWT authentication.
//
// Flow:
//  1. Read access_token cookie.
//  2. If valid → store claims in context locals, call next.
//  3. If missing/expired → try refresh_token cookie.
//     a. If valid refresh token → issue new access_token, store user in locals, call next.
//     b. Otherwise → clear both cookies, redirect to /login.
func AuthMiddleware(jwtSecret string, accessTTL int) fiber.Handler {
	return func(c *fiber.Ctx) error {
		accessStr := c.Cookies("access_token")
		if accessStr != "" {
			claims, err := auth.ParseAccessToken(accessStr, jwtSecret)
			if err == nil {
				c.Locals("userID", claims.Subject)
				c.Locals("email", claims.Email)
				return c.Next()
			}
		}

		// access token missing or expired — try refresh token
		refreshStr := c.Cookies("refresh_token")
		if refreshStr != "" {
			refreshClaims, err := auth.ParseRefreshToken(refreshStr, jwtSecret)
			if err == nil {
				// Look up user to get email
				var user models.User
				if res := db.DB.First(&user, "id = ?", refreshClaims.Subject); res.Error == nil {
					newAccess, err := auth.IssueAccessToken(user.ID, user.Email, jwtSecret, accessTTL)
					if err == nil {
						c.Cookie(&fiber.Cookie{
							Name:     "access_token",
							Value:    newAccess,
							HTTPOnly: true,
							SameSite: "Lax",
							MaxAge:   accessTTL,
						})
						c.Locals("userID", user.ID)
						c.Locals("email", user.Email)
						return c.Next()
					}
				}
			}
		}

		// clear stale cookies and redirect
		c.Cookie(&fiber.Cookie{Name: "access_token", Value: "", MaxAge: -1, HTTPOnly: true, SameSite: "Lax"})
		c.Cookie(&fiber.Cookie{Name: "refresh_token", Value: "", MaxAge: -1, HTTPOnly: true, SameSite: "Lax"})
		return c.Redirect("/login", fiber.StatusFound)
	}
}
