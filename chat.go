package cortiqa

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// ChatService handles OpenAI-compatible chat completions.
type ChatService struct {
	client *Client
}

// MessagesService handles Anthropic-compatible message completions.
type MessagesService struct {
	client *Client
}

func (s *ChatService) normalizeRequest(req *ChatCompletionRequest) {
	if req.Model == "" {
		req.Model = s.client.defaultModel
	}
	for i := range req.Tools {
		if req.Tools[i].Type == "" {
			req.Tools[i].Type = "function"
		}
	}
}

// Create sends a non-streaming chat completion request.
func (s *ChatService) Create(ctx context.Context, req ChatCompletionRequest) (*ChatCompletionResponse, error) {
	req.Stream = false
	s.normalizeRequest(&req)
	var resp ChatCompletionResponse
	err := s.client.sendRequest(ctx, http.MethodPost, "/v1/chat/completions", req, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// Create sends an Anthropic-style message completion request.
func (s *MessagesService) Create(ctx context.Context, req ChatCompletionRequest) (*ChatCompletionResponse, error) {
	return s.client.Chat.Create(ctx, req)
}

// ChatStream represents an active token stream.
type ChatStream struct {
	reader   *bufio.Reader
	response *http.Response
}

// Recv reads the next chunk from the stream.
// Returns io.EOF when stream is finished.
func (s *ChatStream) Recv() (*ChatCompletionStreamResponse, error) {
	for {
		line, err := s.reader.ReadString('\n')
		if err != nil {
			return nil, err
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "data: ") {
			data := strings.TrimPrefix(line, "data: ")
			if data == "[DONE]" {
				return nil, io.EOF
			}

			var chunk ChatCompletionStreamResponse
			if err := json.Unmarshal([]byte(data), &chunk); err != nil {
				continue
			}
			return &chunk, nil
		}
	}
}

// Close closes the underlying stream connection.
func (s *ChatStream) Close() error {
	return s.response.Body.Close()
}

// CreateStream opens an SSE stream for real-time tokens.
func (s *ChatService) CreateStream(ctx context.Context, req ChatCompletionRequest) (*ChatStream, error) {
	req.Stream = true
	s.normalizeRequest(&req)
	bodyData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("cortiqa: failed to encode stream payload: %w", err)
	}

	url := s.client.baseURL + "/v1/chat/completions"
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(bodyData))
	if err != nil {
		return nil, fmt.Errorf("cortiqa: failed to create stream request: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+s.client.apiKey)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("User-Agent", "cortiqa-go/"+Version)

	resp, err := s.client.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("cortiqa: failed to connect to stream: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		defer resp.Body.Close()
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, s.client.parseError(resp.StatusCode, bodyBytes)
	}

	return &ChatStream{
		reader:   bufio.NewReader(resp.Body),
		response: resp,
	}, nil
}

// CreateStream opens an SSE stream for Anthropic-style messages.
func (s *MessagesService) CreateStream(ctx context.Context, req ChatCompletionRequest) (*ChatStream, error) {
	return s.client.Chat.CreateStream(ctx, req)
}
