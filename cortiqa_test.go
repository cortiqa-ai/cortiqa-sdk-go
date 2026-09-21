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
