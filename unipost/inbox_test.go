package unipost

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

const canonicalInboxReplyItem = `{
	"id":"inbox_1",
	"social_account_id":"acct_1",
	"workspace_id":"ws_1",
	"source":"x_dm",
	"external_id":"external_1",
	"thread_key":"thread_1",
	"thread_status":"open",
	"is_read":false,
	"is_own":true,
	"received_at":"2026-07-22T00:00:00Z",
	"created_at":"2026-07-22T00:00:01Z"
}`

func TestNewClientInitializesInbox(t *testing.T) {
	client := NewClient(WithAPIKey("up_test_xxx"))
	if client.Inbox == nil {
		t.Fatal("expected Inbox service")
	}
}

func TestInboxManagedUserListSerializesScopeFiltersAndDecodesEnvelope(t *testing.T) {
	wantQuery := url.Values{
		"external_user_id": {"user A"},
		"inbox_scope":      {"managed_user"},
		"is_own":           {"true"},
		"is_read":          {"false"},
		"limit":            {"25"},
		"source":           {"x_dm"},
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("unexpected method %s", r.Method)
		}
		if r.URL.Path != "/v1/inbox" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		if got := r.URL.Query(); !reflect.DeepEqual(got, wantQuery) {
			t.Fatalf("unexpected query %#v, want %#v", got, wantQuery)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"data":[{
				"id":"inbox_1",
				"social_account_id":"acct_1",
				"workspace_id":"ws_1",
				"source":"x_dm",
				"external_id":"external_1",
				"thread_key":"thread_1",
				"thread_status":"assigned",
				"is_read":false,
				"is_own":true,
				"received_at":"2026-07-22T00:00:00Z",
				"created_at":"2026-07-22T00:00:01Z",
				"parent_external_id":"parent_1",
				"assigned_to":"agent_1",
				"linked_post_id":"post_1",
				"author_name":"Ada",
				"author_id":"author_1",
				"author_avatar_url":"https://example.test/author.png",
				"body":"hello",
				"account_name":"Support",
				"account_platform":"x",
				"account_avatar_url":"https://example.test/account.png",
				"x_credits_counted":2,
				"x_credit_operation":"dm_read",
				"x_credit_catalog_version":"2026-07",
				"x_credit_billing_mode":"metered",
				"url":"https://x.example/message/1"
			}],
			"request_id":"req_1"
		}`))
	}))
	defer server.Close()

	client := NewClient(WithAPIKey("up_test_xxx"), WithBaseURL(server.URL))
	externalUserID := "user A"
	managedInbox := client.Inbox.ManagedUser(externalUserID)
	externalUserID = "attacker"
	read := false
	own := true
	result, err := managedInbox.List(context.Background(), &InboxListParams{
		Source: InboxSourceXDM,
		IsRead: &read,
		IsOwn:  &own,
		Limit:  25,
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.RequestID == nil || *result.RequestID != "req_1" || len(result.Data) != 1 {
		t.Fatalf("unexpected list result %#v", result)
	}
	item := result.Data[0]
	if item.ID != "inbox_1" || item.Source != InboxSourceXDM || item.ThreadStatus != InboxThreadStatusAssigned {
		t.Fatalf("unexpected required fields %#v", item)
	}
	if item.Body == nil || *item.Body != "hello" || item.XCreditsCounted == nil || *item.XCreditsCounted != 2 {
		t.Fatalf("unexpected optional fields %#v", item)
	}
}

func TestInboxWorkspaceListOmitsExternalIDAndNilFilters(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		want := url.Values{"inbox_scope": {"workspace"}}
		if got := r.URL.Query(); !reflect.DeepEqual(got, want) {
			t.Fatalf("unexpected query %#v, want %#v", got, want)
		}
		if r.URL.Query().Has("external_user_id") {
			t.Fatal("workspace request must omit external_user_id")
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":[]}`))
	}))
	defer server.Close()

	client := NewClient(WithAPIKey("up_test_xxx"), WithBaseURL(server.URL))
	result, err := client.Inbox.Workspace().List(context.Background(), &InboxListParams{})
	if err != nil {
		t.Fatal(err)
	}
	if result.RequestID != nil || len(result.Data) != 0 {
		t.Fatalf("unexpected list result %#v", result)
	}
}

