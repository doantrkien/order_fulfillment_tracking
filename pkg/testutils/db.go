package testutils

import (
	"log"
	"main/internal/models"
	"main/pkg/postgresql"
	"path/filepath"
	"runtime"

	"github.com/joho/godotenv"
	"gorm.io/gorm"
)

// SetupTestDB initializes a test database connection.
// It loads the .env file from the root directory.
func SetupTestDB() *gorm.DB {
	// Find the project root relative to this file
	_, b, _, _ := runtime.Caller(0)
	rootPath := filepath.Join(filepath.Dir(b), "../../")
	envPath := filepath.Join(rootPath, ".env")

	if err := godotenv.Load(envPath); err != nil {
		log.Printf("No .env file found at %s, using process environment", envPath)
	}

	db, err := postgresql.ConnectDB()
	if err != nil {
		log.Fatalf("Failed to connect to test database: %v", err)
	}

	// Auto migrate schema
	if err := db.AutoMigrate(&models.Order{}); err != nil {
		log.Fatalf("Failed to auto-migrate: %v", err)
	}

	return db
}

// CleanDatabase truncates the orders table.
func CleanDatabase(db *gorm.DB) {
	db.Exec("TRUNCATE TABLE orders RESTART IDENTITY CASCADE;")
}
