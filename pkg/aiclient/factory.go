package aiclient

import (
	"fmt"
	"os"
	"strings"

	"main/pkg/gemini"
	"main/pkg/groq"
	"main/pkg/ollama"
)

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
