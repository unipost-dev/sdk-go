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
