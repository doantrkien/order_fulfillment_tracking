package configs

import (
	"errors"
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// AIConfig holds AI-related runtime configuration.
type AIConfig struct {
	Enabled   bool // Whether AI is enabled (default: true). Set AI_ENABLED=false to disable.
	TimeoutMs int  // AI call timeout in milliseconds (default: 10000).
}

// LoadAIConfig reads AI configuration from environment variables.
func LoadAIConfig() AIConfig {
	enabled := os.Getenv("AI_ENABLED") != "false" // default: true
	timeoutMs := 5000                              // default: 5 seconds
	if v := os.Getenv("AI_TIMEOUT_MS"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed > 0 {
			timeoutMs = parsed
		}
	}
	return AIConfig{Enabled: enabled, TimeoutMs: timeoutMs}
}

func LoadConfig() error {
	envFile := os.Getenv("ENV_FILE")
	var candidates []string
	if envFile != "" {
		candidates = append(candidates, envFile)
	}
	candidates = append(candidates, ".env.local", ".env")

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
