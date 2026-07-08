package embedding

import (
	"context"
	"fmt"
	"os"

	"google.golang.org/genai"
)

const defaultEmbeddingModel = "models/text-embedding-004"

type geminiEmbeddingClient struct {
	client    *genai.Client
	modelName string
}

// NewGeminiEmbeddingClient creates an embedding client backed by Gemini text-embedding-004.
// Returns (nil, nil) if GEMINI_API_KEY is not set — callers must handle nil gracefully
// by falling back to keyword matching.
func NewGeminiEmbeddingClient() (Client, error) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		fmt.Println("[WARNING][embedding.NewGeminiEmbeddingClient] GEMINI_API_KEY not set — embedding client disabled")
		return nil, nil //nolint:nilnil
	}

	modelName := os.Getenv("GEMINI_EMBEDDING_MODEL")
	if modelName == "" {
		modelName = defaultEmbeddingModel
	}

	client, err := genai.NewClient(context.Background(), &genai.ClientConfig{
		APIKey: apiKey,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini client for embedding: %w", err)
	}

	fmt.Printf("[INFO][embedding.NewGeminiEmbeddingClient] Embedding client ready. Model: %s\n", modelName)
	return &geminiEmbeddingClient{client: client, modelName: modelName}, nil
}

// Embed returns a float32 vector for the given text using Gemini text-embedding-004.
// The returned slice has 768 dimensions (for text-embedding-004).
func (c *geminiEmbeddingClient) Embed(ctx context.Context, text string) ([]float32, error) {
	contents := []*genai.Content{
		genai.NewContentFromText(text, genai.RoleUser),
	}

	result, err := c.client.Models.EmbedContent(ctx, c.modelName, contents, nil)
	if err != nil {
		return nil, fmt.Errorf("gemini EmbedContent failed: %w", err)
	}
	if result == nil || len(result.Embeddings) == 0 || result.Embeddings[0] == nil {
		return nil, fmt.Errorf("gemini EmbedContent returned empty embeddings")
	}
	return result.Embeddings[0].Values, nil
}
