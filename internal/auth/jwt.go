// Package auth provides JWT issuance and parsing utilities shared between
// handlers and middleware.
package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// AccessClaims are the claims embedded in the short-lived access token.
type AccessClaims struct {
	Email string `json:"email"`
	jwt.RegisteredClaims
}

// RefreshClaims are the claims embedded in the long-lived refresh token.
type RefreshClaims struct {
	Type string `json:"type"`
	jwt.RegisteredClaims
}

// IssueAccessToken creates a signed HS256 access token that expires in ttl seconds.
func IssueAccessToken(userID, email, secret string, ttl int) (string, error) {
	claims := AccessClaims{
		Email: email,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(ttl) * time.Second)),
		},
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return t.SignedString([]byte(secret))
}

// IssueRefreshToken creates a signed HS256 refresh token that expires in ttl seconds.
func IssueRefreshToken(userID, secret string, ttl int) (string, error) {
	claims := RefreshClaims{
		Type: "refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(ttl) * time.Second)),
		},
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return t.SignedString([]byte(secret))
}

// ParseAccessToken validates and parses an access token string.
func ParseAccessToken(tokenStr, secret string) (*AccessClaims, error) {
	t, err := jwt.ParseWithClaims(tokenStr, &AccessClaims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := t.Claims.(*AccessClaims)
	if !ok || !t.Valid {
		return nil, fmt.Errorf("invalid access token")
	}
	return claims, nil
}

// ParseRefreshToken validates and parses a refresh token string.
func ParseRefreshToken(tokenStr, secret string) (*RefreshClaims, error) {
	t, err := jwt.ParseWithClaims(tokenStr, &RefreshClaims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := t.Claims.(*RefreshClaims)
	if !ok || !t.Valid || claims.Type != "refresh" {
		return nil, fmt.Errorf("invalid refresh token")
	}
	return claims, nil
}

// GenerateOTP produces an 8-character uppercase alphanumeric code using
// crypto/rand and returns both the plaintext code and its SHA-256 hex digest.
func GenerateOTP() (plaintext, hash string, err error) {
	const alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	buf := make([]byte, 8)
	if _, err = rand.Read(buf); err != nil {
		return
	}
	var sb strings.Builder
	for _, b := range buf {
		sb.WriteByte(alphabet[int(b)%len(alphabet)])
	}
	plaintext = sb.String()
	hash = HashOTP(plaintext)
	return
}

// HashOTP returns the SHA-256 hex digest of the (uppercased) code.
func HashOTP(code string) string {
	sum := sha256.Sum256([]byte(strings.ToUpper(code)))
	return hex.EncodeToString(sum[:])
}
