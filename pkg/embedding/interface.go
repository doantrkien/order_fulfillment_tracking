package embedding

import (
	"context"
)

// Client is the interface for generating text embeddings.
type Client interface {
	// Embed returns a float32 vector representation of the given text.
	// Returns an error if the embedding service is unavailable or fails.
	Embed(ctx context.Context, text string) ([]float32, error)
}

// NewEmbeddingClientFromEnv creates the appropriate embedding client based on
// environment configuration. It checks EMBEDDING_PROVIDER first, then falls
// back to AI_PROVIDER:
//
//   - "GEMINI" → Gemini text-embedding-004
//   - "OLLAMA" → Ollama nomic-embed-text (or custom model)
//
// Returns an error if the selected provider cannot be initialized.
func NewEmbeddingClientFromEnv() (Client, error) {
	return NewOllamaEmbeddingClient()
}
