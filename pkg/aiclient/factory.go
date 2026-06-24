package aiclient

import (
	"fmt"
	"os"
	"strings"

	"main/pkg/gemini"
	"main/pkg/groq"
	"main/pkg/ollama"
)

// NewFromEnv reads the AI_PROVIDER environment variable and returns the
// corresponding AIClient implementation.
//
// Supported values (case-insensitive):
//
//	AI_PROVIDER=GEMINI  → uses pkg/gemini  (default, requires GEMINI_API_KEY + GEMINI_AI_MODEL)
//	AI_PROVIDER=GROQ    → uses pkg/groq    (requires GROQ_API_KEY, optional GROQ_MODEL)
//	AI_PROVIDER=OLLAMA  → uses pkg/ollama  (requires OLLAMA_BASE_URL, OLLAMA_MODEL)
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

	case "OLLAMA":
		return ollama.NewClient()

	default:
		return nil, fmt.Errorf("unsupported AI_PROVIDER %q — supported values: GEMINI, GROQ, OLLAMA", provider)
	}
}

