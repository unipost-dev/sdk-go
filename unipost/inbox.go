package unipost

import (
	"bytes"
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

var (
	errInvalidInboxReplyResponse = errors.New("unipost: invalid Inbox reply response")
	errInvalidInboxResponse      = errors.New("unipost: invalid Inbox response")
)

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

// InboxUnreadCountResult is the unread item count for the bound Inbox scope.
type InboxUnreadCountResult struct {
	Count int `json:"count"`
}

// InboxMarkAllReadResult reports how many Inbox items were marked read.
type InboxMarkAllReadResult struct {
	Marked int `json:"marked"`
}

// InboxThreadStateRequest updates an Inbox thread's workflow state.
type InboxThreadStateRequest struct {
	ThreadStatus InboxThreadStatus `json:"thread_status"`
	AssignedTo   *string           `json:"assigned_to,omitempty"`
}

// InboxMediaContext contains the media metadata associated with an Inbox item.
type InboxMediaContext struct {
	ID        string `json:"id"`
	Caption   string `json:"caption"`
	MediaURL  string `json:"media_url"`
	Timestamp string `json:"timestamp"`
	MediaType string `json:"media_type"`
	Permalink string `json:"permalink"`
}

// XInboxBackfillRequest configures a metered X Inbox backfill.
// IncludeReplies and IncludeDMs are always serialized, including when false.
type XInboxBackfillRequest struct {
	AccountID         *string `json:"account_id,omitempty"`
	LookbackDays      *int    `json:"lookback_days,omitempty"`
	MaxItems          *int    `json:"max_items,omitempty"`
	IncludeReplies    bool    `json:"include_replies"`
	IncludeDMs        bool    `json:"include_dms"`
	ConfirmationToken *string `json:"confirmation_token,omitempty"`
}

// InboxSyncError describes one account or platform step that failed during sync.
type InboxSyncError struct {
	AccountID string `json:"account_id"`
	Platform  string `json:"platform"`
	Step      string `json:"step"`
	Error     string `json:"error"`
}

// InboxSyncAccountDetail describes the items discovered for one account.
type InboxSyncAccountDetail struct {
	AccountID     string `json:"account_id"`
	Platform      string `json:"platform"`
	AccountName   string `json:"account_name"`
	MediaFound    int    `json:"media_found"`
	CommentsFound int    `json:"comments_found"`
}

// InboxSyncResult is the result of an ordinary Inbox sync.
type InboxSyncResult struct {
	NewItems        int                      `json:"new_items"`
	AccountsChecked int                      `json:"accounts_checked"`
	Errors          []InboxSyncError         `json:"errors"`
	Details         []InboxSyncAccountDetail `json:"details"`
}

// XInboxBackfillAccountResult contains X backfill counts for one account.
type XInboxBackfillAccountResult struct {
	AccountID         string   `json:"account_id"`
	Accepted          int      `json:"accepted"`
	Suppressed        int      `json:"suppressed"`
	Duplicates        int      `json:"duplicates"`
	Read              int      `json:"read"`
	StoppedAtBoundary *bool    `json:"stopped_at_boundary,omitempty"`
	StopReason        *string  `json:"stop_reason,omitempty"`
	MissingScopes     []string `json:"missing_scopes,omitempty"`
}

// XInboxBackfillResult is the closed result set for an X Inbox backfill.
type XInboxBackfillResult interface {
	isXInboxBackfillResult()
}

// XInboxBackfillInProgress indicates that a confirmed backfill is executing.
type XInboxBackfillInProgress struct {
	Status                  string                        `json:"status"`
	ConfirmationOperationID string                        `json:"confirmation_operation_id"`
	ExecutionLeaseExpiresAt string                        `json:"execution_lease_expires_at"`
	EstimatedXCredits       *int                          `json:"estimated_x_credits,omitempty"`
	ConfirmationRequired    *bool                         `json:"confirmation_required,omitempty"`
	ConfirmationToken       *string                       `json:"confirmation_token,omitempty"`
	ConfirmationExpiresAt   *string                       `json:"confirmation_expires_at,omitempty"`
	AccountsChecked         *int                          `json:"accounts_checked,omitempty"`
	Accepted                *int                          `json:"accepted,omitempty"`
	Suppressed              *int                          `json:"suppressed,omitempty"`
	Duplicates              *int                          `json:"duplicates,omitempty"`
	Read                    *int                          `json:"read,omitempty"`
	Details                 []XInboxBackfillAccountResult `json:"details,omitempty"`
}

func (*XInboxBackfillInProgress) isXInboxBackfillResult() {}

// XInboxBackfillConfirmationRequired contains the estimate and confirmation
// token required before a metered backfill starts.
type XInboxBackfillConfirmationRequired struct {
	ConfirmationRequired    bool                          `json:"confirmation_required"`
	ConfirmationToken       string                        `json:"confirmation_token"`
	ConfirmationExpiresAt   string                        `json:"confirmation_expires_at"`
	AccountsChecked         int                           `json:"accounts_checked"`
	EstimatedXCredits       *int                          `json:"estimated_x_credits,omitempty"`
	ConfirmationOperationID *string                       `json:"confirmation_operation_id,omitempty"`
	ExecutionLeaseExpiresAt *string                       `json:"execution_lease_expires_at,omitempty"`
	Accepted                *int                          `json:"accepted,omitempty"`
	Suppressed              *int                          `json:"suppressed,omitempty"`
	Duplicates              *int                          `json:"duplicates,omitempty"`
	Read                    *int                          `json:"read,omitempty"`
	Details                 []XInboxBackfillAccountResult `json:"details,omitempty"`
}

func (*XInboxBackfillConfirmationRequired) isXInboxBackfillResult() {}

// XInboxBackfillCompleted contains the final metered backfill counts.
type XInboxBackfillCompleted struct {
	ConfirmationRequired    bool                          `json:"confirmation_required"`
	AccountsChecked         int                           `json:"accounts_checked"`
	Accepted                int                           `json:"accepted"`
	Suppressed              int                           `json:"suppressed"`
	Duplicates              int                           `json:"duplicates"`
	Read                    int                           `json:"read"`
	EstimatedXCredits       *int                          `json:"estimated_x_credits,omitempty"`
	ConfirmationOperationID *string                       `json:"confirmation_operation_id,omitempty"`
	ConfirmationToken       *string                       `json:"confirmation_token,omitempty"`
	ConfirmationExpiresAt   *string                       `json:"confirmation_expires_at,omitempty"`
	ExecutionLeaseExpiresAt *string                       `json:"execution_lease_expires_at,omitempty"`
	Details                 []XInboxBackfillAccountResult `json:"details,omitempty"`
}

func (*XInboxBackfillCompleted) isXInboxBackfillResult() {}

// XInboxOutboundStatus describes reconciliation state for an X Inbox write.
type XInboxOutboundStatus struct {
	ID                     string  `json:"id"`
	Status                 string  `json:"status"`
	CompletionAttempts     int     `json:"completion_attempts"`
	ReconciliationDeadline *string `json:"reconciliation_deadline,omitempty"`
	ReconciliationRequired bool    `json:"reconciliation_required"`
	ResponseInboxItemID    *string `json:"response_inbox_item_id,omitempty"`
	UpdatedAt              string  `json:"updated_at"`
}

// InboxWebSocketConnectionDetails contains local-only WebSocket connection
// details. Headers is a fresh map on every call.
type InboxWebSocketConnectionDetails struct {
	URL     string
	Headers map[string]string
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
// A blank ID is rejected locally before a scoped resource is returned.
func (s *InboxService) ManagedUser(externalUserID string) (*ScopedInboxService, error) {
	if strings.TrimSpace(externalUserID) == "" {
		return nil, fmt.Errorf("unipost: managed user external ID is required")
	}
	return &ScopedInboxService{
		client: s.client,
		scope: inboxScope{
			kind:           inboxScopeManagedUser,
			externalUserID: externalUserID,
		},
	}, nil
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

func (s *ScopedInboxService) scopeQueryMap() (map[string]string, error) {
	values, err := s.scope.query()
	if err != nil {
		return nil, err
	}
	query := make(map[string]string, len(values))
	for key := range values {
		query[key] = values.Get(key)
	}
	return query, nil
}

func (s *ScopedInboxService) getInboxResponse(ctx context.Context, path string) ([]byte, error) {
	query, err := s.scopeQueryMap()
	if err != nil {
		return nil, err
	}
	var response json.RawMessage
	if err := s.client.do(ctx, http.MethodGet, path, query, nil, &response, nil); err != nil {
		return nil, err
	}
	return response, nil
}

// postInboxOnce performs one non-replayable Inbox POST. It deliberately uses
// the response-aware single-attempt transport so it neither retries a 429 nor
// follows a redirect.
func (s *ScopedInboxService) postInboxOnce(ctx context.Context, path string, body any) ([]byte, error) {
	query, err := s.scopeQueryMap()
	if err != nil {
		return nil, err
	}
	response, err := s.client.doResponseOnce(ctx, http.MethodPost, path, query, body, nil)
	if err != nil {
		return nil, err
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return nil, parseAPIError(response.StatusCode, response.Body)
	}
	return response.Body, nil
}

// UnreadCount returns the unread count for the bound Inbox scope.
func (s *ScopedInboxService) UnreadCount(ctx context.Context) (*InboxUnreadCountResult, error) {
	body, err := s.getInboxResponse(ctx, "/v1/inbox/unread-count")
	if err != nil {
		return nil, err
	}
	var result InboxUnreadCountResult
	if err := decodeInboxDataEnvelope(body, []string{"count"}, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Get returns one Inbox item from the bound scope.
func (s *ScopedInboxService) Get(ctx context.Context, id string) (*InboxItem, error) {
	escapedID, err := inboxPathID(id)
	if err != nil {
		return nil, err
	}
	body, err := s.getInboxResponse(ctx, "/v1/inbox/"+escapedID)
	if err != nil {
		return nil, err
	}
	return decodeInboxItemEnvelope(body)
}

// MarkRead marks one Inbox item read. It performs exactly one request.
func (s *ScopedInboxService) MarkRead(ctx context.Context, id string) error {
	escapedID, err := inboxPathID(id)
	if err != nil {
		return err
	}
	_, err = s.postInboxOnce(ctx, "/v1/inbox/"+escapedID+"/read", nil)
	return err
}

// MarkAllRead marks every Inbox item in the bound scope read.
func (s *ScopedInboxService) MarkAllRead(ctx context.Context) (*InboxMarkAllReadResult, error) {
	body, err := s.postInboxOnce(ctx, "/v1/inbox/mark-all-read", nil)
	if err != nil {
		return nil, err
	}
	var result InboxMarkAllReadResult
	if err := decodeInboxDataEnvelope(body, []string{"marked"}, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// UpdateThreadState updates one Inbox thread's workflow state.
func (s *ScopedInboxService) UpdateThreadState(ctx context.Context, id string, request *InboxThreadStateRequest) (*InboxItem, error) {
	escapedID, err := inboxPathID(id)
	if err != nil {
		return nil, err
	}
	if request == nil {
		return nil, fmt.Errorf("unipost: Inbox thread state request is required")
	}
	switch request.ThreadStatus {
	case InboxThreadStatusOpen, InboxThreadStatusAssigned, InboxThreadStatusResolved:
	default:
		return nil, fmt.Errorf("unipost: invalid Inbox thread status")
	}
	body, err := s.postInboxOnce(ctx, "/v1/inbox/"+escapedID+"/thread-state", request)
	if err != nil {
		return nil, err
	}
	return decodeInboxItemEnvelope(body)
}

// MediaContext returns media metadata associated with one Inbox item.
func (s *ScopedInboxService) MediaContext(ctx context.Context, id string) (*InboxMediaContext, error) {
	escapedID, err := inboxPathID(id)
	if err != nil {
		return nil, err
	}
	body, err := s.getInboxResponse(ctx, "/v1/inbox/"+escapedID+"/media-context")
	if err != nil {
		return nil, err
	}
	var result InboxMediaContext
	if err := decodeInboxDataEnvelope(body, []string{"id", "caption", "media_url", "timestamp", "media_type", "permalink"}, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Sync performs an ordinary Inbox sync without a request body.
func (s *ScopedInboxService) Sync(ctx context.Context) (*InboxSyncResult, error) {
	body, err := s.postInboxOnce(ctx, "/v1/inbox/sync", nil)
	if err != nil {
		return nil, err
	}
	return decodeInboxSyncEnvelope(body)
}

// SyncXBackfill performs a metered X Inbox backfill using the same sync route.
// A separate method preserves static result typing in Go.
func (s *ScopedInboxService) SyncXBackfill(ctx context.Context, request *XInboxBackfillRequest) (XInboxBackfillResult, error) {
	if request == nil {
		return nil, fmt.Errorf("unipost: X Inbox backfill request is required")
	}
	body, err := s.postInboxOnce(ctx, "/v1/inbox/sync", struct {
		XBackfill *XInboxBackfillRequest `json:"x_backfill"`
	}{XBackfill: request})
	if err != nil {
		return nil, err
	}
	return decodeXInboxBackfillEnvelope(body)
}

// XOutboundStatus returns the reconciliation status for one X Inbox write.
func (s *ScopedInboxService) XOutboundStatus(ctx context.Context, requestID string) (*XInboxOutboundStatus, error) {
	escapedID, err := inboxPathID(requestID)
	if err != nil {
		return nil, err
	}
	body, err := s.getInboxResponse(ctx, "/v1/inbox/x-outbound-operations/"+escapedID)
	if err != nil {
		return nil, err
	}
	var result XInboxOutboundStatus
	if err := decodeInboxDataEnvelope(body, []string{"id", "status", "completion_attempts", "reconciliation_required", "updated_at"}, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// WebSocketConnectionDetails builds local connection details for the bound
// scope. It never performs a network request and never places the API key in
// the URL.
func (s *ScopedInboxService) WebSocketConnectionDetails() (*InboxWebSocketConnectionDetails, error) {
	query, err := s.scope.query()
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(s.client.apiKey) == "" {
		return nil, errors.New("unipost: API key is required")
	}
	base, err := url.Parse(s.client.baseURL)
	if err != nil || base.Scheme == "" || base.Host == "" || base.User != nil || base.Opaque != "" {
		return nil, errors.New("unipost: invalid WebSocket base URL")
	}
	if base.Scheme != "http" && base.Scheme != "https" {
		return nil, errors.New("unipost: invalid WebSocket base URL")
	}
	hostname := base.Hostname()
	if hostname == "" || strings.ContainsAny(hostname, " 	\r\n/\\?#") {
		return nil, errors.New("unipost: invalid WebSocket base URL")
	}
	// Calling Port also validates malformed numeric port syntax on supported
	// Go versions. url.Parse has already rejected most invalid forms.
	_ = base.Port()

	scheme := "ws"
	if base.Scheme == "https" {
		scheme = "wss"
	}
	connectionURL := (&url.URL{
		Scheme:   scheme,
		Host:     base.Host,
		Path:     "/v1/inbox/ws",
		RawQuery: query.Encode(),
	}).String()
	return &InboxWebSocketConnectionDetails{
		URL: connectionURL,
		Headers: map[string]string{
			"Authorization": "Bearer " + s.client.apiKey,
		},
	}, nil
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
		return nil, parseInboxReplyAPIError(response.StatusCode, response.Body)
	}
}

func parseInboxReplyAPIError(status int, body []byte) error {
	err := parseAPIError(status, body)
	apiErr, ok := err.(*APIError)
	if !ok {
		return err
	}

	var envelope struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if json.Unmarshal(body, &envelope) == nil && strings.TrimSpace(envelope.Error.Code) != "" {
		apiErr.Code = envelope.Error.Code
	}
	return apiErr
}

func inboxPathID(id string) (string, error) {
	trimmed := strings.TrimSpace(id)
	if trimmed == "" || trimmed == "." || trimmed == ".." {
		return "", fmt.Errorf("unipost: Inbox item ID is required and must be safe")
	}
	return url.PathEscape(id), nil
}

func inboxEnvelopeData(body []byte) (json.RawMessage, map[string]json.RawMessage, error) {
	var envelope struct {
		Data json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(body, &envelope); err != nil || len(envelope.Data) == 0 || isJSONNull(envelope.Data) {
		return nil, nil, errInvalidInboxResponse
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(envelope.Data, &fields); err != nil || fields == nil {
		return nil, nil, errInvalidInboxResponse
	}
	return envelope.Data, fields, nil
}

func isJSONNull(value json.RawMessage) bool {
	return string(bytes.TrimSpace(value)) == "null"
}

func requireInboxFields(fields map[string]json.RawMessage, required ...string) error {
	for _, field := range required {
		value, ok := fields[field]
		if !ok || len(value) == 0 || isJSONNull(value) {
			return errInvalidInboxResponse
		}
	}
	return nil
}

func decodeInboxDataEnvelope(body []byte, required []string, out any) error {
	data, fields, err := inboxEnvelopeData(body)
	if err != nil {
		return errInvalidInboxResponse
	}
	if err := requireInboxFields(fields, required...); err != nil {
		return errInvalidInboxResponse
	}
	if err := json.Unmarshal(data, out); err != nil {
		return errInvalidInboxResponse
	}
	return nil
}

var inboxItemRequiredFields = []string{
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

func decodeInboxItemEnvelope(body []byte) (*InboxItem, error) {
	var item InboxItem
	if err := decodeInboxDataEnvelope(body, inboxItemRequiredFields, &item); err != nil {
		return nil, errInvalidInboxResponse
	}
	return &item, nil
}

func validateInboxObjectArray(fields map[string]json.RawMessage, field string, required ...string) error {
	raw, ok := fields[field]
	if !ok || isJSONNull(raw) {
		return errInvalidInboxResponse
	}
	var objects []json.RawMessage
	if err := json.Unmarshal(raw, &objects); err != nil || objects == nil {
		return errInvalidInboxResponse
	}
	for _, object := range objects {
		var objectFields map[string]json.RawMessage
		if err := json.Unmarshal(object, &objectFields); err != nil || objectFields == nil {
			return errInvalidInboxResponse
		}
		if err := requireInboxFields(objectFields, required...); err != nil {
			return errInvalidInboxResponse
		}
	}
	return nil
}

func decodeInboxSyncEnvelope(body []byte) (*InboxSyncResult, error) {
	data, fields, err := inboxEnvelopeData(body)
	if err != nil {
		return nil, errInvalidInboxResponse
	}
	if err := requireInboxFields(fields, "new_items", "accounts_checked", "errors", "details"); err != nil {
		return nil, errInvalidInboxResponse
	}
	if err := validateInboxObjectArray(fields, "errors", "account_id", "platform", "step", "error"); err != nil {
		return nil, errInvalidInboxResponse
	}
	if err := validateInboxObjectArray(fields, "details", "account_id", "platform", "account_name", "media_found", "comments_found"); err != nil {
		return nil, errInvalidInboxResponse
	}
	var result InboxSyncResult
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, errInvalidInboxResponse
	}
	return &result, nil
}

func validateOptionalXBackfillDetails(fields map[string]json.RawMessage) error {
	raw, ok := fields["details"]
	if !ok || isJSONNull(raw) {
		return nil
	}
	var details []json.RawMessage
	if err := json.Unmarshal(raw, &details); err != nil {
		return errInvalidInboxResponse
	}
	for _, detail := range details {
		var detailFields map[string]json.RawMessage
		if err := json.Unmarshal(detail, &detailFields); err != nil || detailFields == nil {
			return errInvalidInboxResponse
		}
		if err := requireInboxFields(detailFields, "account_id", "accepted", "suppressed", "duplicates", "read"); err != nil {
			return errInvalidInboxResponse
		}
	}
	return nil
}

func decodeXInboxBackfillEnvelope(body []byte) (XInboxBackfillResult, error) {
	data, fields, err := inboxEnvelopeData(body)
	if err != nil {
		return nil, errInvalidInboxResponse
	}
	if err := validateOptionalXBackfillDetails(fields); err != nil {
		return nil, errInvalidInboxResponse
	}

	if rawStatus, hasStatus := fields["status"]; hasStatus {
		var status string
		if isJSONNull(rawStatus) || json.Unmarshal(rawStatus, &status) != nil || status != "in_progress" {
			return nil, errInvalidInboxResponse
		}
		if err := requireInboxFields(fields, "confirmation_operation_id", "execution_lease_expires_at"); err != nil {
			return nil, errInvalidInboxResponse
		}
		if rawConfirmation, ok := fields["confirmation_required"]; ok && !isJSONNull(rawConfirmation) {
			var confirmationRequired bool
			if json.Unmarshal(rawConfirmation, &confirmationRequired) != nil || confirmationRequired {
				return nil, errInvalidInboxResponse
			}
		}
		var result XInboxBackfillInProgress
		if err := json.Unmarshal(data, &result); err != nil || strings.TrimSpace(result.ConfirmationOperationID) == "" || strings.TrimSpace(result.ExecutionLeaseExpiresAt) == "" {
			return nil, errInvalidInboxResponse
		}
		return &result, nil
	}

	rawConfirmation, ok := fields["confirmation_required"]
	if !ok || isJSONNull(rawConfirmation) {
		return nil, errInvalidInboxResponse
	}
	var confirmationRequired bool
	if err := json.Unmarshal(rawConfirmation, &confirmationRequired); err != nil {
		return nil, errInvalidInboxResponse
	}
	if confirmationRequired {
		if err := requireInboxFields(fields, "confirmation_token", "confirmation_expires_at", "accounts_checked"); err != nil {
			return nil, errInvalidInboxResponse
		}
		var result XInboxBackfillConfirmationRequired
		if err := json.Unmarshal(data, &result); err != nil || !result.ConfirmationRequired || strings.TrimSpace(result.ConfirmationToken) == "" || strings.TrimSpace(result.ConfirmationExpiresAt) == "" {
			return nil, errInvalidInboxResponse
		}
		return &result, nil
	}

	if err := requireInboxFields(fields, "accounts_checked", "accepted", "suppressed", "duplicates", "read"); err != nil {
		return nil, errInvalidInboxResponse
	}
	var result XInboxBackfillCompleted
	if err := json.Unmarshal(data, &result); err != nil || result.ConfirmationRequired {
		return nil, errInvalidInboxResponse
	}
	return &result, nil
}

func decodeInboxReplyCompleted(response *responseAwareHTTPResult) (*InboxReplyResult, error) {
	item, err := decodeInboxItemEnvelope(response.Body)
	if err != nil {
		return nil, errInvalidInboxReplyResponse
	}
	return &InboxReplyResult{
		State:       InboxReplyStateCompleted,
		Item:        item,
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
