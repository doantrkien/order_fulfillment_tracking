package main

import (
	"fmt"
	"log"
	"main/configs"
	"os"
	"strconv"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	err := configs.LoadConfig()
	if err != nil {
		log.Println("Can not found file .env, system environment variables will be used")
	}

	dbURL := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		os.Getenv("DB_USERNAME"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_DATABASE"),
	)

	m, err := migrate.New("file://db/migrations", dbURL)
	if err != nil {
		log.Fatalf("Error initializing migration: %v", err)
	}

	cmd := "up"
	if len(os.Args) > 1 {
		cmd = os.Args[1]
	}

	switch cmd {
	case "up":
		if err := m.Up(); err != nil && err != migrate.ErrNoChange {
			log.Fatalf("Error running migration Up: %v", err)
		}
		log.Println("Migration Up run successfully!")
	case "down":
		if err := m.Down(); err != nil && err != migrate.ErrNoChange {
			log.Fatalf("Error running migration Down: %v", err)
		}
		log.Println("Migration Down run successfully!")
	case "force":
		if len(os.Args) < 3 {
			log.Fatalf("Usage: migrate force <version>")
		}
		version, err := strconv.Atoi(os.Args[2])
		if err != nil {
			log.Fatalf("Invalid version number: %v", err)
		}
		if err := m.Force(version); err != nil {
			log.Fatalf("Error forcing migration version: %v", err)
		}
		log.Printf("Forced migration version to %d successfully!", version)
	default:
		log.Fatalf("Invalid command. Please use 'up', 'down', or 'force <version>'")
	}
}
