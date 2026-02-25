package db

import (
	"github.com/mlhmz/finances/internal/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// DB is the application-wide GORM database instance.
var DB *gorm.DB

// Connect opens the SQLite database, auto-migrates all models, and stores the
// instance in the package-level DB variable.
func Connect(dbPath string) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	if err := db.AutoMigrate(&models.User{}, &models.OTPToken{}); err != nil {
		return nil, err
	}
	DB = db
	return db, nil
}
