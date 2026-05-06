package configs

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

func LoadConfig() error {
	envFile := ".env." + os.Getenv("APP_ENV")

	fmt.Println("file env: ", envFile)

	err := godotenv.Load(envFile)
	if err != nil {
		return fmt.Errorf("error loading %s: %w", envFile, err)
	}

	return nil
}
