package cortiqa

import (
	"fmt"
)

// APIError represents an error returned by the Cortiqa API.
type APIError struct {
	StatusCode int    `json:"status_code"`
	Message    string `json:"message"`
	Code       string `json:"code,omitempty"`
}

func (e *APIError) Error() string {
	if e.Code != "" {
		return fmt.Sprintf("cortiqa: API error (status %d, code %s): %s", e.StatusCode, e.Code, e.Message)
	}
	return fmt.Sprintf("cortiqa: API error (status %d): %s", e.StatusCode, e.Message)
}

// AuthenticationError is returned when an invalid or missing API key is used (HTTP 401).
type AuthenticationError struct {
	APIError
}

// RateLimitError is returned when rate limits are exceeded (HTTP 429).
type RateLimitError struct {
	APIError
}

// NotFoundError is returned when a requested resource is not found (HTTP 404).
type NotFoundError struct {
	APIError
}
