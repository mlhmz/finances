package repository

import (
	"testing"
	"time"

	"github.com/mlhmz/finances/internal/currency"
	"github.com/mlhmz/finances/internal/db"
	"github.com/mlhmz/finances/internal/models"
	"github.com/mlhmz/finances/internal/money"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTxTestDB(t *testing.T) {
	t.Helper()
	testDB, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	if err := testDB.AutoMigrate(&models.User{}, &models.Transaction{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	db.DB = testDB
}

func eur(t *testing.T) currency.Currency {
	t.Helper()
	c, ok := currency.Get("EUR")
	if !ok {
		t.Fatal("EUR currency not found")
	}
	return c
}

func newTx(userID, title string, amountMinor int64, c currency.Currency) models.Transaction {
	return models.Transaction{
		UserID: userID,
		Title:  title,
		Amount: money.New(amountMinor, c),
		Date:   time.Now().Truncate(time.Second),
	}
}

func TestTransactionRepository_Create(t *testing.T) {
	setupTxTestDB(t)
	c := eur(t)

	repo := NewTransactionRepository("u1")
	tx := newTx("u1", "Lunch", 1250, c)
	if err := repo.Create(&tx); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if tx.ID == "" {
		t.Error("expected non-empty ID after create")
	}
	if tx.UserID != "u1" {
		t.Errorf("expected UserID u1, got %s", tx.UserID)
	}
}

func TestTransactionRepository_List(t *testing.T) {
	setupTxTestDB(t)
	c := eur(t)

	repo := NewTransactionRepository("u1")
	t1 := newTx("u1", "A", 100, c)
	t2 := newTx("u1", "B", 200, c)
	repo.Create(&t1) //nolint
	repo.Create(&t2) //nolint

	txs, total, err := repo.List(1, 20)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if total != 2 {
		t.Errorf("expected total 2, got %d", total)
	}
	if len(txs) != 2 {
		t.Errorf("expected 2 transactions, got %d", len(txs))
	}
}

func TestTransactionRepository_List_IsolatesUsers(t *testing.T) {
	setupTxTestDB(t)
	c := eur(t)

	repoA := NewTransactionRepository("u1")
	repoB := NewTransactionRepository("u2")

	txA := newTx("u1", "A's tx", 100, c)
	txB := newTx("u2", "B's tx", 200, c)
	repoA.Create(&txA) //nolint
	repoB.Create(&txB) //nolint

	txs, total, err := repoA.List(1, 20)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if total != 1 {
		t.Errorf("u1 should see 1 transaction, got %d", total)
	}
	if len(txs) > 0 && txs[0].Title != "A's tx" {
		t.Errorf("expected 'A's tx', got %s", txs[0].Title)
	}
}

func TestTransactionRepository_List_Pagination(t *testing.T) {
	setupTxTestDB(t)
	c := eur(t)
	repo := NewTransactionRepository("u1")

	for i := range 5 {
		tx := newTx("u1", "tx", int64((i+1)*100), c)
		tx.Date = time.Now().Add(time.Duration(i) * time.Minute)
		repo.Create(&tx) //nolint
	}

	txs, total, err := repo.List(1, 3)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if total != 5 {
		t.Errorf("expected total 5, got %d", total)
	}
	if len(txs) != 3 {
		t.Errorf("expected 3 transactions on page 1, got %d", len(txs))
	}

	txs2, _, err := repo.List(2, 3)
	if err != nil {
		t.Fatalf("List page 2: %v", err)
	}
	if len(txs2) != 2 {
		t.Errorf("expected 2 transactions on page 2, got %d", len(txs2))
	}
}

func TestTransactionRepository_GetByID(t *testing.T) {
	setupTxTestDB(t)
	c := eur(t)

	repo := NewTransactionRepository("u1")
	tx := newTx("u1", "Coffee", 350, c)
	repo.Create(&tx) //nolint

	got, err := repo.GetByID(tx.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got.Title != "Coffee" {
		t.Errorf("expected Title 'Coffee', got %s", got.Title)
	}
	if got.Amount.Amount != 350 {
		t.Errorf("expected Amount 350, got %d", got.Amount.Amount)
	}
}

func TestTransactionRepository_GetByID_IsolatesUsers(t *testing.T) {
	setupTxTestDB(t)
	c := eur(t)

	repoA := NewTransactionRepository("u1")
	repoB := NewTransactionRepository("u2")

	txA := newTx("u1", "A's secret", 100, c)
	repoA.Create(&txA) //nolint

	// u2 cannot access u1's transaction by ID
	_, err := repoB.GetByID(txA.ID)
	if err == nil {
		t.Error("expected error when u2 fetches u1's transaction, got nil")
	}
}

func TestTransactionRepository_Update(t *testing.T) {
	setupTxTestDB(t)
	c := eur(t)

	repo := NewTransactionRepository("u1")
	tx := newTx("u1", "Original", 1000, c)
	repo.Create(&tx) //nolint

	tx.Title = "Updated"
	tx.Amount = money.New(-500, c)
	if err := repo.Update(&tx); err != nil {
		t.Fatalf("Update: %v", err)
	}

	got, _ := repo.GetByID(tx.ID)
	if got.Title != "Updated" {
		t.Errorf("expected Title 'Updated', got %s", got.Title)
	}
	if got.Amount.Amount != -500 {
		t.Errorf("expected Amount -500, got %d", got.Amount.Amount)
	}
}

func TestTransactionRepository_Delete(t *testing.T) {
	setupTxTestDB(t)
	c := eur(t)

	repo := NewTransactionRepository("u1")
	tx := newTx("u1", "To delete", 100, c)
	repo.Create(&tx) //nolint

	if err := repo.Delete(tx.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	_, err := repo.GetByID(tx.ID)
	if err == nil {
		t.Error("expected error after delete, got nil")
	}
}

func TestTransactionRepository_Delete_DoesNotAffectOtherUsers(t *testing.T) {
	setupTxTestDB(t)
	c := eur(t)

	repoA := NewTransactionRepository("u1")
	repoB := NewTransactionRepository("u2")

	txA := newTx("u1", "A's tx", 100, c)
	txB := newTx("u2", "B's tx", 200, c)
	repoA.Create(&txA) //nolint
	repoB.Create(&txB) //nolint

	// u2 tries to delete u1's transaction — should be a no-op
	repoB.Delete(txA.ID) //nolint

	got, err := repoA.GetByID(txA.ID)
	if err != nil {
		t.Fatalf("u1's transaction should still exist: %v", err)
	}
	if got.Title != "A's tx" {
		t.Errorf("expected 'A's tx', got %s", got.Title)
	}
}

func TestTransactionRepository_NegativeAmount_Expense(t *testing.T) {
	setupTxTestDB(t)
	c := eur(t)

	repo := NewTransactionRepository("u1")
	tx := newTx("u1", "Groceries", -4250, c) // expense: -42.50 EUR
	repo.Create(&tx)                          //nolint

	got, err := repo.GetByID(tx.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got.Amount.Amount != -4250 {
		t.Errorf("expected Amount -4250, got %d", got.Amount.Amount)
	}
	if got.Amount.Currency.Code != "EUR" {
		t.Errorf("expected Currency EUR, got %s", got.Amount.Currency.Code)
	}
}
