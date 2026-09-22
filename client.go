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
	var raw map[string]interface{}
	errMsg := string(body)
	var param, code, errorType string

	if err := json.Unmarshal(body, &raw); err == nil {
		// 1. Check FastAPI / Pydantic detail array
		if detailRaw, ok := raw["detail"]; ok {
			if detailArr, isArr := detailRaw.([]interface{}); isArr && len(detailArr) > 0 {
				if firstMap, isMap := detailArr[0].(map[string]interface{}); isMap {
					if locArr, hasLoc := firstMap["loc"].([]interface{}); hasLoc && len(locArr) > 0 {
						param = fmt.Sprintf("%v", locArr[len(locArr)-1])
					}
					if t, hasT := firstMap["type"].(string); hasT {
						code = t
					}
					if m, hasM := firstMap["msg"].(string); hasM {
						if param != "" {
							errMsg = fmt.Sprintf("Parameter '%s': %s", param, m)
						} else {
							errMsg = m
						}
					}
				}
			} else if detailStr, isStr := detailRaw.(string); isStr {
				errMsg = detailStr
			}
		} else if errObj, hasErr := raw["error"]; hasErr {
			// 2. Check OpenAI standard error object
			if m, ok := errObj.(map[string]interface{}); ok {
				if msg, exists := m["message"].(string); exists {
					errMsg = msg
				}
				if p, exists := m["param"].(string); exists {
					param = p
				}
				if cStr, exists := m["code"].(string); exists {
					code = cStr
				}
				if t, exists := m["type"].(string); exists {
					errorType = t
				}
			} else if s, ok := errObj.(string); ok && s != "" {
				errMsg = s
			}
		} else if msg, exists := raw["message"].(string); exists && msg != "" {
			errMsg = msg
		}
	}

	baseErr := APIError{
		StatusCode: statusCode,
		Message:    errMsg,
		Param:      param,
		Code:       code,
		ErrorType:  errorType,
	}

	switch statusCode {
	case http.StatusBadRequest:
		return &BadRequestError{APIError: baseErr}
	case http.StatusUnauthorized:
		return &AuthenticationError{APIError: baseErr}
	case http.StatusForbidden:
		return &PermissionDeniedError{APIError: baseErr}
	case http.StatusNotFound:
		return &NotFoundError{APIError: baseErr}
	case http.StatusUnprocessableEntity:
		return &UnprocessableEntityError{APIError: baseErr}
	case http.StatusTooManyRequests:
		return &RateLimitError{APIError: baseErr}
	case http.StatusInternalServerError, http.StatusBadGateway, http.StatusServiceUnavailable, http.StatusGatewayTimeout:
		return &InternalServerError{APIError: baseErr}
	default:
		return &baseErr
	}
}
