package unipost

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAnalyticsPostsSendsExplorerFilters(t *testing.T) {
	var seenQuery map[string]string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/analytics/posts" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		seenQuery = map[string]string{
			"platform":   r.URL.Query().Get("platform"),
			"account_id": r.URL.Query().Get("account_id"),
			"post_id":    r.URL.Query().Get("post_id"),
			"sort":       r.URL.Query().Get("sort"),
			"limit":      r.URL.Query().Get("limit"),
			"cursor":     r.URL.Query().Get("cursor"),
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":[{"post_id":"post_1","platform":"pinterest"}],"meta":{"next_cursor":"25"}}`))
	}))
	defer server.Close()

	client := NewClient(WithAPIKey("up_test_xxx"), WithBaseURL(server.URL))
	result, err := client.Analytics.Posts(context.Background(), &AnalyticsPostsParams{
		Platform:  "pinterest",
		AccountID: "sa_1",
		PostID:    "post_1",
		Sort:      "engagement_rate",
		Limit:     25,
		Cursor:    "0",
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Data[0].PostID != "post_1" {
		t.Fatalf("unexpected post id %q", result.Data[0].PostID)
	}
	if result.NextCursor != "25" {
		t.Fatalf("unexpected next cursor %q", result.NextCursor)
	}
	if seenQuery["platform"] != "pinterest" || seenQuery["account_id"] != "sa_1" || seenQuery["post_id"] != "post_1" || seenQuery["sort"] != "engagement_rate" || seenQuery["limit"] != "25" || seenQuery["cursor"] != "0" {
		t.Fatalf("unexpected query %#v", seenQuery)
	}
}

func TestAnalyticsExportPostsCSVReturnsRawText(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/analytics/posts/export" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		if got := r.URL.Query().Get("platform"); got != "tiktok" {
			t.Fatalf("unexpected platform %q", got)
		}
		w.Header().Set("Content-Type", "text/csv")
		_, _ = w.Write([]byte("post_id,platform\npost_1,tiktok\n"))
	}))
	defer server.Close()

	client := NewClient(WithAPIKey("up_test_xxx"), WithBaseURL(server.URL))
	csv, err := client.Analytics.ExportPostsCSV(context.Background(), &AnalyticsPostsParams{Platform: "tiktok"})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(csv, "post_id,platform") {
		t.Fatalf("unexpected csv %q", csv)
	}
}

func TestAnalyticsPlatformsAndPlatformDetail(t *testing.T) {
	var requests []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests = append(requests, r.URL.RequestURI())
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/v1/analytics/platforms":
			_, _ = w.Write([]byte(`{"data":[{"platform":"tiktok","health":"ready"}]}`))
		case "/v1/analytics/platforms/tiktok":
			_, _ = w.Write([]byte(`{"data":{"platform":"tiktok","summary":{"posts":3}}}`))
		default:
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
	}))
	defer server.Close()

	client := NewClient(WithAPIKey("up_test_xxx"), WithBaseURL(server.URL))
	platforms, err := client.Analytics.Platforms(context.Background(), &AnalyticsPlatformParams{From: "2026-05-01", To: "2026-05-31"})
	if err != nil {
		t.Fatal(err)
	}
	detail, err := client.Analytics.Platform(context.Background(), "tiktok", &AnalyticsPlatformParams{ProfileID: "prof_1"})
	if err != nil {
		t.Fatal(err)
	}
	if platforms[0].Platform != "tiktok" || detail.Summary.Posts != 3 {
		t.Fatalf("unexpected payload %#v %#v", platforms, detail)
	}
	if !strings.Contains(requests[0], "from=2026-05-01") || !strings.Contains(requests[1], "profile_id=prof_1") {
		t.Fatalf("unexpected requests %#v", requests)
	}
}

func TestAnalyticsRefreshPostsBody(t *testing.T) {
	var body map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/v1/analytics/refresh" {
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"status":"queued","matched_count":7,"requested_count":5,"limit":5}}`))
	}))
	defer server.Close()

	client := NewClient(WithAPIKey("up_test_xxx"), WithBaseURL(server.URL))
	result, err := client.Analytics.Refresh(context.Background(), &AnalyticsRefreshParams{Platform: "threads", Limit: 5})
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != "queued" || result.RequestedCount != 5 {
		t.Fatalf("unexpected refresh result %#v", result)
	}
	if body["platform"] != "threads" || body["limit"].(float64) != 5 {
		t.Fatalf("unexpected body %#v", body)
	}
}
