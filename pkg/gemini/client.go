package gemini

import (
	"context"
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
	if apiKey == "" {
		return nil, errs.ERR_GEMINI_API_KEY_EMPTY
	}

	modelName := os.Getenv("GEMINI_AI_MODEL")
	if modelName == "" {
		modelName = "gemini-2.5-flash"
	}

	timeout := 10 * time.Second
	if timeoutStr := os.Getenv("AI_TIMEOUT_MS"); timeoutStr != "" {
		if ms, err := strconv.Atoi(timeoutStr); err == nil {
			timeout = time.Duration(ms) * time.Millisecond
		}
	}

	retryLimit := 5
	if retryStr := os.Getenv("AI_RETRY_LIMIT"); retryStr != "" {
		if value, err := strconv.Atoi(retryStr); err == nil {
			retryLimit = value
		}
	}

	enabled := true
	if enabledStr := os.Getenv("AI_ENABLED"); enabledStr != "" {
		if value, err := strconv.ParseBool(enabledStr); err == nil {
			enabled = value
		}
	}

	maxInputSize := 4000
	if sizeStr := os.Getenv("AI_MAX_INPUT_SIZE"); sizeStr != "" {
		if value, err := strconv.Atoi(sizeStr); err == nil {
			maxInputSize = value
		}
	}

	ctx := context.Background()

	genaiClient, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey: apiKey,
	})
	if err != nil {
		return nil, errs.ERR_GEMINI_CLIENT_CREATE_FAILED
	}

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

	if !c.enabled {
		return "", errs.ERR_AI_DISABLED
	}

	if len(prompt) > c.maxInputSize {
		return "", errs.ERR_AI_INPUT_TOO_LARGE
	}

	for attempt := 0; attempt <= c.retryLimit; attempt++ {

		timeoutCtx, cancel := context.WithTimeout(ctx, c.timeout)

		resp, err := c.genaiClient.Models.GenerateContent(
			timeoutCtx,
			c.modelName,
			genai.Text(prompt),
			nil,
		)

		cancel()

		if err != nil {

			if attempt < c.retryLimit {
				time.Sleep(time.Second)
			}

			continue
		}

		if resp == nil {

			if attempt < c.retryLimit {
				time.Sleep(time.Second)
			}

			continue
		}

		text := resp.Text()

		if text == "" {

			if attempt < c.retryLimit {
				time.Sleep(time.Second)
			}

			continue
		}

		return text, nil
	}

	return "", errs.ERR_GEMINI_GENERATE_CONTENT_FAILED
}
