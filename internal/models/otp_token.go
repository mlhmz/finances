package models

import "time"

// OTPToken stores a hashed one-time passcode for passwordless authentication.
// At most one record per user should exist at any time; old records are deleted
// before a new one is created.
type OTPToken struct {
	ID           string    `gorm:"primaryKey"`
	UserID       string    `gorm:"not null;index"`
	CodeHash     string    `gorm:"not null"` // SHA-256 hex of plaintext code
	ExpiresAt    time.Time `gorm:"not null"`
	AttemptCount int       `gorm:"not null;default:0"`
	CreatedAt    time.Time
}
