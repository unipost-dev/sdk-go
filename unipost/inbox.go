package unipost

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// InboxSource identifies the normalized provider source for an Inbox item.
type InboxSource string

const (
	InboxSourceIGComment    InboxSource = "ig_comment"
	InboxSourceIGDM         InboxSource = "ig_dm"
	InboxSourceThreadsReply InboxSource = "threads_reply"
	InboxSourceFBComment    InboxSource = "fb_comment"
	InboxSourceFBDM         InboxSource = "fb_dm"
	InboxSourceXReply       InboxSource = "x_reply"
	InboxSourceXDM          InboxSource = "x_dm"
)

// InboxThreadStatus identifies the workflow state of an Inbox thread.
type InboxThreadStatus string

const (
	InboxThreadStatusOpen     InboxThreadStatus = "open"
	InboxThreadStatusAssigned InboxThreadStatus = "assigned"
	InboxThreadStatusResolved InboxThreadStatus = "resolved"
)

// InboxItem is a normalized direct message, comment, or reply.
type InboxItem struct {
	ID                    string            `json:"id"`
	SocialAccountID       string            `json:"social_account_id"`
	WorkspaceID           string            `json:"workspace_id"`
	Source                InboxSource       `json:"source"`
	ExternalID            string            `json:"external_id"`
	ThreadKey             string            `json:"thread_key"`
	ThreadStatus          InboxThreadStatus `json:"thread_status"`
	IsRead                bool              `json:"is_read"`
	IsOwn                 bool              `json:"is_own"`
	ReceivedAt            string            `json:"received_at"`
	CreatedAt             string            `json:"created_at"`
	ParentExternalID      *string           `json:"parent_external_id,omitempty"`
	AssignedTo            *string           `json:"assigned_to,omitempty"`
	LinkedPostID          *string           `json:"linked_post_id,omitempty"`
	AuthorName            *string           `json:"author_name,omitempty"`
	AuthorID              *string           `json:"author_id,omitempty"`
	AuthorAvatarURL       *string           `json:"author_avatar_url,omitempty"`
	Body                  *string           `json:"body,omitempty"`
	AccountName           *string           `json:"account_name,omitempty"`
	AccountPlatform       *string           `json:"account_platform,omitempty"`
	AccountAvatarURL      *string           `json:"account_avatar_url,omitempty"`
	XCreditsCounted       *int              `json:"x_credits_counted,omitempty"`
	XCreditOperation      *string           `json:"x_credit_operation,omitempty"`
	XCreditCatalogVersion *string           `json:"x_credit_catalog_version,omitempty"`
	XCreditBillingMode    *string           `json:"x_credit_billing_mode,omitempty"`
	URL                   *string           `json:"url,omitempty"`
}

// InboxListParams are the optional filters for one Inbox list request.
type InboxListParams struct {
	Source InboxSource
	IsRead *bool
	IsOwn  *bool
	Limit  int
}

// InboxListResponse is the non-paginated Inbox list response.
type InboxListResponse struct {
	Data      []InboxItem `json:"data"`
	RequestID *string     `json:"request_id,omitempty"`
}

// InboxReplyRequest is the body of an Inbox reply request.
type InboxReplyRequest struct {
	Text string `json:"text"`
}

// InboxReplyState identifies whether a reply completed or requires polling.
type InboxReplyState string

const (
	InboxReplyStateCompleted   InboxReplyState = "completed"
	InboxReplyStateReconciling InboxReplyState = "reconciling"
)

const inboxReplyReconcilingCode = "X_REMOTE_ACCEPTED_RECONCILING"

var errInvalidInboxReplyResponse = errors.New("unipost: invalid Inbox reply response")

// InboxReplyResult is the response-aware result of an Inbox reply.
//
// When State is InboxReplyStateCompleted, Item is non-nil and OperationID is
// optional. When State is InboxReplyStateReconciling, Item is nil and
// OperationID, Code, and Message are non-empty; RequestID is optional.
type InboxReplyResult struct {
	State       InboxReplyState
	Item        *InboxItem
	OperationID string
	Code        string
	Message     string
	RequestID   *string
}

type inboxScopeKind string

const (
	inboxScopeManagedUser inboxScopeKind = "managed_user"
	inboxScopeWorkspace   inboxScopeKind = "workspace"
)

type inboxScope struct {
	kind           inboxScopeKind
	externalUserID string
}

func (s inboxScope) query() (url.Values, error) {
	query := make(url.Values, 2)
	query.Set("inbox_scope", string(s.kind))
	if s.kind == inboxScopeManagedUser {
		if strings.TrimSpace(s.externalUserID) == "" {
			return nil, fmt.Errorf("unipost: managed user external ID is required")
		}
		query.Set("external_user_id", s.externalUserID)
	}
	return query, nil
}

// InboxService creates Inbox resources bound to an explicit access scope.
type InboxService struct {
	client *Client
}

// ManagedUser returns an Inbox resource bound to one managed user's stable external ID.
// A blank ID is rejected locally when an operation is attempted.
func (s *InboxService) ManagedUser(externalUserID string) *ScopedInboxService {
	return &ScopedInboxService{
		client: s.client,
		scope: inboxScope{
			kind:           inboxScopeManagedUser,
			externalUserID: externalUserID,
		},
	}
}

