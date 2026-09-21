package cortiqa

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"time"
)

func (c *Client) sendRequest(ctx context.Context, method, path string, body interface{}, out interface{}) error {
	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("cortiqa: failed to encode request body: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	url := c.baseURL + path

	var resp *http.Response
	var err error

	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		// Re-create reader for retries if body exists
		if body != nil && attempt > 0 {
			data, _ := json.Marshal(body)
			bodyReader = bytes.NewReader(data)
		}

		req, reqErr := http.NewRequestWithContext(ctx, method, url, bodyReader)
		if reqErr != nil {
			return fmt.Errorf("cortiqa: failed to create request: %w", reqErr)
		}

		req.Header.Set("Authorization", "Bearer "+c.apiKey)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("User-Agent", "cortiqa-go/"+Version)

		resp, err = c.httpClient.Do(req)
		if err != nil {
			if attempt < c.maxRetries {
				sleepDuration := time.Duration(math.Pow(2, float64(attempt))*500) * time.Millisecond
				select {
				case <-time.After(sleepDuration):
					continue
				case <-ctx.Done():
					return ctx.Err()
				}
			}
			return fmt.Errorf("cortiqa: network error: %w", err)
		}

		// Check for retryable HTTP status codes (429, 500, 502, 503, 504)
		if (resp.StatusCode == 429 || resp.StatusCode >= 500) && attempt < c.maxRetries {
			resp.Body.Close()
			sleepDuration := time.Duration(math.Pow(2, float64(attempt))*500) * time.Millisecond
			select {
			case <-time.After(sleepDuration):
				continue
			case <-ctx.Done():
				return ctx.Err()
			}
		}

		break
	}

	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("cortiqa: failed to read response body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return c.parseError(resp.StatusCode, respBytes)
	}

	if out != nil {
		if err := json.Unmarshal(respBytes, out); err != nil {
			return fmt.Errorf("cortiqa: failed to decode response JSON: %w", err)
		}
	}

	return nil
}

func (c *Client) parseError(statusCode int, body []byte) error {
	var errResp struct {
		Success bool `json:"success"`
		Error   interface{} `json:"error"`
		Message string      `json:"message"`
	}

	errMsg := string(body)
	if err := json.Unmarshal(body, &errResp); err == nil {
		if s, ok := errResp.Error.(string); ok && s != "" {
			errMsg = s
		} else if m, ok := errResp.Error.(map[string]interface{}); ok {
			if msg, exists := m["message"].(string); exists {
				errMsg = msg
			}
		} else if errResp.Message != "" {
			errMsg = errResp.Message
		}
	}

	baseErr := APIError{
		StatusCode: statusCode,
		Message:    errMsg,
	}

	switch statusCode {
	case http.StatusUnauthorized:
		return &AuthenticationError{APIError: baseErr}
	case http.StatusTooManyRequests:
		return &RateLimitError{APIError: baseErr}
	case http.StatusNotFound:
		return &NotFoundError{APIError: baseErr}
	default:
		return &baseErr
	}
}
