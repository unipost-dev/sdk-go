package unipost

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"
)

func TestMediaUploadOmitsOptionalSizeBytes(t *testing.T) {
	var body map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/media" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"id":"media_audio_1","status":"reserved","upload_url":"https://upload.example/audio"}}`))
	}))
	defer server.Close()

	client := NewClient(WithAPIKey("up_test_xxx"), WithBaseURL(server.URL))
	result, err := client.Media.Upload(context.Background(), &MediaUploadRequest{
		Filename:    "voiceover.mp3",
		ContentType: "audio/mpeg",
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.MediaID != "media_audio_1" {
		t.Fatalf("unexpected media id %q", result.MediaID)
	}
	if _, ok := body["size_bytes"]; ok {
		t.Fatalf("expected size_bytes to be omitted, got %#v", body)
	}
	if body["filename"] != "voiceover.mp3" || body["content_type"] != "audio/mpeg" {
		t.Fatalf("unexpected body %#v", body)
	}
}

func TestGIFConversionCreateAndWait(t *testing.T) {
	var gets atomic.Int32
	var createBody map[string]any
	var idempotency string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/v1/media/gif-conversions":
			idempotency = r.Header.Get("Idempotency-Key")
			_ = json.NewDecoder(r.Body).Decode(&createBody)
			w.WriteHeader(http.StatusAccepted)
			_, _ = w.Write([]byte(`{"data":{"id":"mpj_gif_1","kind":"gif_to_mp4","status":"queued","gif_media_id":"media_gif_1","background_color":"#FFFFFF","output_profile":"universal_mp4_v1","output_media_id":null,"created_at":"2026-07-17T12:00:00Z","error":null}}`))
		case r.Method == http.MethodGet && r.URL.Path == "/v1/media/gif-conversions/mpj_gif_1":
			if gets.Add(1) == 1 {
				_, _ = w.Write([]byte(`{"data":{"id":"mpj_gif_1","kind":"gif_to_mp4","status":"processing","gif_media_id":"media_gif_1","background_color":"#FFFFFF","output_profile":"universal_mp4_v1","output_media_id":null,"created_at":"2026-07-17T12:00:00Z","error":null}}`))
				return
			}
			_, _ = w.Write([]byte(`{"data":{"id":"mpj_gif_1","kind":"gif_to_mp4","status":"succeeded","gif_media_id":"media_gif_1","background_color":"#FFFFFF","output_profile":"universal_mp4_v1","output_media_id":"media_mp4_1","created_at":"2026-07-17T12:00:00Z","error":null}}`))
		default:
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
	}))
	defer server.Close()

	client := NewClient(WithAPIKey("up_test_xxx"), WithBaseURL(server.URL))
	created, err := client.Media.GIFConversions.Create(context.Background(), &GIFConversionCreateRequest{
		GIFMediaID: "media_gif_1", BackgroundColor: "#FFFFFF",
	}, WithIdempotencyKey("gif-1"))
	if err != nil {
		t.Fatal(err)
	}
	job, err := client.Media.GIFConversions.Wait(context.Background(), created.ID, &GIFConversionWaitOptions{
		PollInterval: time.Millisecond, Timeout: time.Second,
	})
	if err != nil {
		t.Fatal(err)
	}
	if job.OutputMediaID == nil || *job.OutputMediaID != "media_mp4_1" {
		t.Fatalf("unexpected job %#v", job)
	}
	if idempotency != "gif-1" || createBody["gif_media_id"] != "media_gif_1" {
		t.Fatalf("unexpected create request %#v key=%q", createBody, idempotency)
	}
}

func TestGIFConversionWaitFailureTimeoutAndCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"id":"mpj_gif_1","kind":"gif_to_mp4","status":"failed","gif_media_id":"media_gif_1","background_color":"#FFFFFF","output_profile":"universal_mp4_v1","output_media_id":null,"created_at":"2026-07-17T12:00:00Z","error":{"code":"gif_decode_failed","message":"bad gif","retryable":false}}}`))
	}))
	defer server.Close()
	client := NewClient(WithAPIKey("up_test_xxx"), WithBaseURL(server.URL))
	_, err := client.Media.GIFConversions.Wait(context.Background(), "mpj_gif_1", nil)
	var conversionErr *GIFConversionError
	if !errors.As(err, &conversionErr) || conversionErr.Code != "gif_decode_failed" || conversionErr.Retryable {
		t.Fatalf("unexpected error %#v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err = client.Media.GIFConversions.Wait(ctx, "mpj_gif_1", nil)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context cancellation, got %v", err)
	}

	processing := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"id":"mpj_gif_1","kind":"gif_to_mp4","status":"processing","gif_media_id":"media_gif_1","background_color":"#FFFFFF","output_profile":"universal_mp4_v1","output_media_id":null,"created_at":"2026-07-17T12:00:00Z","error":null}}`))
	}))
	defer processing.Close()
	timedClient := NewClient(WithAPIKey("up_test_xxx"), WithBaseURL(processing.URL))
	_, err = timedClient.Media.GIFConversions.Wait(context.Background(), "mpj_gif_1", &GIFConversionWaitOptions{
		PollInterval: time.Second, Timeout: time.Millisecond,
	})
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected polling timeout, got %v", err)
	}
}

