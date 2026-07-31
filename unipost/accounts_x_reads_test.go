package unipost

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"sync/atomic"
	"testing"
)

func TestAccountsProfileAndPostsUseExactRequestsAndDecodeFullEnvelopes(t *testing.T) {
	var requestNumber atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch requestNumber.Add(1) {
		case 1:
			if r.Method != http.MethodGet || r.URL.Path != "/v1/accounts/sa_x_123/profile" {
				t.Fatalf("unexpected profile request %s %s", r.Method, r.URL.Path)
			}
			if got := r.URL.Query(); !reflect.DeepEqual(got, url.Values{
				"external_user_id": {"user_42"},
			}) {
				t.Fatalf("unexpected profile query %#v", got)
			}
			if got := r.Header.Get("Idempotency-Key"); got != "profile-user-42" {
				t.Fatalf("unexpected profile idempotency key %q", got)
			}
			_, _ = w.Write([]byte(`{
				"data":{
					"account_id":"sa_x_123",
					"platform":"twitter",
					"external_account_id":"2244994945",
					"username":"unipost",
					"display_name":"UniPost",
					"description":"Social publishing infrastructure",
					"profile_image_url":"https://example.test/profile.jpg",
					"account_created_at":"2022-01-12T09:30:00Z",
					"verified":false,
					"public_metrics":{"followers":1200,"following":180,"posts":640,"listed":12},
					"retrieved_at":"2026-07-31T10:00:00Z"
				},
				"meta":{
					"credits":{
						"operation_id":"xread_profile_1",
						"status":"succeeded",
						"accounting_enabled":true,
						"billing_mode":"unipost_managed_app",
						"operation":"user.read",
						"estimated":10,
						"reserved":10,
						"charged":10,
						"released":0,
						"catalog_version":"2026-07-29"
					}
				},
				"request_id":"req_profile_1"
			}`))
		case 2:
			if r.Method != http.MethodGet || r.URL.Path != "/v1/accounts/sa_x_123/posts" {
				t.Fatalf("unexpected posts request %s %s", r.Method, r.URL.Path)
			}
			if got := r.URL.Query(); !reflect.DeepEqual(got, url.Values{
				"external_user_id":          {"user_42"},
				"limit":                     {"20"},
				"cursor":                    {"xc_current"},
				"start_time":                {"2026-07-01T00:00:00Z"},
				"end_time":                  {"2026-08-01T00:00:00Z"},
				"exclude_reposts":           {"true"},
				"exclude_replies_to_others": {"true"},
			}) {
				t.Fatalf("unexpected posts query %#v", got)
			}
			if got := r.Header.Get("Idempotency-Key"); got != "posts-user-42-page-2" {
				t.Fatalf("unexpected posts idempotency key %q", got)
			}
			_, _ = w.Write([]byte(`{
				"data":[{
					"account_id":"sa_x_123",
					"external_post_id":"1819012345678901234",
					"text":"Shipping today.",
					"created_at":"2026-07-30T17:20:00Z",
					"conversation_id":"1819012345678901234",
					"content_type":"original_post",
					"is_reply":false,
					"is_self_reply":false,
					"is_repost":false,
					"is_quote":false,
					"media":[{"type":"image"}],
					"public_metrics":{"likes":18,"replies":2,"reposts":3,"quotes":1,"impressions":940},
					"thread":{"thread_id":"1819012345678901234"}
				}],
				"meta":{
					"limit":20,
					"scanned_count":20,
					"returned_count":1,
					"has_more":true,
					"next_cursor":"xc_next",
					"cursor_expires_at":"2026-08-07T10:00:00Z",
					"credits":{
						"operation_id":"xread_posts_1",
						"status":"succeeded",
						"accounting_enabled":false,
						"billing_mode":"customer_x_app",
						"bypass_reason":"customer_x_app",
						"operation":"post.read",
						"estimated":100,
						"reserved":0,
						"charged":0,
						"released":0,
						"catalog_version":"2026-07-29"
					},
					"replayed":false
				},
				"request_id":"req_posts_1"
			}`))
		default:
			t.Fatalf("unexpected extra request")
		}
	}))
	defer server.Close()

	client := NewClient(WithAPIKey("up_test_xxx"), WithBaseURL(server.URL))
	profile, err := client.Accounts.Profile(context.Background(), "sa_x_123", &XAccountProfileParams{
		ExternalUserID: "user_42",
		IdempotencyKey: "profile-user-42",
	})
	if err != nil {
		t.Fatal(err)
	}
	if profile.RequestID != "req_profile_1" || profile.Data.Username != "unipost" {
		t.Fatalf("unexpected profile %#v", profile)
	}
	if profile.Meta.Replayed != nil || profile.Meta.Credits.Charged != 10 {
		t.Fatalf("unexpected profile metadata %#v", profile.Meta)
	}

	posts, err := client.Accounts.ListPosts(context.Background(), "sa_x_123", &XAccountPostsParams{
		ExternalUserID:         "user_42",
		IdempotencyKey:         "posts-user-42-page-2",
		Limit:                  20,
		Cursor:                 "xc_current",
		StartTime:              "2026-07-01T00:00:00Z",
		EndTime:                "2026-08-01T00:00:00Z",
		ExcludeReposts:         true,
		ExcludeRepliesToOthers: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if posts.RequestID != "req_posts_1" || len(posts.Data) != 1 {
		t.Fatalf("unexpected posts %#v", posts)
	}
	if posts.Data[0].Media[0].Type != "image" || posts.Meta.NextCursor != "xc_next" {
		t.Fatalf("unexpected posts envelope %#v", posts)
	}
	if posts.Meta.Credits.BypassReason != "customer_x_app" {
		t.Fatalf("unexpected credits %#v", posts.Meta.Credits)
	}
}

func TestAccountsXReadsRejectInvalidInputsBeforeRequest(t *testing.T) {
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()
	client := NewClient(WithAPIKey("up_test_xxx"), WithBaseURL(server.URL))

	profileCases := []struct {
		accountID string
		params    *XAccountProfileParams
	}{
		{"", &XAccountProfileParams{ExternalUserID: "user_42", IdempotencyKey: "key"}},
		{"sa_x_123", &XAccountProfileParams{ExternalUserID: " ", IdempotencyKey: "key"}},
		{"sa_x_123", &XAccountProfileParams{ExternalUserID: "user_42", IdempotencyKey: " "}},
		{"sa_x_123", nil},
	}
	for _, tt := range profileCases {
		if _, err := client.Accounts.Profile(context.Background(), tt.accountID, tt.params); err == nil {
			t.Fatalf("expected profile validation failure for %#v", tt)
		}
	}

	for _, params := range []*XAccountPostsParams{
		{ExternalUserID: "user_42", IdempotencyKey: "key", Limit: 4},
		{ExternalUserID: "user_42", IdempotencyKey: "key", Limit: 101},
		{ExternalUserID: "user_42", IdempotencyKey: "key", Limit: 20, StartTime: "2026-07-01"},
		{
			ExternalUserID: "user_42",
			IdempotencyKey: "key",
			Limit:          20,
			StartTime:      "2026-08-01T00:00:00Z",
			EndTime:        "2026-08-01T00:00:00Z",
		},
	} {
		if _, err := client.Accounts.ListPosts(context.Background(), "sa_x_123", params); err == nil {
			t.Fatalf("expected posts validation failure for %#v", params)
		}
	}
	if got := requests.Load(); got != 0 {
		t.Fatalf("invalid inputs made %d requests", got)
	}
}

func TestXAccountReadErrorPreservesDetailsRetryAndAPIErrorCompatibility(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Retry-After", "7")
		w.WriteHeader(http.StatusConflict)
		_, _ = w.Write([]byte(`{
			"error":{
				"code":"READ_IN_PROGRESS",
				"message":"This logical read is still running",
				"details":{"operation_id":"xread_1","retry_cursor":"xc_retry"},
				"is_retriable":true
			},
			"request_id":"req_error_1"
		}`))
	}))
	defer server.Close()

	client := NewClient(WithAPIKey("up_test_xxx"), WithBaseURL(server.URL))
	_, err := client.Accounts.Profile(context.Background(), "sa_x_123", &XAccountProfileParams{
		ExternalUserID: "user_42",
		IdempotencyKey: "profile-user-42",
	})
	var readErr *XAccountReadError
	if !errors.As(err, &readErr) {
		t.Fatalf("expected XAccountReadError, got %T: %v", err, err)
	}
	if readErr.RetryAfter != 7 || readErr.IsRetriable == nil || !*readErr.IsRetriable {
		t.Fatalf("unexpected retry metadata %#v", readErr)
	}
	if got := readErr.Details["retry_cursor"]; got != "xc_retry" {
		t.Fatalf("unexpected details %#v", readErr.Details)
	}
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected wrapped APIError, got %T: %v", err, err)
	}
	if apiErr.Code != "READ_IN_PROGRESS" || apiErr.RequestID != "req_error_1" {
		t.Fatalf("unexpected APIError %#v", apiErr)
	}
}

func TestAPIErrorUnkeyedLiteralCompatibility(t *testing.T) {
	_ = APIError{0, "", "", "", nil, "", 0, "", "", "", nil, nil}
}
