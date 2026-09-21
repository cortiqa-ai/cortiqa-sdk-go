package cortiqa

import (
	"context"
	"net/http"
)

// ModelsService handles AI model discovery.
type ModelsService struct {
	client *Client
}

// List returns all available Cortiqa AI models.
func (s *ModelsService) List(ctx context.Context) ([]ModelInfo, error) {
	var resp ModelListResponse
	err := s.client.sendRequest(ctx, http.MethodGet, "/api/v1/ai/models", nil, &resp)
	if err != nil {
		return nil, err
	}
	return resp.Data, nil
}
