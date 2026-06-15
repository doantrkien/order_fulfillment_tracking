package aiclient

import "context"

// AIClient is the common interface implemented by every LLM provider (Gemini, Ollama, …).
// Any provider that satisfies this interface can be plugged in via the AI_PROVIDER env var.
type AIClient interface {
	// GenerateContent sends prompt to the underlying model and returns the raw text response.
	GenerateContent(ctx context.Context, prompt string) (string, error)
}
