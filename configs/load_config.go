package configs

import (
	"errors"
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

// LoadConfig loads dotenv files into the process environment.
// When APP_ENV is set (e.g. production), it tries .env.{APP_ENV} first, then .env.
// When APP_ENV is empty, only .env is used (not ".env." from concatenation).
// If no file exists, configuration continues using only the OS environment.
func LoadConfig() error {
	appEnv := os.Getenv("APP_ENV")
	var candidates []string
	if appEnv != "" {
		candidates = append(candidates, ".env."+appEnv)
	}
	candidates = append(candidates, ".env")

	fmt.Println(candidates)

	for _, path := range candidates {
		err := godotenv.Load(path)
		if err == nil {
			fmt.Println("loaded env file:", path)
			return nil
		}
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		return fmt.Errorf("error loading %s: %w", path, err)
	}
	return nil
}
