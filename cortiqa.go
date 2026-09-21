package cortiqa

import (
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
	// Version of the Go SDK.
	Version = "0.1.0"
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

// Client is the main Cortiqa API client.
type Client struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
	maxRetries int

	// Services
	Chat     *ChatService
	Messages *MessagesService
	Models   *ModelsService
}

// NewClient creates a new Cortiqa client.
// If apiKey is empty, it reads from the CORTIQA_API_KEY environment variable.
func NewClient(apiKey string, opts ...Option) *Client {
	if apiKey == "" {
		apiKey = os.Getenv("CORTIQA_API_KEY")
	}

	c := &Client{
		apiKey:     apiKey,
		baseURL:    DefaultBaseURL,
		httpClient: &http.Client{Timeout: DefaultTimeout},
		maxRetries: 2,
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
