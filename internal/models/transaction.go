package models

import (
	"time"

	"github.com/mlhmz/finances/internal/money"
)

// Transaction represents a single income or expense event for a user.
// Amount is stored in signed minor units: positive = income, negative = expense.
type Transaction struct {
	ID          string      `gorm:"primaryKey"`
	UserID      string      `gorm:"not null;index"`
	Title       string      `gorm:"not null"`
	Description string
	Amount      money.Money `gorm:"embedded;embeddedPrefix:amount_"`
	Date        time.Time   `gorm:"not null"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