// Workspace returns an aggregate Inbox resource for workspace owners and admins.
func (s *InboxService) Workspace() *ScopedInboxService {
	return &ScopedInboxService{
		client: s.client,
		scope:  inboxScope{kind: inboxScopeWorkspace},
	}
}

// ScopedInboxService handles Inbox operations within its bound scope.
type ScopedInboxService struct {
	client *Client
	scope  inboxScope
}

// List returns one limit-only Inbox collection response for the bound scope.
func (s *ScopedInboxService) List(ctx context.Context, params *InboxListParams) (*InboxListResponse, error) {
	values, err := s.scope.query()
	if err != nil {
		return nil, err
	}
	if params != nil {
		if params.Source != "" {
			values.Set("source", string(params.Source))
		}
		if params.IsRead != nil {
			values.Set("is_read", strconv.FormatBool(*params.IsRead))
		}
		if params.IsOwn != nil {
			values.Set("is_own", strconv.FormatBool(*params.IsOwn))
		}
		if params.Limit != 0 {
			values.Set("limit", strconv.Itoa(params.Limit))
		}
	}

	query := make(map[string]string, len(values))
	for key := range values {
		query[key] = values.Get(key)
	}

	var result InboxListResponse
	if err := s.client.do(ctx, http.MethodGet, "/v1/inbox", query, nil, &result, nil); err != nil {
		return nil, err
	}
	return &result, nil
}

// Reply sends one reply in the bound scope. It never automatically retries or
// follows redirects because doing so could duplicate an external write.
func (s *ScopedInboxService) Reply(ctx context.Context, id string, request *InboxReplyRequest, opts ...RequestOption) (*InboxReplyResult, error) {
	escapedID, err := inboxPathID(id)
	if err != nil {
		return nil, err
	}
	if request == nil {
		return nil, fmt.Errorf("unipost: Inbox reply request is required")
	}
	values, err := s.scope.query()
	if err != nil {
		return nil, err
	}
	query := make(map[string]string, len(values))
	for key := range values {
		query[key] = values.Get(key)
	}

	headers := make(map[string]string, 1)
	options := collectRequestOptions(opts)
	if options.idempotencyKey != "" {
		headers["Idempotency-Key"] = options.idempotencyKey
	}
	response, err := s.client.doResponseOnce(
		ctx,
		http.MethodPost,
		"/v1/inbox/"+escapedID+"/reply",
		query,
		request,
		headers,
	)
	if err != nil {
		return nil, err
	}

	switch response.StatusCode {
	case http.StatusOK:
		return decodeInboxReplyCompleted(response)
	case http.StatusAccepted:
		return decodeInboxReplyReconciling(response)
	default:
		if response.StatusCode >= 200 && response.StatusCode < 300 {
			return nil, errInvalidInboxReplyResponse
		}
		return nil, parseAPIError(response.StatusCode, response.Body)
	}
}

func inboxPathID(id string) (string, error) {
	trimmed := strings.TrimSpace(id)
	if trimmed == "" || trimmed == "." || trimmed == ".." {
		return "", fmt.Errorf("unipost: Inbox item ID is required and must be safe")
	}
	return url.PathEscape(id), nil
}

func decodeInboxReplyCompleted(response *responseAwareHTTPResult) (*InboxReplyResult, error) {
	var envelope struct {
		Data json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(response.Body, &envelope); err != nil || len(envelope.Data) == 0 || string(envelope.Data) == "null" {
		return nil, errInvalidInboxReplyResponse
	}

	var fields map[string]json.RawMessage
	if err := json.Unmarshal(envelope.Data, &fields); err != nil {
		return nil, errInvalidInboxReplyResponse
	}
	requiredFields := [...]string{
		"id",
		"social_account_id",
		"workspace_id",
		"source",
		"external_id",
		"thread_key",
		"thread_status",
		"is_read",
		"is_own",
		"received_at",
		"created_at",
	}
	for _, field := range requiredFields {
		value, ok := fields[field]
		if !ok || string(value) == "null" {
			return nil, errInvalidInboxReplyResponse
		}
	}

	var item InboxItem
	if err := json.Unmarshal(envelope.Data, &item); err != nil {
		return nil, errInvalidInboxReplyResponse
	}
	return &InboxReplyResult{
		State:       InboxReplyStateCompleted,
		Item:        &item,
		OperationID: strings.TrimSpace(response.Header.Get("X-UniPost-Operation-Id")),
	}, nil
}

func decodeInboxReplyReconciling(response *responseAwareHTTPResult) (*InboxReplyResult, error) {
	var envelope struct {
		Error *struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
		RequestID *string `json:"request_id,omitempty"`
	}
	if err := json.Unmarshal(response.Body, &envelope); err != nil || envelope.Error == nil {
		return nil, errInvalidInboxReplyResponse
	}
	operationID := strings.TrimSpace(response.Header.Get("X-UniPost-Operation-Id"))
	if envelope.Error.Code != inboxReplyReconcilingCode || strings.TrimSpace(envelope.Error.Message) == "" || operationID == "" {
		return nil, errInvalidInboxReplyResponse
	}
	return &InboxReplyResult{
		State:       InboxReplyStateReconciling,
		OperationID: operationID,
		Code:        inboxReplyReconcilingCode,
		Message:     envelope.Error.Message,
		RequestID:   envelope.RequestID,
	}, nil
}
