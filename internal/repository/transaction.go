package repository

import (
	"github.com/google/uuid"
	"github.com/mlhmz/finances/internal/db"
	"github.com/mlhmz/finances/internal/models"
	"gorm.io/gorm"
)

// TransactionRepository scopes all DB operations to a single authenticated user.
type TransactionRepository struct {
	db     *gorm.DB
	userID string
}

// NewTransactionRepository returns a TransactionRepository bound to the given userID.
// userID must come from the auth middleware context — never from request input.
func NewTransactionRepository(userID string) *TransactionRepository {
	return &TransactionRepository{db: db.DB, userID: userID}
}

// List returns a page of transactions for the user, sorted newest first.
// Returns the slice, total record count, and any error.
func (r *TransactionRepository) List(page, pageSize int) ([]models.Transaction, int64, error) {
	var transactions []models.Transaction
	var total int64

	q := r.db.Model(&models.Transaction{}).Where("user_id = ?", r.userID)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := q.Order("date desc, created_at desc").Offset(offset).Limit(pageSize).Find(&transactions).Error
	return transactions, total, err
}

// Create inserts a new transaction. Sets ID (UUID) and UserID before insert.
func (r *TransactionRepository) Create(t *models.Transaction) error {
	t.ID = uuid.New().String()
	t.UserID = r.userID
	return r.db.Create(t).Error
}

// GetByID returns the transaction with the given ID scoped to the user.
func (r *TransactionRepository) GetByID(id string) (*models.Transaction, error) {
	var t models.Transaction
	err := r.db.First(&t, "id = ? AND user_id = ?", id, r.userID).Error
	if err != nil {
		return nil, err
	}
	return &t, nil
}

// Update saves changes to an existing transaction (scoped to user).
func (r *TransactionRepository) Update(t *models.Transaction) error {
	return r.db.Model(&models.Transaction{}).
		Where("id = ? AND user_id = ?", t.ID, r.userID).
		Updates(map[string]interface{}{
			"title":                t.Title,
			"description":          t.Description,
			"amount_amount":        t.Amount.Amount,
			"amount_currency_code": t.Amount.Currency.Code,
			"date":                 t.Date,
		}).Error
}

// Delete removes the transaction with the given ID (scoped to user).
func (r *TransactionRepository) Delete(id string) error {
	return r.db.Where("id = ? AND user_id = ?", id, r.userID).
		Delete(&models.Transaction{}).Error
}
