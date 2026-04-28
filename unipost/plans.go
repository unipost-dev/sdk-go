package unipost

import (
	"context"
	"net/http"
)

// Plan represents a UniPost subscription plan.
type Plan struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	PriceCents int32  `json:"price_cents"`
	PostLimit  int32  `json:"post_limit"`
}

// PlansService handles plan operations.
type PlansService struct {
	client *Client
}

// List returns available subscription plans.
func (s *PlansService) List(ctx context.Context) ([]Plan, error) {
	var env apiEnvelope[[]Plan]
	if err := s.client.do(ctx, http.MethodGet, "/v1/plans", nil, nil, &env, nil); err != nil {
		return nil, err
	}
	return env.Data, nil
}
