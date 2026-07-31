package unipost

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"sync/atomic"
	"testing"
)

func TestBillingXCreditsReturnsTypedFullEnvelopes(t *testing.T) {
	var requestNumber atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch requestNumber.Add(1) {
		case 1:
			if r.URL.Path != "/v1/billing/x-credits" {
				t.Fatalf("unexpected allowance path %s", r.URL.Path)
			}
			_, _ = w.Write([]byte(`{
				"data":{
					"mode":"unipost_managed_app",
					"plan_id":"pro",
					"monthly_allowance":10000,
					"monthly_used":200,
					"monthly_finalized":150,
					"monthly_pending":50,
					"monthly_effective":200,
					"monthly_remaining":9800,
					"billing_period_start":"2026-07-19T00:00:00Z",
					"billing_period_end":"2026-08-19T00:00:00Z",
					"catalog_version":"2026-07-29",
					"inbound_daily_usage":0,
					"inbound_daily_limit":1000,
					"inbound_events_accepted":0,
					"inbound_events_suppressed":0,
					"inbound_daily_reset_at":"2026-08-01T00:00:00Z",
					"inbound_daily_percent":0,
					"pause_paid_sources":false,
					"connection_mode_note":"UniPost-managed X app"
				},
				"request_id":"req_allowance_1"
			}`))
		case 2:
			if r.URL.Path != "/v1/billing/x-credits/events" {
				t.Fatalf("unexpected events path %s", r.URL.Path)
			}
			if got := r.URL.Query(); !reflect.DeepEqual(got, url.Values{
				"account_id":       {"sa_x_123"},
				"external_user_id": {"user_42"},
				"operation":        {"post.read"},
				"status":           {"succeeded"},
				"start_time":       {"2026-07-01T00:00:00Z"},
				"end_time":         {"2026-08-01T00:00:00Z"},
				"cursor":           {"events_current"},
				"limit":            {"50"},
			}) {
				t.Fatalf("unexpected events query %#v", got)
			}
			_, _ = w.Write([]byte(`{
				"data":[{
					"operation_id":"xread_posts_1",
					"account_id":"sa_x_123",
					"external_user_id":"user_42",
					"operation":"post.read",
					"catalog_version":"2026-07-29",
					"estimated":100,
					"reserved":100,
					"charged":80,
					"released":20,
					"status":"succeeded",
					"created_at":"2026-07-31T10:00:00Z",
					"updated_at":"2026-07-31T10:00:01Z",
					"finalized_at":"2026-07-31T10:00:01Z"
				}],
				"meta":{"next_cursor":"events_next","has_more":true,"limit":50},
				"request_id":"req_events_1"
			}`))
		default:
			t.Fatal("unexpected extra request")
		}
	}))
	defer server.Close()

	client := NewClient(WithAPIKey("up_test_xxx"), WithBaseURL(server.URL))
	allowance, err := client.Billing.GetXCredits(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if allowance.RequestID != "req_allowance_1" || allowance.Data.MonthlyRemaining != 9800 {
		t.Fatalf("unexpected allowance %#v", allowance)
	}
	events, err := client.Billing.ListXCreditEvents(context.Background(), &ListXCreditEventsParams{
		AccountID:      "sa_x_123",
		ExternalUserID: "user_42",
		Operation:      "post.read",
		Status:         "succeeded",
		StartTime:      "2026-07-01T00:00:00Z",
		EndTime:        "2026-08-01T00:00:00Z",
		Cursor:         "events_current",
		Limit:          50,
	})
	if err != nil {
		t.Fatal(err)
	}
	if events.RequestID != "req_events_1" || events.Meta.NextCursor != "events_next" {
		t.Fatalf("unexpected events %#v", events)
	}
	if len(events.Data) != 1 || events.Data[0].Charged != 80 {
		t.Fatalf("unexpected event data %#v", events.Data)
	}
}

func TestBillingRejectsInvalidEventLimitBeforeRequest(t *testing.T) {
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
	}))
	defer server.Close()
	client := NewClient(WithAPIKey("up_test_xxx"), WithBaseURL(server.URL))
	for _, limit := range []int{-1, 101} {
		if _, err := client.Billing.ListXCreditEvents(
			context.Background(),
			&ListXCreditEventsParams{Limit: limit},
		); err == nil {
			t.Fatalf("expected limit %d to fail", limit)
		}
	}
	if got := requests.Load(); got != 0 {
		t.Fatalf("invalid limits made %d requests", got)
	}
}
