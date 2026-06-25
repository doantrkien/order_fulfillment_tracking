package groq

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"main/errs"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

type Client struct {
	apiKey       string
	modelName    string
	baseURL      string
	timeout      time.Duration
	retryLimit   int
	enabled      bool
	maxInputSize int
	httpClient   *http.Client
}

func NewClient() (*Client, error) {
	apiKey := os.Getenv("GROQ_API_KEY")
	fmt.Printf("[DEBUG][groq.NewClient] GROQ_API_KEY set: %v\n", apiKey != "")
	if apiKey == "" {
		fmt.Println("[DEBUG][groq.NewClient] ERROR: GROQ_API_KEY is empty")
		return nil, errs.ERR_GEMINI_API_KEY_EMPTY
	}

	modelName := os.Getenv("GROQ_MODEL")
	if modelName == "" {
		modelName = "meta-llama/llama-4-scout-17b-16e-instruct"
	}
	fmt.Printf("[DEBUG][groq.NewClient] Model: %s\n", modelName)

	baseURL := os.Getenv("GROQ_BASE_URL")
	if baseURL == "" {
		baseURL = "https://api.groq.com"
	}

	timeout := 60 * time.Second
	if v := os.Getenv("AI_TIMEOUT_MS"); v != "" {
		if ms, err := strconv.Atoi(v); err == nil {
			timeout = time.Duration(ms) * time.Millisecond
		}
	}
	fmt.Printf("[DEBUG][groq.NewClient] Timeout: %v\n", timeout)

	retryLimit := 3
	if v := os.Getenv("AI_RETRY_LIMIT"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			retryLimit = n
		}
	}
	fmt.Printf("[DEBUG][groq.NewClient] RetryLimit: %d\n", retryLimit)

	enabled := true
	if v := os.Getenv("AI_ENABLED"); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			enabled = b
		}
	}
	fmt.Printf("[DEBUG][groq.NewClient] AI Enabled: %v\n", enabled)

	maxInputSize := 4000
	if v := os.Getenv("AI_MAX_INPUT_SIZE"); v != "" {
		if n, err := strconv.Atoi(strings.TrimSpace(v)); err == nil {
			maxInputSize = n
		} else {
			fmt.Printf("[DEBUG][groq.NewClient] Invalid AI_MAX_INPUT_SIZE: '%s'\n", v)
		}
	}
	fmt.Printf("[DEBUG][groq.NewClient] MaxInputSize: %d\n", maxInputSize)

	fmt.Println("[DEBUG][groq.NewClient] groq client created successfully")

	return &Client{
		apiKey:       apiKey,
		modelName:    modelName,
		baseURL:      baseURL,
		timeout:      timeout,
		retryLimit:   retryLimit,
		enabled:      enabled,
		maxInputSize: maxInputSize,
		httpClient:   &http.Client{},
	}, nil
}

