package configs

import (
	"errors"
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

func LoadConfig() error {
	envFile := os.Getenv("ENV_FILE")
	var candidates []string
	if envFile != "" {
		candidates = append(candidates, envFile)
	}
	candidates = append(candidates, ".env")

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
