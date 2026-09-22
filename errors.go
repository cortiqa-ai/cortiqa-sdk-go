package cortiqa

import (
	"fmt"
	"strings"
)

// APIError represents an error returned by the Cortiqa API.
type APIError struct {
	StatusCode int    `json:"status_code"`
	Message    string `json:"message"`
	Param      string `json:"param,omitempty"`
	Code       string `json:"code,omitempty"`
	ErrorType  string `json:"type,omitempty"`
}

func (e *APIError) Error() string {
	msg := e.Message
	if e.Param != "" && !strings.Contains(msg, e.Param) {
		msg = fmt.Sprintf("[%s] %s", e.Param, msg)
	}
	if e.Code != "" {
		return fmt.Sprintf("cortiqa: API error (status %d, code %s): %s", e.StatusCode, e.Code, msg)
	}
	return fmt.Sprintf("cortiqa: API error (status %d): %s", e.StatusCode, msg)
}

// BadRequestError is returned when request parameters are invalid (HTTP 400).
type BadRequestError struct {
	APIError
}

// AuthenticationError is returned when an invalid or missing API key is used (HTTP 401).
type AuthenticationError struct {
	APIError
}

// PermissionDeniedError is returned when access is forbidden (HTTP 403).
type PermissionDeniedError struct {
	APIError
}

// NotFoundError is returned when a requested resource is not found (HTTP 404).
type NotFoundError struct {
	APIError
}

// UnprocessableEntityError is returned when request validation fails (HTTP 422).
type UnprocessableEntityError struct {
	APIError
}

// RateLimitError is returned when rate limits are exceeded (HTTP 429).
type RateLimitError struct {
	APIError
}

// InternalServerError is returned on server errors (HTTP 500+).
type InternalServerError struct {
	APIError
}
