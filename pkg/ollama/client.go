package ollama

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"main/errs"
)

// Client implements aiclient.AIClient for a locally running Ollama instance.
type Client struct {
	baseURL      string
	modelName    string
	timeout      time.Duration
	retryLimit   int
	enabled      bool
	maxInputSize int
	httpClient   *http.Client
}

// NewClient reads configuration from environment variables and returns a
// ready-to-use Ollama Client.
//
// Environment variables:
//
//	OLLAMA_BASE_URL   – base URL of the Ollama server (default: http://localhost:11434)
//	OLLAMA_MODEL      – model tag to use              (default: qwen3.5:latest)
//	AI_TIMEOUT_MS     – per-request timeout in ms     (default: 30000)
//	AI_RETRY_LIMIT    – max retry attempts             (default: 3)
//	AI_ENABLED        – enable/disable AI calls        (default: true)
//	AI_MAX_INPUT_SIZE – max allowed prompt length      (default: 4000)
func NewClient() (*Client, error) {
	baseURL := os.Getenv("OLLAMA_BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:11434"
	}
	fmt.Printf("[DEBUG][ollama.NewClient] OLLAMA_BASE_URL: %s\n", baseURL)

	modelName := os.Getenv("OLLAMA_MODEL")
	if modelName == "" {
		modelName = "qwen3.5:latest"
	}
	fmt.Printf("[DEBUG][ollama.NewClient] Model: %s\n", modelName)

	timeout := 30 * time.Second
	if v := os.Getenv("AI_TIMEOUT_MS"); v != "" {
		if ms, err := strconv.Atoi(v); err == nil {
			timeout = time.Duration(ms) * time.Millisecond
		}
	}
	fmt.Printf("[DEBUG][ollama.NewClient] Timeout: %v\n", timeout)

	retryLimit := 3
	if v := os.Getenv("AI_RETRY_LIMIT"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			retryLimit = n
		}
	}
	fmt.Printf("[DEBUG][ollama.NewClient] RetryLimit: %d\n", retryLimit)

	enabled := true
	if v := os.Getenv("AI_ENABLED"); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			enabled = b
		}
	}
	fmt.Printf("[DEBUG][ollama.NewClient] AI Enabled: %v\n", enabled)

	maxInputSize := 4000
	if v := os.Getenv("AI_MAX_INPUT_SIZE"); v != "" {
		if n, err := strconv.Atoi(strings.TrimSpace(v)); err == nil {
			maxInputSize = n
		} else {
			fmt.Printf("[DEBUG][ollama.NewClient] Invalid AI_MAX_INPUT_SIZE: '%s'\n", v)
		}
	}
	fmt.Printf("[DEBUG][ollama.NewClient] MaxInputSize: %d\n", maxInputSize)

	fmt.Println("[DEBUG][ollama.NewClient] ollama client created successfully")

	return &Client{
		baseURL:      baseURL,
		modelName:    modelName,
		timeout:      timeout,
		retryLimit:   retryLimit,
		enabled:      enabled,
		maxInputSize: maxInputSize,
		httpClient:   &http.Client{},
	}, nil
}

// generateRequest is the JSON payload sent to POST /api/generate.
type generateRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
	Format string `json:"format"`
}

// generateResponse is the JSON response body returned by Ollama.
type generateResponse struct {
	Response string `json:"response"`
	Done     bool   `json:"done"`
	Error    string `json:"error,omitempty"`
}

