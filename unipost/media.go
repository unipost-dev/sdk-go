package unipost

import (
	"context"
	"encoding/json"
)

// MediaService handles media upload operations.
type MediaService struct {
	http *httpClient
}

// Upload requests a presigned upload URL.
func (s *MediaService) Upload(ctx context.Context, req *MediaUploadRequest) (*MediaUploadResponse, error) {
	data, err := s.http.post(ctx, "/v1/media/upload", req, nil)
	if err != nil {
		return nil, err
	}

	var resp dataEnvelope[MediaUploadResponse]
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}
	return &resp.Data, nil
}
