package embedding

import (
	"context"
	"fmt"
	"os"

	"google.golang.org/genai"
)

type geminiEmbeddingClient struct {
	client    *genai.Client
	modelName string
}

// NewGeminiEmbeddingClient creates an embedding client using Gemini's text-embedding-004 model.
// It requires GEMINI_API_KEY to be set. The model outputs 768 dimensions which matches
// the pgvector schema (vector(768)).
func NewGeminiEmbeddingClient() (Client, error) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("GEMINI_API_KEY is required for Gemini embedding client")
	}

	ctx := context.Background()
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini client: %w", err)
	}

	modelName := os.Getenv("GEMINI_EMBEDDING_MODEL")
	if modelName == "" {
		modelName = "text-embedding-004" // 768 dims, matches DB schema
	}

	fmt.Printf("[INFO][embedding.NewGeminiEmbeddingClient] Embedding client ready. Model: %s\n", modelName)

	return &geminiEmbeddingClient{
		client:    client,
		modelName: modelName,
	}, nil
}

func (c *geminiEmbeddingClient) Embed(ctx context.Context, text string) ([]float32, error) {
	contents := []*genai.Content{
		genai.NewContentFromText(text, genai.RoleUser),
	}

	result, err := c.client.Models.EmbedContent(ctx, c.modelName, contents, nil)
	if err != nil {
		return nil, fmt.Errorf("gemini embed error: %w", err)
	}

	if result.Embeddings == nil || len(result.Embeddings) == 0 {
		return nil, fmt.Errorf("empty embedding returned from Gemini")
	}

	return result.Embeddings[0].Values, nil
}
