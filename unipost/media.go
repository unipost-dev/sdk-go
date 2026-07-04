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
	client        *Client
	AudioOverlays *AudioOverlaysService
}

func (s *MediaService) Upload(ctx context.Context, params *MediaUploadRequest) (*MediaUploadResponse, error) {
	body := map[string]any{
		"filename":     params.Filename,
		"content_type": params.ContentType,
	}
	if params.SizeBytes > 0 {
		body["size_bytes"] = params.SizeBytes
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

// RequestOption configures request-scoped behavior.
type RequestOption func(*requestOptions)

type requestOptions struct {
	idempotencyKey string
}

// WithIdempotencyKey sends the Idempotency-Key header for endpoints that support it.
func WithIdempotencyKey(key string) RequestOption {
	return func(opts *requestOptions) {
		opts.idempotencyKey = key
	}
}

func collectRequestOptions(opts []RequestOption) requestOptions {
	var out requestOptions
	for _, opt := range opts {
		if opt != nil {
			opt(&out)
		}
	}
	return out
}

// AudioOverlayCreateRequest configures an audio overlay processing job.
type AudioOverlayCreateRequest struct {
	VideoMediaID string
	AudioMediaID string
	Mode         string
	VideoVolume  *int32
	AudioVolume  *int32
	AudioStartMs *int32
	Fit          string
}

// AudioOverlayError describes a failed audio overlay job.
type AudioOverlayError struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	Retryable bool   `json:"retryable"`
}

// AudioOverlayJob is the processing job returned by the audio overlay API.
type AudioOverlayJob struct {
	ID            string             `json:"id"`
	Status        string             `json:"status"`
	VideoMediaID  string             `json:"video_media_id"`
	AudioMediaID  string             `json:"audio_media_id"`
	OutputMediaID *string            `json:"output_media_id"`
	Mode          string             `json:"mode"`
	Fit           string             `json:"fit"`
	CreatedAt     time.Time          `json:"created_at"`
	StartedAt     *time.Time         `json:"started_at"`
	CompletedAt   *time.Time         `json:"completed_at"`
	Error         *AudioOverlayError `json:"error"`
}

// AudioOverlaysService handles media audio overlay processing jobs.
type AudioOverlaysService struct {
	client *Client
}

func (s *AudioOverlaysService) Create(ctx context.Context, params *AudioOverlayCreateRequest, opts ...RequestOption) (*AudioOverlayJob, error) {
	body := map[string]any{
		"video_media_id": params.VideoMediaID,
		"audio_media_id": params.AudioMediaID,
	}
	if params.Mode != "" {
		body["mode"] = params.Mode
	}
	if params.VideoVolume != nil {
		body["video_volume"] = *params.VideoVolume
	}
	if params.AudioVolume != nil {
		body["audio_volume"] = *params.AudioVolume
	}
	if params.AudioStartMs != nil {
		body["audio_start_ms"] = *params.AudioStartMs
	}
	if params.Fit != "" {
		body["fit"] = params.Fit
	}

	headers := map[string]string{}
	options := collectRequestOptions(opts)
	if options.idempotencyKey != "" {
		headers["Idempotency-Key"] = options.idempotencyKey
	}

	var env apiEnvelope[AudioOverlayJob]
	if err := s.client.do(ctx, http.MethodPost, "/v1/media/audio-overlays", nil, body, &env, headers); err != nil {
		return nil, err
	}
	return &env.Data, nil
}

func (s *AudioOverlaysService) Get(ctx context.Context, jobID string) (*AudioOverlayJob, error) {
	var env apiEnvelope[AudioOverlayJob]
	if err := s.client.do(ctx, http.MethodGet, "/v1/media/audio-overlays/"+jobID, nil, nil, &env, nil); err != nil {
		return nil, err
	}
	return &env.Data, nil
}
