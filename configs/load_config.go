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

	// Scoring parameters
	ScoreChannelSMS              int
	ScoreToneApologeticProactive int
	ScoreLongReason              int
	LikelyReasonLengthThreshold  int
	AIScoreThreshold             int
}

// getEnvInt reads an integer from the environment, returning defaultVal if empty or invalid.
func getEnvInt(key string, defaultVal int) int {
	if v := os.Getenv(key); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil {
			return parsed
		}
	}
	return defaultVal
}

// LoadAIConfig reads AI configuration from environment variables.
func LoadAIConfig() AIConfig {
	enabled := os.Getenv("AI_ENABLED") != "false" // default: true
	timeoutMs := getEnvInt("AI_TIMEOUT_MS", 10000)
	
	return AIConfig{
		Enabled:                      enabled,
		TimeoutMs:                    timeoutMs,
		ScoreChannelSMS:              getEnvInt("AI_SCORE_CHANNEL_SMS", 3),
		ScoreToneApologeticProactive: getEnvInt("AI_SCORE_TONE_APOLOGETIC_PROACTIVE", 3),
		ScoreLongReason:              getEnvInt("AI_SCORE_LONG_REASON", 1),
		LikelyReasonLengthThreshold:  getEnvInt("AI_REASON_LENGTH_THRESHOLD", 50),
		AIScoreThreshold:             getEnvInt("AI_SCORE_THRESHOLD", 3),
	}
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
		err := godotenv.Overload(path)
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
