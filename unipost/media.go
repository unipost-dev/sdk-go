package unipost

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
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
	client         *Client
	AudioOverlays  *AudioOverlaysService
	GIFConversions *GIFConversionsService
}

// UploadFile reserves Media storage and uploads a local file with a presigned PUT.
func (s *MediaService) UploadFile(ctx context.Context, filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("unipost: open media file: %w", err)
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		return "", fmt.Errorf("unipost: stat media file: %w", err)
	}
	contentType := mediaContentType(filepath.Ext(filePath))
	reserved, err := s.Upload(ctx, &MediaUploadRequest{
		Filename: filepath.Base(filePath), ContentType: contentType, SizeBytes: info.Size(),
	})
	if err != nil {
		return "", err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, reserved.UploadURL, file)
	if err != nil {
		return "", fmt.Errorf("unipost: create media upload request: %w", err)
	}
	req.Header.Set("Content-Type", contentType)
	resp, err := s.client.http.Do(req)
	if err != nil {
		return "", fmt.Errorf("unipost: media upload failed: %w", err)
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("unipost: media upload failed with status %d", resp.StatusCode)
	}
	return reserved.MediaID, nil
}

func mediaContentType(ext string) string {
	switch strings.ToLower(ext) {
	case ".gif":
		return "image/gif"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".mp4":
		return "video/mp4"
	default:
		return "application/octet-stream"
	}
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

// GIFConversionCreateRequest configures a GIF-to-MP4 processing job.
type GIFConversionCreateRequest struct {
	GIFMediaID      string
	BackgroundColor string
}

// GIFConversionJobError describes a failed GIF conversion job.
type GIFConversionJobError struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	Retryable bool   `json:"retryable"`
}

// GIFConversionJob is returned by the GIF conversion API.
type GIFConversionJob struct {
	ID              string                 `json:"id"`
	Kind            string                 `json:"kind"`
	Status          string                 `json:"status"`
	GIFMediaID      string                 `json:"gif_media_id"`
	BackgroundColor string                 `json:"background_color"`
	OutputProfile   string                 `json:"output_profile"`
	OutputMediaID   *string                `json:"output_media_id"`
	CreatedAt       time.Time              `json:"created_at"`
	StartedAt       *time.Time             `json:"started_at"`
	CompletedAt     *time.Time             `json:"completed_at"`
	Error           *GIFConversionJobError `json:"error"`
}

// GIFConversionError is returned when a conversion reaches terminal failure.
type GIFConversionError struct {
	Code      string
	Message   string
	Retryable bool
}

func (e *GIFConversionError) Error() string {
	return e.Message
}

// GIFConversionWaitOptions configures client-side polling only.
type GIFConversionWaitOptions struct {
	PollInterval time.Duration
	Timeout      time.Duration
}

// GIFUploadAndConvertOptions configures upload, conversion, and polling.
type GIFUploadAndConvertOptions struct {
	BackgroundColor string
	IdempotencyKey  string
	PollInterval    time.Duration
	Timeout         time.Duration
}

// GIFConversionsService handles GIF-to-MP4 jobs.
type GIFConversionsService struct {
	client *Client
	media  *MediaService
}

func (s *GIFConversionsService) Create(ctx context.Context, params *GIFConversionCreateRequest, opts ...RequestOption) (*GIFConversionJob, error) {
	body := map[string]any{"gif_media_id": params.GIFMediaID}
	if params.BackgroundColor != "" {
		body["background_color"] = params.BackgroundColor
	}
	options := collectRequestOptions(opts)
	headers := map[string]string{}
	if options.idempotencyKey != "" {
		headers["Idempotency-Key"] = options.idempotencyKey
	}
	var env apiEnvelope[GIFConversionJob]
	if err := s.client.do(ctx, http.MethodPost, "/v1/media/gif-conversions", nil, body, &env, headers); err != nil {
		return nil, err
	}
	return &env.Data, nil
}

func (s *GIFConversionsService) Get(ctx context.Context, conversionID string) (*GIFConversionJob, error) {
	var env apiEnvelope[GIFConversionJob]
	if err := s.client.do(ctx, http.MethodGet, "/v1/media/gif-conversions/"+conversionID, nil, nil, &env, nil); err != nil {
		return nil, err
	}
	return &env.Data, nil
}

func (s *GIFConversionsService) Wait(ctx context.Context, conversionID string, options *GIFConversionWaitOptions) (*GIFConversionJob, error) {
	pollInterval := 2 * time.Second
	timeout := 5 * time.Minute
	if options != nil {
		if options.PollInterval > 0 {
			pollInterval = options.PollInterval
		}
		if options.Timeout > 0 {
			timeout = options.Timeout
		}
	}
	waitCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	for {
		job, err := s.Get(waitCtx, conversionID)
		if err != nil {
			return nil, err
		}
		switch job.Status {
		case "succeeded":
			return job, nil
		case "failed":
			jobErr := job.Error
			if jobErr == nil {
				jobErr = &GIFConversionJobError{Code: "gif_conversion_failed", Message: "GIF conversion failed"}
			}
			return nil, &GIFConversionError{Code: jobErr.Code, Message: jobErr.Message, Retryable: jobErr.Retryable}
		}
		timer := time.NewTimer(pollInterval)
		select {
		case <-waitCtx.Done():
			timer.Stop()
			return nil, waitCtx.Err()
		case <-timer.C:
		}
	}
}

func (s *GIFConversionsService) UploadAndConvert(ctx context.Context, filePath string, options *GIFUploadAndConvertOptions) (*GIFConversionJob, error) {
	if options == nil {
		options = &GIFUploadAndConvertOptions{}
	}
	mediaID, err := s.media.UploadFile(ctx, filePath)
	if err != nil {
		return nil, err
	}
	idempotencyKey := options.IdempotencyKey
	if idempotencyKey == "" {
		idempotencyKey, err = randomIdempotencyKey()
		if err != nil {
			return nil, err
		}
	}
	created, err := s.Create(ctx, &GIFConversionCreateRequest{
		GIFMediaID: mediaID, BackgroundColor: options.BackgroundColor,
	}, WithIdempotencyKey(idempotencyKey))
	if err != nil {
		return nil, err
	}
	return s.Wait(ctx, created.ID, &GIFConversionWaitOptions{
		PollInterval: options.PollInterval, Timeout: options.Timeout,
	})
}

func randomIdempotencyKey() (string, error) {
	raw := make([]byte, 16)
	if _, err := rand.Read(raw); err != nil {
		return "", fmt.Errorf("unipost: generate idempotency key: %w", err)
	}
	return hex.EncodeToString(raw), nil
}

func (s *MediaService) CreateGIFConversion(ctx context.Context, params *GIFConversionCreateRequest, opts ...RequestOption) (*GIFConversionJob, error) {
	return s.GIFConversions.Create(ctx, params, opts...)
}

func (s *MediaService) GetGIFConversion(ctx context.Context, conversionID string) (*GIFConversionJob, error) {
	return s.GIFConversions.Get(ctx, conversionID)
}

func (s *MediaService) WaitForGIFConversion(ctx context.Context, conversionID string, options *GIFConversionWaitOptions) (*GIFConversionJob, error) {
	return s.GIFConversions.Wait(ctx, conversionID, options)
}

func (s *MediaService) UploadAndConvertGIF(ctx context.Context, filePath string, options *GIFUploadAndConvertOptions) (*GIFConversionJob, error) {
	return s.GIFConversions.UploadAndConvert(ctx, filePath, options)
}
