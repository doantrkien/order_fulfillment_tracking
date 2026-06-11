package gemini

import (
	"context"
	"fmt"
	"main/errs"
	"os"
	"strconv"
	"time"

	"google.golang.org/genai"
)

type Client struct {
	genaiClient  *genai.Client
	modelName    string
	timeout      time.Duration
	retryLimit   int
	enabled      bool
	maxInputSize int
}

func NewClient() (*Client, error) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	fmt.Printf("[DEBUG][gemini.NewClient] GEMINI_API_KEY set: %v\n", apiKey != "")
	if apiKey == "" {
		fmt.Println("[DEBUG][gemini.NewClient] ERROR: GEMINI_API_KEY is empty")
		return nil, errs.ERR_GEMINI_API_KEY_EMPTY
	}

	modelName := os.Getenv("GEMINI_AI_MODEL")
	if modelName == "" {
		modelName = "gemini-2.0-flash"
	}
	fmt.Printf("[DEBUG][gemini.NewClient] Model: %s\n", modelName)

	timeout := 10 * time.Second
	if timeoutStr := os.Getenv("AI_TIMEOUT_MS"); timeoutStr != "" {
		if ms, err := strconv.Atoi(timeoutStr); err == nil {
			timeout = time.Duration(ms) * time.Millisecond
		}
	}
	fmt.Printf("[DEBUG][gemini.NewClient] Timeout: %v\n", timeout)

	retryLimit := 5
	if retryStr := os.Getenv("AI_RETRY_LIMIT"); retryStr != "" {
		if value, err := strconv.Atoi(retryStr); err == nil {
			retryLimit = value
		}
	}
	fmt.Printf("[DEBUG][gemini.NewClient] RetryLimit: %d\n", retryLimit)

	enabled := true
	if enabledStr := os.Getenv("AI_ENABLED"); enabledStr != "" {
		if value, err := strconv.ParseBool(enabledStr); err == nil {
			enabled = value
		}
	}
	fmt.Printf("[DEBUG][gemini.NewClient] AI Enabled: %v\n", enabled)

	maxInputSize := 4000
	if sizeStr := os.Getenv("AI_MAX_INPUT_SIZE"); sizeStr != "" {
		if value, err := strconv.Atoi(sizeStr); err == nil {
			maxInputSize = value
		}
	}
	fmt.Printf("[DEBUG][gemini.NewClient] MaxInputSize: %d\n", maxInputSize)

	ctx := context.Background()

	fmt.Println("[DEBUG][gemini.NewClient] Creating genai client...")
	genaiClient, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey: apiKey,
	})

	if err != nil {
		fmt.Printf("[DEBUG][gemini.NewClient] ERROR creating genai client: %v\n", err)
		return nil, errs.ERR_GEMINI_CLIENT_CREATE_FAILED
	}

	fmt.Println("[DEBUG][gemini.NewClient] genai client created successfully")

	return &Client{
		genaiClient:  genaiClient,
		modelName:    modelName,
		timeout:      timeout,
		retryLimit:   retryLimit,
		enabled:      enabled,
		maxInputSize: maxInputSize,
	}, nil
}

func (c *Client) GenerateContent(
	ctx context.Context,
	prompt string,
) (string, error) {
	fmt.Printf("[DEBUG][gemini.GenerateContent] Called. Enabled: %v, Model: %s, PromptLen: %d\n", c.enabled, c.modelName, len(prompt))

	if !c.enabled {
		fmt.Println("[DEBUG][gemini.GenerateContent] AI is DISABLED, returning error")
		return "", errs.ERR_AI_DISABLED
	}

	if len(prompt) > c.maxInputSize {
		fmt.Printf("[DEBUG][gemini.GenerateContent] Prompt too large: %d > %d\n", len(prompt), c.maxInputSize)
		return "", errs.ERR_AI_INPUT_TOO_LARGE
	}

	for attempt := 0; attempt <= c.retryLimit; attempt++ {
		fmt.Printf("[DEBUG][gemini.GenerateContent] Attempt %d/%d, timeout: %v\n", attempt+1, c.retryLimit+1, c.timeout)
		timeoutCtx, cancel := context.WithTimeout(context.Background(), c.timeout)

		resp, err := c.genaiClient.Models.GenerateContent(
			timeoutCtx,
			c.modelName,
			genai.Text(prompt),
			nil,
		)

		cancel()

		if err != nil {
			fmt.Printf("[DEBUG][gemini.GenerateContent] Attempt %d ERROR: %v\n", attempt+1, err)

			if attempt < c.retryLimit {
				fmt.Printf("[DEBUG][gemini.GenerateContent] Retrying in 1s...\n")
				time.Sleep(time.Second)
			}

			continue
		}

		if resp == nil {
			fmt.Printf("[DEBUG][gemini.GenerateContent] Attempt %d: resp is nil\n", attempt+1)

			if attempt < c.retryLimit {
				time.Sleep(time.Second)
			}

			continue
		}

		text := resp.Text()

		if text == "" {
			fmt.Printf("[DEBUG][gemini.GenerateContent] Attempt %d: resp.Text() is empty\n", attempt+1)

			if attempt < c.retryLimit {
				time.Sleep(time.Second)
			}

			continue
		}

		fmt.Printf("[DEBUG][gemini.GenerateContent] Success on attempt %d, responseLen: %d\n", attempt+1, len(text))
		return text, nil
	}

	fmt.Println("[DEBUG][gemini.GenerateContent] All attempts exhausted, returning ERR_GEMINI_GENERATE_CONTENT_FAILED")
	return "", errs.ERR_GEMINI_GENERATE_CONTENT_FAILED
}
