package models

import (
	"strings"
	"time"
)

// User represents an authenticated user of the application.
type User struct {
	ID        string `gorm:"primaryKey"`
	Email     string `gorm:"uniqueIndex;not null"`
	FullName  string `gorm:"not null"`
	Currency  string `gorm:"not null;default:'EUR'"`
	Initials  string `gorm:"not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

// DeriveInitials computes uppercase initials from a full name.
// Single word → first character. Two or more words → first char of first + first char of last.
func DeriveInitials(fullName string) string {
	parts := strings.Fields(fullName)
	switch len(parts) {
	case 0:
		return ""
	case 1:
		return strings.ToUpper(string([]rune(parts[0])[:1]))
	default:
		first := []rune(parts[0])[0]
		last := []rune(parts[len(parts)-1])[0]
		return strings.ToUpper(string([]rune{first, last}))
	}
}
