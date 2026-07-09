package embedding

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

type ollamaEmbeddingClient struct {
	baseURL   string
	apiKey    string
	modelName string
	client    *http.Client
}

// NewOllamaEmbeddingClient creates an embedding client compatible with OpenAI's /v1/embeddings endpoint
// (which Ollama supports). It requires the model to output 768 dimensions (like nomic-embed-text)
// to match the pgvector schema.
func NewOllamaEmbeddingClient() (Client, error) {
	baseURL := os.Getenv("OLLAMA_EMBED_BASE_URL")
	if baseURL == "" {
		baseURL = os.Getenv("OLLAMA_BASE_URL")
		if baseURL == "" || baseURL == "https://ollama.com" {
			baseURL = "http://localhost:11434" // Default local ollama
		}
	}

	modelName := os.Getenv("OLLAMA_EMBEDDING_MODEL")
	if modelName == "" {
		// nomic-embed-text outputs 768 dims which perfectly matches our DB
		modelName = "nomic-embed-text"
	}

	apiKey := os.Getenv("OLLAMA_API_KEY")

	fmt.Printf("[INFO][embedding.NewOllamaEmbeddingClient] Embedding client ready. Model: %s, BaseURL: %s\n", modelName, baseURL)
	
	return &ollamaEmbeddingClient{
		baseURL:   baseURL,
		apiKey:    apiKey,
		modelName: modelName,
		client:    &http.Client{},
	}, nil
}

func (c *ollamaEmbeddingClient) Embed(ctx context.Context, text string) ([]float32, error) {
	reqBody, err := json.Marshal(map[string]any{
		"model": c.modelName,
		"input": text,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	endpoint := os.Getenv("OLLAMA_EMBED_ENDPOINT")
	if endpoint == "" {
		endpoint = "/v1/embeddings"
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+endpoint, bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http request failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ollama embed error: status=%d body=%s", resp.StatusCode, string(body))
	}

	var parsed struct {
		Data []struct {
			Embedding []float32 `json:"embedding"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(parsed.Data) == 0 || len(parsed.Data[0].Embedding) == 0 {
		return nil, fmt.Errorf("empty embedding returned")
	}

	return parsed.Data[0].Embedding, nil
}
