package cortiqa

import (
	"context"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	// DefaultBaseURL is the official Cortiqa API endpoint.
	DefaultBaseURL = "https://api.cortiqa.co"
	// DefaultTimeout for HTTP requests.
	DefaultTimeout = 60 * time.Second
	// DefaultModel is the default foundation model used across completions.
	DefaultModel = "openai/gpt-oss-120b"
	// Version of the Go SDK.
	Version = "0.1.1"
)

// Option represents a configuration option for the Client.
type Option func(*Client)

// WithBaseURL overrides the default API base URL.
func WithBaseURL(url string) Option {
	return func(c *Client) {
		c.baseURL = strings.TrimRight(url, "/")
	}
}

// WithTimeout sets a custom request timeout.
func WithTimeout(d time.Duration) Option {
	return func(c *Client) {
		c.httpClient.Timeout = d
	}
}

// WithHTTPClient sets a custom http.Client.
func WithHTTPClient(httpClient *http.Client) Option {
	return func(c *Client) {
		if httpClient != nil {
			c.httpClient = httpClient
		}
	}
}

// WithMaxRetries sets the number of retry attempts for 429/5xx errors.
func WithMaxRetries(retries int) Option {
	return func(c *Client) {
		c.maxRetries = retries
	}
}

// WithDefaultModel sets the default model to use when omitted in requests.
func WithDefaultModel(model string) Option {
	return func(c *Client) {
		if model != "" {
			c.defaultModel = model
		}
	}
}

// Client is the main Cortiqa API client.
type Client struct {
	apiKey       string
	baseURL      string
	defaultModel string
	httpClient   *http.Client
	maxRetries   int

	// Services
	Chat     *ChatService
	Messages *MessagesService
	Models   *ModelsService
}

// DefaultModel returns the configured default model.
func (c *Client) DefaultModel() string {
	return c.defaultModel
}

// PromptOption defines optional parameters for the Prompt shortcut.
type PromptOption func(*ChatCompletionRequest)

// WithSystem sets a system prompt for the Prompt shortcut.
func WithSystem(system string) PromptOption {
	return func(req *ChatCompletionRequest) {
		if system != "" {
			req.Messages = append([]ChatMessage{{Role: "system", Content: system}}, req.Messages...)
		}
	}
}

// WithPromptModel overrides the model used for the Prompt call.
func WithPromptModel(model string) PromptOption {
	return func(req *ChatCompletionRequest) {
		req.Model = model
	}
}

// Prompt is a one-liner shortcut to send a prompt and get the assistant response text directly.
func (c *Client) Prompt(ctx context.Context, prompt string, opts ...PromptOption) (string, error) {
	req := ChatCompletionRequest{
		Model: c.defaultModel,
		Messages: []ChatMessage{
			{Role: "user", Content: prompt},
		},
	}
	for _, opt := range opts {
		opt(&req)
	}
	resp, err := c.Chat.Create(ctx, req)
	if err != nil {
		return "", err
	}
	return resp.Content(), nil
}

// NewClient creates a new Cortiqa client.
// If apiKey is empty, it reads from the CORTIQA_API_KEY environment variable.
func NewClient(apiKey string, opts ...Option) *Client {
	if apiKey == "" {
		apiKey = os.Getenv("CORTIQA_API_KEY")
	}

	c := &Client{
		apiKey:       apiKey,
		baseURL:      DefaultBaseURL,
		defaultModel: DefaultModel,
		httpClient:   &http.Client{Timeout: DefaultTimeout},
		maxRetries:   2,
	}

	for _, opt := range opts {
		opt(c)
	}

	// Initialize services
	c.Chat = &ChatService{client: c}
	c.Messages = &MessagesService{client: c}
	c.Models = &ModelsService{client: c}

	return c
}
