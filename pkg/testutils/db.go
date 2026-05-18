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

func SetupTestDB() *gorm.DB {

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

	if err := db.AutoMigrate(&models.Order{}); err != nil {
		log.Fatalf("Failed to auto-migrate: %v", err)
	}

	return db
}

func CleanDatabase(db *gorm.DB) {
	db.Exec("TRUNCATE TABLE orders RESTART IDENTITY CASCADE;")
}
