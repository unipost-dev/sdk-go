package unipost

import (
	"context"
	"net/http"
	"time"
)

// MediaUploadRequest configures a media upload session.
type MediaUploadRequest struct {
	Filename    string
	ContentType string
	SizeBytes   int64
	ContentHash string
}

// MediaUploadResponse is the result of requesting a presigned upload URL.
type MediaUploadResponse struct {
	ID          string    `json:"id,omitempty"`
	MediaID     string    `json:"media_id,omitempty"`
	Status      string    `json:"status"`
	ContentType string    `json:"content_type"`
	SizeBytes   int64     `json:"size_bytes"`
	UploadURL   string    `json:"upload_url,omitempty"`
	DownloadURL string    `json:"download_url,omitempty"`
	ExpiresAt   time.Time `json:"expires_at,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

// MediaService handles media upload operations.
type MediaService struct {
	client *Client
}

func (s *MediaService) Upload(ctx context.Context, params *MediaUploadRequest) (*MediaUploadResponse, error) {
	body := map[string]any{
		"filename":     params.Filename,
		"content_type": params.ContentType,
		"size_bytes":   params.SizeBytes,
	}
	if params.ContentHash != "" {
		body["content_hash"] = params.ContentHash
	}
	var env apiEnvelope[MediaUploadResponse]
	if err := s.client.do(ctx, http.MethodPost, "/v1/media", nil, body, &env, nil); err != nil {
		return nil, err
	}
	if env.Data.MediaID == "" {
		env.Data.MediaID = env.Data.ID
	}
	return &env.Data, nil
}

func (s *MediaService) Get(ctx context.Context, mediaID string) (*MediaUploadResponse, error) {
	var env apiEnvelope[MediaUploadResponse]
	if err := s.client.do(ctx, http.MethodGet, "/v1/media/"+mediaID, nil, nil, &env, nil); err != nil {
		return nil, err
	}
	if env.Data.MediaID == "" {
		env.Data.MediaID = env.Data.ID
	}
	return &env.Data, nil
}

func (s *MediaService) Delete(ctx context.Context, mediaID string) error {
	return s.client.do(ctx, http.MethodDelete, "/v1/media/"+mediaID, nil, nil, nil, nil)
}
