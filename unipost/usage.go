package unipost

import (
	"context"
	"net/http"
)

// Usage describes the workspace's monthly post usage.
type Usage struct {
	Period     string  `json:"period"`
	PostCount  int     `json:"post_count"`
	PostLimit  int     `json:"post_limit"`
	Plan       string  `json:"plan"`
	Percentage float64 `json:"percentage"`
	Warning    string  `json:"warning,omitempty"`
}

// UsageService handles usage-meter operations.
type UsageService struct {
	client *Client
}

func (s *UsageService) Get(ctx context.Context) (*Usage, error) {
	var env apiEnvelope[Usage]
	if err := s.client.do(ctx, http.MethodGet, "/v1/usage", nil, nil, &env, nil); err != nil {
		return nil, err
	}
	return &env.Data, nil
}
