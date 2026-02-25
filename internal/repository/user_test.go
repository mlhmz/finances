package repository

import (
	"testing"

	"github.com/mlhmz/finances/internal/db"
	"github.com/mlhmz/finances/internal/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) {
	t.Helper()
	testDB, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	if err := testDB.AutoMigrate(&models.User{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	db.DB = testDB
}

func seedUser(t *testing.T, id, email, fullName, currency string) models.User {
	t.Helper()
	u := models.User{
		ID:       id,
		Email:    email,
		FullName: fullName,
		Currency: currency,
		Initials: models.DeriveInitials(fullName),
	}
	if res := db.DB.Create(&u); res.Error != nil {
		t.Fatalf("seed user: %v", res.Error)
	}
	return u
}

func TestUserRepository_Get(t *testing.T) {
	setupTestDB(t)
	seedUser(t, "u1", "alice@example.com", "Alice Smith", "EUR")
	seedUser(t, "u2", "bob@example.com", "Bob Jones", "EUR")

	repo := NewUserRepository("u1")
	got, err := repo.Get()
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.ID != "u1" {
		t.Errorf("expected ID u1, got %s", got.ID)
	}
	if got.FullName != "Alice Smith" {
		t.Errorf("expected FullName 'Alice Smith', got %s", got.FullName)
	}
}

func TestUserRepository_Get_IsolatesUsers(t *testing.T) {
	setupTestDB(t)
	seedUser(t, "u1", "alice@example.com", "Alice Smith", "EUR")
	seedUser(t, "u2", "bob@example.com", "Bob Jones", "EUR")

	// u2's repo must not return u1's record
	repo := NewUserRepository("u2")
	got, err := repo.Get()
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.ID != "u2" {
		t.Errorf("expected ID u2, got %s", got.ID)
	}
}

func TestUserRepository_Get_NotFound(t *testing.T) {
	setupTestDB(t)

	repo := NewUserRepository("nonexistent")
	_, err := repo.Get()
	if err == nil {
		t.Error("expected error for non-existent user, got nil")
	}
}

func TestUserRepository_Update(t *testing.T) {
	setupTestDB(t)
	seedUser(t, "u1", "alice@example.com", "Alice Smith", "EUR")

	repo := NewUserRepository("u1")
	if err := repo.Update("Alice Wonder", "EUR"); err != nil {
		t.Fatalf("Update: %v", err)
	}

	got, _ := repo.Get()
	if got.FullName != "Alice Wonder" {
		t.Errorf("expected FullName 'Alice Wonder', got %s", got.FullName)
	}
	if got.Initials != "AW" {
		t.Errorf("expected Initials 'AW', got %s", got.Initials)
	}
}

func TestUserRepository_Update_DoesNotAffectOtherUsers(t *testing.T) {
	setupTestDB(t)
	seedUser(t, "u1", "alice@example.com", "Alice Smith", "EUR")
	seedUser(t, "u2", "bob@example.com", "Bob Jones", "EUR")

	NewUserRepository("u1").Update("Alice Updated", "EUR") //nolint

	repoB := NewUserRepository("u2")
	got, _ := repoB.Get()
	if got.FullName != "Bob Jones" {
		t.Errorf("u2 should be unchanged, got %s", got.FullName)
	}
}
