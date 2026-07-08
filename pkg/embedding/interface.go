package embedding

import "context"

// Client is the interface for generating text embeddings.
type Client interface {
	// Embed returns a float32 vector representation of the given text.
	// Returns an error if the embedding service is unavailable or fails.
	Embed(ctx context.Context, text string) ([]float32, error)
}
