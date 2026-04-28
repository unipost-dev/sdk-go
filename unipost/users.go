package unipost

import (
	"context"
	"net/http"
	"net/url"
)

// ManagedUser represents a user managed through Connect.
type ManagedUser struct {
	ExternalUserID    string         `json:"external_user_id"`
	ExternalUserEmail string         `json:"external_user_email,omitempty"`
	AccountCount      int            `json:"account_count"`
	PlatformCounts    map[string]int `json:"platform_counts,omitempty"`
	ReconnectCount    int            `json:"reconnect_count"`
}

// PaginatedManagedUsers wraps a list response.
type PaginatedManagedUsers struct {
	Data []ManagedUser
	Meta PageMeta
}

// UsersService handles managed-user operations.
type UsersService struct {
	client *Client
}

// List returns managed users (data slice only).
func (s *UsersService) List(ctx context.Context) ([]ManagedUser, error) {
	page, err := s.ListPage(ctx)
	if err != nil {
		return nil, err
	}
	return page.Data, nil
}

// ListPage returns managed users with pagination metadata.
func (s *UsersService) ListPage(ctx context.Context) (*PaginatedManagedUsers, error) {
	var env apiEnvelope[[]ManagedUser]
	if err := s.client.do(ctx, http.MethodGet, "/v1/users", nil, nil, &env, nil); err != nil {
		return nil, err
	}
	return &PaginatedManagedUsers{Data: env.Data, Meta: pageMetaFromEnvelope(env)}, nil
}

// Get returns a single managed user by external_user_id.
func (s *UsersService) Get(ctx context.Context, externalUserID string) (*ManagedUser, error) {
	var env apiEnvelope[ManagedUser]
	if err := s.client.do(ctx, http.MethodGet, "/v1/users/"+url.PathEscape(externalUserID), nil, nil, &env, nil); err != nil {
		return nil, err
	}
	return &env.Data, nil
}
