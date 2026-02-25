// Package repository provides data-access objects scoped to a single user.
// Every method implicitly filters by the userID set at construction time,
// enforcing the multi-tenancy invariant: no handler can read or write
// another user's data.
package repository

import (
	"github.com/mlhmz/finances/internal/db"
	"github.com/mlhmz/finances/internal/models"
	"gorm.io/gorm"
)

// UserRepository scopes all DB operations to a single authenticated user.
type UserRepository struct {
	db     *gorm.DB
	userID string
}

// NewUserRepository returns a UserRepository bound to the given userID.
// userID must come from the auth middleware context — never from request input.
func NewUserRepository(userID string) *UserRepository {
	return &UserRepository{db: db.DB, userID: userID}
}

// Get returns the authenticated user's record.
func (r *UserRepository) Get() (*models.User, error) {
	var user models.User
	err := r.db.First(&user, "id = ?", r.userID).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// Update writes a new full name and currency to the authenticated user's record.
// Initials are re-derived from the new full name automatically.
func (r *UserRepository) Update(fullName, currency string) error {
	return r.db.Model(&models.User{}).
		Where("id = ?", r.userID).
		Updates(map[string]interface{}{
			"full_name": fullName,
			"currency":  currency,
			"initials":  models.DeriveInitials(fullName),
		}).Error
}
