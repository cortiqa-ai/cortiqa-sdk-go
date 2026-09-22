package cortiqa

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewClientOptions(t *testing.T) {
	client := NewClient(
		"sk-cortiqa-test",
		WithBaseURL("https://custom.api.cortiqa.co/"),
		WithTimeout(15*time.Second),
		WithMaxRetries(4),
	)

	if client.apiKey != "sk-cortiqa-test" {
		t.Fatalf("expected apiKey 'sk-cortiqa-test', got '%s'", client.apiKey)
	}
	if client.baseURL != "https://custom.api.cortiqa.co" {
		t.Fatalf("expected baseURL 'https://custom.api.cortiqa.co', got '%s'", client.baseURL)
	}
	if client.httpClient.Timeout != 15*time.Second {
		t.Fatalf("expected timeout 15s, got %v", client.httpClient.Timeout)
	}
	if client.maxRetries != 4 {
		t.Fatalf("expected maxRetries 4, got %d", client.maxRetries)
	}
}

func TestChatCompletionMock(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/chat/completions" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer sk-test-key" {
			t.Fatalf("unexpected auth header: %s", r.Header.Get("Authorization"))
		}

		resp := ChatCompletionResponse{
			ID:      "chatcmpl-123",
			Object:  "chat.completion",
			Created: 1726900000,
			Model:   "falin-01",
			Choices: []ChatCompletionChoice{
				{
					Index: 0,
					Message: ChatMessage{
						Role:    "assistant",
						Content: "Hello from Cortiqa Go SDK!",
					},
					FinishReason: "stop",
				},
			},
			Usage: Usage{
				PromptTokens:     10,
				CompletionTokens: 5,
				TotalTokens:      15,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient("sk-test-key", WithBaseURL(server.URL))
	ctx := context.Background()

	resp, err := client.Chat.Create(ctx, ChatCompletionRequest{
		Model: "falin-01",
		Messages: []ChatMessage{
			{Role: "user", Content: "Hello!"},
		},
	})

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if resp.Content() != "Hello from Cortiqa Go SDK!" {
		t.Fatalf("unexpected content: %s", resp.Content())
	}
	if resp.Usage.TotalTokens != 15 {
		t.Fatalf("unexpected total tokens: %d", resp.Usage.TotalTokens)
	}
}

func TestAuthenticationError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Authorization header required.",
		})
	}))
	defer server.Close()

	client := NewClient("invalid-key", WithBaseURL(server.URL), WithMaxRetries(0))
	ctx := context.Background()

	_, err := client.Chat.Create(ctx, ChatCompletionRequest{
		Model:    "falin-01",
		Messages: []ChatMessage{{Role: "user", Content: "Hi"}},
	})

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if _, ok := err.(*AuthenticationError); !ok {
		t.Fatalf("expected AuthenticationError, got: %T", err)
	}
}

func TestDefaultModelAndPrompt(t *testing.T) {
	var capturedReq map[string]interface{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&capturedReq)
		resp := ChatCompletionResponse{
			ID:      "chatcmpl-prompt",
			Object:  "chat.completion",
			Created: 123456,
			Model:   DefaultModel,
			Choices: []ChatCompletionChoice{
				{
					Index: 0,
					Message: ChatMessage{
						Role:    "assistant",
						Content: "One-liner answer!",
					},
					FinishReason: "stop",
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient("sk-test", WithBaseURL(server.URL))
	if client.DefaultModel() != "openai/gpt-oss-120b" {
		t.Fatalf("expected DefaultModel 'openai/gpt-oss-120b', got '%s'", client.DefaultModel())
	}

	ans, err := client.Prompt(context.Background(), "Hello test")
	if err != nil {
		t.Fatalf("unexpected prompt error: %v", err)
	}
	if ans != "One-liner answer!" {
		t.Fatalf("unexpected prompt answer: %s", ans)
	}
	if capturedReq["model"] != "openai/gpt-oss-120b" {
		t.Fatalf("expected request model 'openai/gpt-oss-120b', got '%v'", capturedReq["model"])
	}
}

func TestReasoningAndToolNormalization(t *testing.T) {
	var capturedReq map[string]interface{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&capturedReq)
		resp := ChatCompletionResponse{
			ID:      "chatcmpl-reasoning",
			Object:  "chat.completion",
			Created: 123456,
			Model:   DefaultModel,
			Choices: []ChatCompletionChoice{
				{
					Index: 0,
					Message: ChatMessage{
						Role:      "assistant",
						Content:   "Result: 42",
						Reasoning: "Calculated 6 * 7 = 42",
					},
					FinishReason: "stop",
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient("sk-test", WithBaseURL(server.URL))
	resp, err := client.Chat.Create(context.Background(), ChatCompletionRequest{
		Messages: []ChatMessage{{Role: "user", Content: "Calculate"}},
		Tools: []Tool{
			{Function: FunctionDefinition{Name: "calc"}}, // Type is omitted; should default to "function"
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.Reasoning() != "Calculated 6 * 7 = 42" {
		t.Fatalf("expected reasoning 'Calculated 6 * 7 = 42', got '%s'", resp.Reasoning())
	}

	// Verify tools normalized to type="function"
	toolsArr, ok := capturedReq["tools"].([]interface{})
	if !ok || len(toolsArr) != 1 {
		t.Fatalf("expected tools array of length 1, got %v", capturedReq["tools"])
	}
	toolMap := toolsArr[0].(map[string]interface{})
	if toolMap["type"] != "function" {
		t.Fatalf("expected tool type 'function', got '%v'", toolMap["type"])
	}
}

func TestErrorDetailsParsing(t *testing.T) {
	// 1. HTTP 400 Bad Request
	badReqServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"error": map[string]interface{}{
				"message": "Invalid temperature value",
				"param":   "temperature",
				"code":    "invalid_parameter",
				"type":    "invalid_request_error",
			},
		})
	}))
	defer badReqServer.Close()

	client := NewClient("sk-test", WithBaseURL(badReqServer.URL), WithMaxRetries(0))
	_, err := client.Chat.Create(context.Background(), ChatCompletionRequest{
		Messages: []ChatMessage{{Role: "user", Content: "Hi"}},
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	badReqErr, ok := err.(*BadRequestError)
	if !ok {
		t.Fatalf("expected BadRequestError, got: %T", err)
	}
	if badReqErr.Param != "temperature" {
		t.Fatalf("expected param 'temperature', got '%s'", badReqErr.Param)
	}
	if badReqErr.Code != "invalid_parameter" {
		t.Fatalf("expected code 'invalid_parameter', got '%s'", badReqErr.Code)
	}

	// 2. HTTP 422 Unprocessable Entity with FastAPI detail array
	unprocessableServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"detail": []interface{}{
				map[string]interface{}{
					"loc":  []interface{}{"body", "messages"},
					"msg":  "field required",
					"type": "value_error.missing",
				},
			},
		})
	}))
	defer unprocessableServer.Close()

	client2 := NewClient("sk-test", WithBaseURL(unprocessableServer.URL), WithMaxRetries(0))
	_, err2 := client2.Chat.Create(context.Background(), ChatCompletionRequest{})
	if err2 == nil {
		t.Fatal("expected error, got nil")
	}
	unprocErr, ok := err2.(*UnprocessableEntityError)
	if !ok {
		t.Fatalf("expected UnprocessableEntityError, got: %T", err2)
	}
	if unprocErr.Param != "messages" {
		t.Fatalf("expected param 'messages', got '%s'", unprocErr.Param)
	}
}