func TestInboxManagedUserRejectsBlankIDBeforeRequest(t *testing.T) {
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := NewClient(WithAPIKey("up_test_xxx"), WithBaseURL(server.URL))
	for _, externalUserID := range []string{"", "   ", "\t\n"} {
		t.Run(externalUserID, func(t *testing.T) {
			result, err := client.Inbox.ManagedUser(externalUserID).List(context.Background(), nil)
			if err == nil {
				t.Fatalf("expected error, got result %#v", result)
			}
		})
	}
	if got := requests.Load(); got != 0 {
		t.Fatalf("invalid managed-user IDs made %d requests", got)
	}
}

func TestInboxPublicListTypesExposeOnlyTheStableContract(t *testing.T) {
	paramsType := reflect.TypeOf(InboxListParams{})
	wantParamFields := []string{"Source", "IsRead", "IsOwn", "Limit"}
	if paramsType.NumField() != len(wantParamFields) {
		t.Fatalf("InboxListParams exposes %d fields, want %d", paramsType.NumField(), len(wantParamFields))
	}
	for i, want := range wantParamFields {
		if got := paramsType.Field(i).Name; got != want {
			t.Fatalf("InboxListParams field %d is %q, want %q", i, got, want)
		}
	}

	responseType := reflect.TypeOf(InboxListResponse{})
	wantResponseFields := []string{"Data", "RequestID"}
	if responseType.NumField() != len(wantResponseFields) {
		t.Fatalf("InboxListResponse exposes pagination or other fields: %#v", responseType)
	}
	for i, want := range wantResponseFields {
		if got := responseType.Field(i).Name; got != want {
			t.Fatalf("InboxListResponse field %d is %q, want %q", i, got, want)
		}
	}
}

func TestInboxItemMatchesCanonicalJSONContract(t *testing.T) {
	wantFields := []struct {
		name     string
		jsonName string
		optional bool
	}{
		{"ID", "id", false},
		{"SocialAccountID", "social_account_id", false},
		{"WorkspaceID", "workspace_id", false},
		{"Source", "source", false},
		{"ExternalID", "external_id", false},
		{"ThreadKey", "thread_key", false},
		{"ThreadStatus", "thread_status", false},
		{"IsRead", "is_read", false},
		{"IsOwn", "is_own", false},
		{"ReceivedAt", "received_at", false},
		{"CreatedAt", "created_at", false},
		{"ParentExternalID", "parent_external_id", true},
		{"AssignedTo", "assigned_to", true},
		{"LinkedPostID", "linked_post_id", true},
		{"AuthorName", "author_name", true},
		{"AuthorID", "author_id", true},
		{"AuthorAvatarURL", "author_avatar_url", true},
		{"Body", "body", true},
		{"AccountName", "account_name", true},
		{"AccountPlatform", "account_platform", true},
		{"AccountAvatarURL", "account_avatar_url", true},
		{"XCreditsCounted", "x_credits_counted", true},
		{"XCreditOperation", "x_credit_operation", true},
		{"XCreditCatalogVersion", "x_credit_catalog_version", true},
		{"XCreditBillingMode", "x_credit_billing_mode", true},
		{"URL", "url", true},
	}

	itemType := reflect.TypeOf(InboxItem{})
	if itemType.NumField() != len(wantFields) {
		t.Fatalf("InboxItem has %d fields, want %d", itemType.NumField(), len(wantFields))
	}
	for i, want := range wantFields {
		field := itemType.Field(i)
		if field.Name != want.name {
			t.Fatalf("InboxItem field %d is %q, want %q", i, field.Name, want.name)
		}
		wantTag := want.jsonName
		if want.optional {
			wantTag += ",omitempty"
			if field.Type.Kind() != reflect.Pointer {
				t.Fatalf("optional field %s must be a pointer, got %s", field.Name, field.Type)
			}
		}
		if got := field.Tag.Get("json"); got != wantTag {
			t.Fatalf("InboxItem.%s json tag is %q, want %q", field.Name, got, wantTag)
		}
	}
}

