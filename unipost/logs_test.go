package unipost

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestLogsListSendsFiltersAndReadsCursor(t *testing.T) {
	var seenQuery map[string]string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/logs" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		seenQuery = map[string]string{
			"status":     r.URL.Query().Get("status"),
			"level":      r.URL.Query().Get("level"),
			"profile_id": r.URL.Query().Get("profile_id"),
			"error_code": r.URL.Query().Get("error_code"),
			"limit":      r.URL.Query().Get("limit"),
			"cursor":     r.URL.Query().Get("cursor"),
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":[{"id":110,"action":"post.publish.failed","status":"error"}],"meta":{"limit":25,"has_more":true,"next_cursor":"cur_abc"}}`))
	}))
	defer server.Close()

	client := NewClient(WithAPIKey("up_test_xxx"), WithBaseURL(server.URL))
	result, err := client.Logs.List(context.Background(), &LogListParams{
		Status:    "error",
		Level:     "warn",
		ProfileID: "prof_1",
		ErrorCode: "provider_failed",
		Limit:     25,
		Cursor:    "cur_prev",
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Data[0].ID != 110 || result.NextCursor != "cur_abc" {
		t.Fatalf("unexpected result %#v", result)
	}
	if seenQuery["status"] != "error" || seenQuery["level"] != "warn" || seenQuery["profile_id"] != "prof_1" || seenQuery["error_code"] != "provider_failed" || seenQuery["limit"] != "25" || seenQuery["cursor"] != "cur_prev" {
		t.Fatalf("unexpected query %#v", seenQuery)
	}
}

func TestLogsGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/logs/110" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"id":110,"action":"post.publish.failed","request_payload":null}}`))
	}))
	defer server.Close()

	client := NewClient(WithAPIKey("up_test_xxx"), WithBaseURL(server.URL))
	log, err := client.Logs.Get(context.Background(), 110)
	if err != nil {
		t.Fatal(err)
	}
	if log.ID != 110 || log.Action != "post.publish.failed" {
		t.Fatalf("unexpected log %#v", log)
	}
}

func TestLogsStreamReadsReplayEvent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/logs/stream" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		if r.Header.Get("Accept") != "text/event-stream" {
			t.Fatalf("missing SSE accept header")
		}
		if r.URL.Query().Get("status") != "error" || r.URL.Query().Get("after_id") != "109" {
			t.Fatalf("unexpected query %s", r.URL.RawQuery)
		}
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = w.Write([]byte("event: log.created\nid: 110\ndata: {\"id\":110,\"action\":\"post.publish.failed\",\"status\":\"error\"}\n\n"))
	}))
	defer server.Close()

	client := NewClient(WithAPIKey("up_test_xxx"), WithBaseURL(server.URL))
	stream, err := client.Logs.Stream(context.Background(), &LogStreamParams{Status: "error", AfterID: 109})
	if err != nil {
		t.Fatal(err)
	}
	defer stream.Close()

	if !stream.Next() {
		t.Fatalf("expected event, err=%v", stream.Err())
	}
	event := stream.Event()
	if event.ID != 110 || !strings.Contains(event.Action, "failed") {
		t.Fatalf("unexpected event %#v", event)
	}
}