// GenerateContent sends prompt to the Ollama /api/generate endpoint and
// returns the raw text response. It respects the context passed by the caller
// for timeout and cancellation — matching the Groq/Gemini retry pattern.
func (c *Client) GenerateContent(ctx context.Context, prompt string) (string, error) {
	fmt.Printf("[DEBUG][ollama.GenerateContent] Called. Enabled: %v, Model: %s, PromptLen: %d\n",
		c.enabled, c.modelName, len(prompt))

	if !c.enabled {
		fmt.Println("[DEBUG][ollama.GenerateContent] AI is DISABLED, returning error")
		return "", errs.ERR_AI_DISABLED
	}

	if len(prompt) > c.maxInputSize {
		fmt.Printf("[DEBUG][ollama.GenerateContent] Prompt too large: %d > %d\n", len(prompt), c.maxInputSize)
		return "", errs.ERR_AI_INPUT_TOO_LARGE
	}

	body, err := json.Marshal(generateRequest{
		Model:  c.modelName,
		Prompt: prompt,
		Stream: false,
	})
	if err != nil {
		return "", fmt.Errorf("ollama marshal request: %w", err)
	}

	endpoint := c.baseURL + "/api/generate"

	for attempt := 0; attempt <= c.retryLimit; attempt++ {
		fmt.Printf("[DEBUG][ollama.GenerateContent] Attempt %d/%d, timeout: %v\n",
			attempt+1, c.retryLimit+1, c.timeout)

		timeoutCtx, cancel := context.WithTimeout(ctx, c.timeout)

		req, reqErr := http.NewRequestWithContext(
			timeoutCtx,
			http.MethodPost,
			endpoint,
			bytes.NewReader(body),
		)
		if reqErr != nil {
			cancel()
			fmt.Printf("[DEBUG][ollama.GenerateContent] Attempt %d ERROR building request: %v\n",
				attempt+1, reqErr)
			if attempt < c.retryLimit {
				fmt.Println("[DEBUG][ollama.GenerateContent] Retrying in 1s...")
				time.Sleep(time.Second)
			}
			continue
		}
		req.Header.Set("Content-Type", "application/json")

		resp, doErr := c.httpClient.Do(req)

		if doErr != nil {
			cancel()
			fmt.Printf("[DEBUG][ollama.GenerateContent] Attempt %d ERROR: %v\n", attempt+1, doErr)
			if attempt < c.retryLimit {
				fmt.Println("[DEBUG][ollama.GenerateContent] Retrying in 1s...")
				time.Sleep(time.Second)
			}
			continue
		}

		if resp == nil {
			cancel()
			fmt.Printf("[DEBUG][ollama.GenerateContent] Attempt %d: resp is nil\n", attempt+1)
			if attempt < c.retryLimit {
				time.Sleep(time.Second)
			}
			continue
		}

		rawBody, readErr := io.ReadAll(resp.Body)
		resp.Body.Close()
		cancel() // Safe to cancel now since we've read the body


		if readErr != nil || resp.StatusCode != http.StatusOK {
			fmt.Printf("[DEBUG][ollama.GenerateContent] Attempt %d: bad status=%d, readErr=%v\n",
				attempt+1, resp.StatusCode, readErr)
			if attempt < c.retryLimit {
				time.Sleep(time.Second)
			}
			continue
		}

		var parsed generateResponse
		if jsonErr := json.Unmarshal(rawBody, &parsed); jsonErr != nil {
			fmt.Printf("[DEBUG][ollama.GenerateContent] Attempt %d: JSON parse error: %v\n",
				attempt+1, jsonErr)
			if attempt < c.retryLimit {
				time.Sleep(time.Second)
			}
			continue
		}

		if parsed.Error != "" {
			fmt.Printf("[DEBUG][ollama.GenerateContent] Attempt %d: Ollama API error: %s\n",
				attempt+1, parsed.Error)
			if attempt < c.retryLimit {
				time.Sleep(time.Second)
			}
			continue
		}

		if parsed.Response == "" {
			fmt.Printf("[DEBUG][ollama.GenerateContent] Attempt %d: response is empty\n", attempt+1)
			if attempt < c.retryLimit {
				time.Sleep(time.Second)
			}
			continue
		}

		fmt.Printf("[DEBUG][ollama.GenerateContent] Success on attempt %d, responseLen: %d\n",
			attempt+1, len(parsed.Response))
		return parsed.Response, nil
	}

	fmt.Println("[DEBUG][ollama.GenerateContent] All attempts exhausted, returning ERR_GEMINI_GENERATE_CONTENT_FAILED")
	return "", errs.ERR_GEMINI_GENERATE_CONTENT_FAILED
}