func TestInboxConstantsMatchWireValues(t *testing.T) {
	sources := []InboxSource{
		InboxSourceIGComment,
		InboxSourceIGDM,
		InboxSourceThreadsReply,
		InboxSourceFBComment,
		InboxSourceFBDM,
		InboxSourceXReply,
		InboxSourceXDM,
	}
	if got := []InboxSource{"ig_comment", "ig_dm", "threads_reply", "fb_comment", "fb_dm", "x_reply", "x_dm"}; !reflect.DeepEqual(sources, got) {
		t.Fatalf("unexpected InboxSource constants %#v", sources)
	}
	statuses := []InboxThreadStatus{InboxThreadStatusOpen, InboxThreadStatusAssigned, InboxThreadStatusResolved}
	if got := []InboxThreadStatus{"open", "assigned", "resolved"}; !reflect.DeepEqual(statuses, got) {
		t.Fatalf("unexpected InboxThreadStatus constants %#v", statuses)
	}
}

func TestInboxReplyCompletedUsesExactRequestAndRetainsOperationID(t *testing.T) {
	var attempts atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts.Add(1)
		if r.Method != http.MethodPost {
			t.Fatalf("unexpected method %s", r.Method)
		}
		if r.URL.EscapedPath() != "/v1/inbox/item%2Fwith%20space/reply" {
			t.Fatalf("unexpected escaped path %q", r.URL.EscapedPath())
		}
		wantQuery := url.Values{
			"external_user_id": {"managed user"},
			"inbox_scope":      {"managed_user"},
		}
		if got := r.URL.Query(); !reflect.DeepEqual(got, wantQuery) {
			t.Fatalf("unexpected query %#v, want %#v", got, wantQuery)
		}
		if got := r.Header.Get("Idempotency-Key"); got != "reply-01" {
			t.Fatalf("unexpected idempotency key %q", got)
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(body, []byte(`{"text":"hello"}`)) {
			t.Fatalf("unexpected request body %q", body)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-UniPost-Operation-Id", " op_completed ")
		_, _ = w.Write([]byte(`{"data":` + canonicalInboxReplyItem + `}`))
	}))
	defer server.Close()

	client := NewClient(WithAPIKey("up_test_xxx"), WithBaseURL(server.URL))
	result, err := client.Inbox.ManagedUser("managed user").Reply(
		context.Background(),
		"item/with space",
		&InboxReplyRequest{Text: "hello"},
		WithIdempotencyKey("reply-01"),
	)
	if err != nil {
		t.Fatal(err)
	}
	if result.State != InboxReplyStateCompleted {
		t.Fatalf("unexpected state %q", result.State)
	}
	if result.Item == nil || result.Item.ID != "inbox_1" {
		t.Fatalf("unexpected completed item %#v", result.Item)
	}
	if result.OperationID != "op_completed" {
		t.Fatalf("unexpected operation ID %q", result.OperationID)
	}
	if result.Code != "" || result.Message != "" || result.RequestID != nil {
		t.Fatalf("completed result exposed reconciling fields %#v", result)
	}
	if got := attempts.Load(); got != 1 {
		t.Fatalf("reply made %d attempts, want 1", got)
	}
}

func TestInboxReplyReconcilingIsAnExplicitNonCompletedResult(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-UniPost-Operation-Id", " op_reconcile ")
		w.WriteHeader(http.StatusAccepted)
		_, _ = w.Write([]byte(`{
			"error": {
				"code": "X_REMOTE_ACCEPTED_RECONCILING",
				"message": "Remote accepted; poll the operation"
			},
			"request_id": "req_202"
		}`))
	}))
	defer server.Close()

	client := NewClient(WithAPIKey("up_test_xxx"), WithBaseURL(server.URL))
	result, err := client.Inbox.Workspace().Reply(
		context.Background(),
		"inbox_1",
		&InboxReplyRequest{Text: "hello"},
	)
	if err != nil {
		t.Fatal(err)
	}
	if result.State != InboxReplyStateReconciling {
		t.Fatalf("unexpected state %q", result.State)
	}
	if result.Item != nil {
		t.Fatalf("reconciling result must not look completed: %#v", result.Item)
	}
	if result.OperationID != "op_reconcile" || result.Code != "X_REMOTE_ACCEPTED_RECONCILING" || result.Message == "" {
		t.Fatalf("unexpected reconciling result %#v", result)
	}
	if result.RequestID == nil || *result.RequestID != "req_202" {
		t.Fatalf("unexpected request ID %#v", result.RequestID)
	}
}

