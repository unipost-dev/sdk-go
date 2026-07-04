package unipost

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
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