func TestUploadAndConvertGIFDoesNotPublish(t *testing.T) {
	var server *httptest.Server
	var sawPost bool
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/v1/posts" {
			sawPost = true
		}
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/v1/media":
			_, _ = w.Write([]byte(`{"data":{"id":"media_gif_1","status":"reserved","upload_url":"` + server.URL + `/upload"}}`))
		case r.Method == http.MethodPut && r.URL.Path == "/upload":
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodPost && r.URL.Path == "/v1/media/gif-conversions":
			w.WriteHeader(http.StatusAccepted)
			_, _ = w.Write([]byte(`{"data":{"id":"mpj_gif_1","kind":"gif_to_mp4","status":"queued","gif_media_id":"media_gif_1","background_color":"#FFFFFF","output_profile":"universal_mp4_v1","output_media_id":null,"created_at":"2026-07-17T12:00:00Z"}}`))
		case r.Method == http.MethodGet && r.URL.Path == "/v1/media/gif-conversions/mpj_gif_1":
			_, _ = w.Write([]byte(`{"data":{"id":"mpj_gif_1","kind":"gif_to_mp4","status":"succeeded","gif_media_id":"media_gif_1","background_color":"#FFFFFF","output_profile":"universal_mp4_v1","output_media_id":"media_mp4_1","created_at":"2026-07-17T12:00:00Z"}}`))
		default:
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
	}))
	defer server.Close()
	dir := t.TempDir()
	path := filepath.Join(dir, "animation.gif")
	if err := os.WriteFile(path, []byte("GIF89a"), 0o600); err != nil {
		t.Fatal(err)
	}
	client := NewClient(WithAPIKey("up_test_xxx"), WithBaseURL(server.URL))
	job, err := client.Media.UploadAndConvertGIF(context.Background(), path, &GIFUploadAndConvertOptions{
		IdempotencyKey: "upload-gif-1", PollInterval: time.Millisecond,
	})
	if err != nil {
		t.Fatal(err)
	}
	if job.OutputMediaID == nil || sawPost {
		t.Fatalf("unexpected result %#v sawPost=%v", job, sawPost)
	}
}

func TestAudioOverlayCreateSendsBodyAndIdempotency(t *testing.T) {
	var body map[string]any
	var idempotencyKey string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/media/audio-overlays" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		idempotencyKey = r.Header.Get("Idempotency-Key")
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		_, _ = w.Write([]byte(`{"data":{"id":"mpj_1","status":"queued","video_media_id":"media_video_1","audio_media_id":"media_audio_1","output_media_id":null,"mode":"mix","fit":"trim_to_video","created_at":"2026-07-03T12:00:00Z"}}`))
	}))
	defer server.Close()

	client := NewClient(WithAPIKey("up_test_xxx"), WithBaseURL(server.URL))
	videoVolume := int32(70)
	job, err := client.Media.AudioOverlays.Create(context.Background(), &AudioOverlayCreateRequest{
		VideoMediaID: "media_video_1",
		AudioMediaID: "media_audio_1",
		Mode:         "mix",
		VideoVolume:  &videoVolume,
		Fit:          "trim_to_video",
	}, WithIdempotencyKey("overlay-1"))
	if err != nil {
		t.Fatal(err)
	}
	if job.ID != "mpj_1" || job.VideoMediaID != "media_video_1" {
		t.Fatalf("unexpected job %#v", job)
	}
	if idempotencyKey != "overlay-1" {
		t.Fatalf("missing idempotency header")
	}
	if body["video_media_id"] != "media_video_1" || body["audio_media_id"] != "media_audio_1" || body["mode"] != "mix" || body["fit"] != "trim_to_video" {
		t.Fatalf("unexpected body %#v", body)
	}
	if body["video_volume"].(float64) != 70 {
		t.Fatalf("unexpected volume %#v", body)
	}
}

func TestAudioOverlayGetParsesOutputMedia(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/media/audio-overlays/mpj_1" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"id":"mpj_1","status":"succeeded","video_media_id":"media_video_1","audio_media_id":"media_audio_1","output_media_id":"media_output_1","mode":"replace","fit":"loop_to_video","created_at":"2026-07-03T12:00:00Z","completed_at":"2026-07-03T12:00:20Z"}}`))
	}))
	defer server.Close()

	client := NewClient(WithAPIKey("up_test_xxx"), WithBaseURL(server.URL))
	job, err := client.Media.AudioOverlays.Get(context.Background(), "mpj_1")
	if err != nil {
		t.Fatal(err)
	}
	if job.Status != "succeeded" || job.OutputMediaID == nil || *job.OutputMediaID != "media_output_1" {
		t.Fatalf("unexpected job %#v", job)
	}
}