func TestInboxReplyMalformedSuccessFailsClosedWithoutLeakingSecrets(t *testing.T) {
	tests := []struct {
		name   string
		status int
		header string
		body   string
	}{
		{name: "200 missing data", status: http.StatusOK, body: `{"secret":"response-secret"}`},
		{name: "200 data missing required field", status: http.StatusOK, body: `{"data":{"id":"inbox_1"},"secret":"response-secret"}`},
		{name: "202 missing operation id", status: http.StatusAccepted, body: `{"error":{"code":"X_REMOTE_ACCEPTED_RECONCILING","message":"response-secret"}}`},
		{name: "202 wrong code", status: http.StatusAccepted, header: "op_1", body: `{"error":{"code":"PLATFORM_ERROR","message":"response-secret"}}`},
		{name: "202 missing message", status: http.StatusAccepted, header: "op_1", body: `{"error":{"code":"X_REMOTE_ACCEPTED_RECONCILING"},"secret":"response-secret"}`},
		{name: "unexpected success status", status: http.StatusNoContent, body: `response-secret`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var attempts atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				attempts.Add(1)
				if tt.header != "" {
					w.Header().Set("X-UniPost-Operation-Id", tt.header)
				}
				w.WriteHeader(tt.status)
				_, _ = w.Write([]byte(tt.body))
			}))
			defer server.Close()

			client := NewClient(WithAPIKey("up_test_xxx"), WithBaseURL(server.URL))
			result, err := client.Inbox.Workspace().Reply(
				context.Background(),
				"inbox_1",
				&InboxReplyRequest{Text: "hello"},
				WithIdempotencyKey("idempotency-secret"),
			)
			if err == nil {
				t.Fatalf("expected error, got %#v", result)
			}
			if got := err.Error(); got != "unipost: invalid Inbox reply response" {
				t.Fatalf("unexpected safe error %q", got)
			}
			if strings.Contains(err.Error(), "response-secret") || strings.Contains(err.Error(), "idempotency-secret") {
				t.Fatalf("decode error leaked a secret: %q", err)
			}
			if got := attempts.Load(); got != 1 {
				t.Fatalf("malformed reply made %d attempts, want 1", got)
			}
		})
	}
}

func TestInboxReplyPreservesExplicitAPIErrorsWithoutRetry(t *testing.T) {
	tests := []struct {
		name   string
		status int
		code   string
	}{
		{name: "bad request validation", status: http.StatusBadRequest, code: "VALIDATION_ERROR"},
		{name: "monthly usage", status: http.StatusPaymentRequired, code: "X_MONTHLY_USAGE_LIMIT_EXCEEDED"},
		{name: "reconnect", status: http.StatusConflict, code: "X_RECONNECT_REQUIRED"},
		{name: "legacy reconnect", status: http.StatusConflict, code: "NEEDS_RECONNECT"},
		{name: "idempotency conflict", status: http.StatusConflict, code: "IDEMPOTENCY_KEY_CONFLICT"},
		{name: "write pending", status: http.StatusConflict, code: "X_WRITE_OUTCOME_PENDING"},
		{name: "write reconciliation", status: http.StatusConflict, code: "X_WRITE_NEEDS_RECONCILIATION"},
		{name: "usage reversal", status: http.StatusConflict, code: "X_USAGE_REVERSAL_PENDING"},
		{name: "unprocessable validation", status: http.StatusUnprocessableEntity, code: "VALIDATION_ERROR"},
		{name: "platform", status: http.StatusUnprocessableEntity, code: "PLATFORM_ERROR"},
		{name: "rate limited", status: http.StatusTooManyRequests, code: "RATE_LIMITED"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var attempts atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				attempts.Add(1)
				w.Header().Set("Retry-After", "1")
				w.WriteHeader(tt.status)
				_, _ = w.Write([]byte(`{"error":{"code":"` + tt.code + `","message":"expected message"},"request_id":"req_error"}`))
			}))
			defer server.Close()

			client := NewClient(WithAPIKey("up_test_xxx"), WithBaseURL(server.URL))
			result, err := client.Inbox.Workspace().Reply(context.Background(), "inbox_1", &InboxReplyRequest{Text: "hello"})
			if err == nil {
				t.Fatalf("expected API error, got %#v", result)
			}
			var apiErr *APIError
			if !errors.As(err, &apiErr) {
				t.Fatalf("expected *APIError, got %T: %v", err, err)
			}
			if apiErr.Status != tt.status || apiErr.Code != tt.code || apiErr.RequestID != "req_error" {
				t.Fatalf("unexpected API error %#v", apiErr)
			}
			if got := attempts.Load(); got != 1 {
				t.Fatalf("reply made %d attempts for %s, want 1", got, tt.code)
			}
		})
	}
}