func (c *Client) GenerateContent(ctx context.Context, prompt string) (string, error) {
	fmt.Printf("[DEBUG][groq.GenerateContent] Called. Enabled: %v, Model: %s, PromptLen: %d\n",
		c.enabled, c.modelName, len(prompt))

	if !c.enabled {
		fmt.Println("[DEBUG][groq.GenerateContent] AI is DISABLED, returning error")
		return "", errs.ERR_AI_DISABLED
	}

	if len(prompt) > c.maxInputSize {
		fmt.Printf("[DEBUG][groq.GenerateContent] Prompt too large: %d > %d\n", len(prompt), c.maxInputSize)
		return "", errs.ERR_AI_INPUT_TOO_LARGE
	}

	body, err := json.Marshal(map[string]any{
		"model": c.modelName,
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
	})
	if err != nil {
		return "", errs.ERR_INTERNAL_SERVER
	}

	for attempt := 0; attempt <= c.retryLimit; attempt++ {
		fmt.Printf("[DEBUG][groq.GenerateContent] Attempt %d/%d, timeout: %v\n", attempt+1, c.retryLimit+1, c.timeout)
		timeoutCtx, cancel := context.WithTimeout(context.Background(), c.timeout)

		var chatEndpoint = os.Getenv("GROQ_ENDPOINT")
		if chatEndpoint == "" {
			chatEndpoint = "/openai/v1/chat/completions"
		}

		req, reqErr := http.NewRequestWithContext(
			timeoutCtx, http.MethodPost,
			c.baseURL+chatEndpoint,
			bytes.NewReader(body),
		)
		if reqErr != nil {
			cancel()
			fmt.Printf("[DEBUG][groq.GenerateContent] Attempt %d ERROR building request: %v\n", attempt+1, reqErr)
			if attempt < c.retryLimit {
				fmt.Println("[DEBUG][groq.GenerateContent] Retrying in 1s...")
				time.Sleep(time.Second)
			}
			continue
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+c.apiKey)

		resp, doErr := c.httpClient.Do(req)
		cancel()

		if doErr != nil {
			fmt.Printf("[DEBUG][groq.GenerateContent] Attempt %d ERROR: %v\n", attempt+1, doErr)
			if attempt < c.retryLimit {
				fmt.Println("[DEBUG][groq.GenerateContent] Retrying in 1s...")
				time.Sleep(time.Second)
			}
			continue
		}

		if resp == nil {
			fmt.Printf("[DEBUG][groq.GenerateContent] Attempt %d: resp is nil\n", attempt+1)
			if attempt < c.retryLimit {
				time.Sleep(time.Second)
			}
			continue
		}

		rawBody, readErr := io.ReadAll(resp.Body)
		resp.Body.Close()

		if readErr != nil || resp.StatusCode != http.StatusOK {
			fmt.Printf("[DEBUG][groq.GenerateContent] Attempt %d: bad status=%d\n", attempt+1, resp.StatusCode)
			if attempt < c.retryLimit {
				waitSec := 2
				if resp.StatusCode == http.StatusTooManyRequests {
					// Respect Retry-After header if present
					if ra := resp.Header.Get("Retry-After"); ra != "" {
						if secs, err := strconv.Atoi(strings.TrimSpace(ra)); err == nil && secs > 0 {
							waitSec = secs + 1
						}
					} else {
						waitSec = 10
					}
					fmt.Printf("[DEBUG][groq.GenerateContent] Rate limited (429), waiting %ds before retry...\n", waitSec)
				}
				time.Sleep(time.Duration(waitSec) * time.Second)
			}
			continue
		}

		// parse only the fields we need — same spirit as resp.Text() in Gemini
		var parsed struct {
			Choices []struct {
				Message struct {
					Content string `json:"content"`
				} `json:"message"`
			} `json:"choices"`
			Error *struct {
				Message string `json:"message"`
			} `json:"error"`
		}
		if jsonErr := json.Unmarshal(rawBody, &parsed); jsonErr != nil {
			fmt.Printf("[DEBUG][groq.GenerateContent] Attempt %d: JSON parse error: %v\n", attempt+1, jsonErr)
			if attempt < c.retryLimit {
				time.Sleep(time.Second)
			}
			continue
		}

		if parsed.Error != nil {
			fmt.Printf("[DEBUG][groq.GenerateContent] Attempt %d: Groq API error: %s\n", attempt+1, parsed.Error.Message)
			if attempt < c.retryLimit {
				time.Sleep(time.Second)
			}
			continue
		}

		text := ""
		if len(parsed.Choices) > 0 {
			text = parsed.Choices[0].Message.Content
		}

		if text == "" {
			fmt.Printf("[DEBUG][groq.GenerateContent] Attempt %d: resp.Text() is empty\n", attempt+1)
			if attempt < c.retryLimit {
				time.Sleep(time.Second)
			}
			continue
		}

		fmt.Printf("[DEBUG][groq.GenerateContent] Success on attempt %d, responseLen: %d\n", attempt+1, len(text))
		return text, nil
	}

	fmt.Println("[DEBUG][groq.GenerateContent] All attempts exhausted, returning ERR_GEMINI_GENERATE_CONTENT_FAILED")
	return "", errs.ERR_GEMINI_GENERATE_CONTENT_FAILED
}
