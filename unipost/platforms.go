package unipost

import (
	"context"
	"net/http"
)

// PlatformsService handles platform metadata.
type PlatformsService struct {
	client *Client
}

// Capabilities returns the per-platform capability matrix.
func (s *PlatformsService) Capabilities(ctx context.Context) (JSONMap, error) {
	var env apiEnvelope[JSONMap]
	if err := s.client.do(ctx, http.MethodGet, "/v1/platforms/capabilities", nil, nil, &env, nil); err != nil {
		return nil, err
	}
	return env.Data, nil
}