func TestInboxReplyDoesNotFollowRedirectAndPreservesCustomClientTimeout(t *testing.T) {
	var sourceAttempts atomic.Int32
	var targetAttempts atomic.Int32
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		targetAttempts.Add(1)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data":` + canonicalInboxReplyItem + `}`))
	}))
	defer target.Close()
	source := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sourceAttempts.Add(1)
		http.Redirect(w, r, target.URL, http.StatusTemporaryRedirect)
	}))
	defer source.Close()

	customClient := &http.Client{Timeout: 137 * time.Millisecond}
	client := NewClient(
		WithAPIKey("up_test_xxx"),
		WithBaseURL(source.URL),
		WithHTTPClient(customClient),
	)
	result, err := client.Inbox.Workspace().Reply(context.Background(), "inbox_1", &InboxReplyRequest{Text: "hello"})
	if result != nil {
		t.Fatalf("expected no result, got %#v", result)
	}
	var apiErr *APIError
	if !errors.As(err, &apiErr) || apiErr.Status != http.StatusTemporaryRedirect {
		t.Fatalf("expected redirect *APIError, got %T: %v", err, err)
	}
	if sourceAttempts.Load() != 1 || targetAttempts.Load() != 0 {
		t.Fatalf("redirect attempts source=%d target=%d", sourceAttempts.Load(), targetAttempts.Load())
	}
	if customClient.Timeout != 137*time.Millisecond || customClient.CheckRedirect != nil {
		t.Fatalf("reply mutated custom client: %#v", customClient)
	}
}

func TestInboxReplyRejectsUnsafeItemIDsBeforeRequest(t *testing.T) {
	var attempts atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts.Add(1)
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := NewClient(WithAPIKey("up_test_xxx"), WithBaseURL(server.URL))
	for _, id := range []string{"", "   ", ".", "..", " .. "} {
		t.Run(id, func(t *testing.T) {
			result, err := client.Inbox.Workspace().Reply(context.Background(), id, &InboxReplyRequest{Text: "hello"})
			if err == nil {
				t.Fatalf("expected error, got %#v", result)
			}
		})
	}
	if got := attempts.Load(); got != 0 {
		t.Fatalf("unsafe IDs made %d requests", got)
	}
}

type inboxReplyRoundTripperFunc func(*http.Request) (*http.Response, error)

func (f inboxReplyRoundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestResponseAwareTransportRetainsClonedMetadata(t *testing.T) {
	sharedHeader := make(http.Header)
	sharedHeader.Set("X-UniPost-Operation-Id", "op_original")
	customClient := &http.Client{
		Timeout: 211 * time.Millisecond,
		Transport: inboxReplyRoundTripperFunc(func(req *http.Request) (*http.Response, error) {
			if req.Header.Get("Authorization") != "Bearer up_test_xxx" || req.Header.Get("User-Agent") != userAgent {
				t.Fatalf("missing shared transport headers: %#v", req.Header)
			}
			return &http.Response{
				StatusCode: http.StatusAccepted,
				Header:     sharedHeader,
				Body:       io.NopCloser(strings.NewReader(`{"ok":true}`)),
				Request:    req,
			}, nil
		}),
	}
	client := NewClient(WithAPIKey("up_test_xxx"), WithBaseURL("https://example.test"), WithHTTPClient(customClient))
	response, err := client.doResponseOnce(
		context.Background(),
		http.MethodPost,
		"/v1/inbox/inbox_1/reply",
		map[string]string{"inbox_scope": "workspace"},
		&InboxReplyRequest{Text: "hello"},
		nil,
	)
	if err != nil {
		t.Fatal(err)
	}
	sharedHeader.Set("X-UniPost-Operation-Id", "op_mutated")
	if response.StatusCode != http.StatusAccepted || response.Header.Get("X-UniPost-Operation-Id") != "op_original" || string(response.Body) != `{"ok":true}` {
		t.Fatalf("response metadata was not retained independently: %#v", response)
	}
	if customClient.Timeout != 211*time.Millisecond || customClient.CheckRedirect != nil {
		t.Fatalf("response-aware transport mutated custom client: %#v", customClient)
	}
}
