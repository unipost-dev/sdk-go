package unipost

import (
	"context"
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
