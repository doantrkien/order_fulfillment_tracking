package aiclient

import (
	"fmt"
	"main/pkg/gemini"
	"main/pkg/groq"
	"os"
	"strings"
)

// NewFromEnv reads the AI_PROVIDER environment variable and returns the
// corresponding AIClient implementation.
//
// Supported values (case-insensitive):
//
//	AI_PROVIDER=GEMINI  → uses pkg/gemini  (default, requires GEMINI_API_KEY + GEMINI_AI_MODEL)
//	AI_PROVIDER=GROQ    → uses pkg/groq    (requires GROQ_API_KEY, optional GROQ_MODEL)
func NewFromEnv() (AIClient, error) {
	provider := strings.ToUpper(strings.TrimSpace(os.Getenv("AI_PROVIDER")))
	if provider == "" {
		provider = "GEMINI"
	}

	fmt.Printf("[INFO][aiclient.NewFromEnv] AI_PROVIDER=%s\n", provider)

	switch provider {
	case "GEMINI":
		return gemini.NewClient()

	case "GROQ":
		return groq.NewClient()

	default:
		return nil, fmt.Errorf("unsupported AI_PROVIDER %q — supported values: GEMINI, GROQ", provider)
	}
}

