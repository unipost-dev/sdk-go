package unipost

import (
	"context"
	"net/http"
	"time"
)

// Platform represents a supported social media platform.
type Platform string

const (
	PlatformTwitter   Platform = "twitter"
	PlatformLinkedIn  Platform = "linkedin"
	PlatformInstagram Platform = "instagram"
	PlatformThreads   Platform = "threads"
	PlatformTikTok    Platform = "tiktok"
	PlatformYouTube   Platform = "youtube"
	PlatformBluesky   Platform = "bluesky"
)

// SocialAccount represents a connected social media account.
type SocialAccount struct {
	ID                string `json:"id"`
	ProfileID         string `json:"profile_id,omitempty"`
	ProfileName       string `json:"profile_name,omitempty"`
	Platform          string `json:"platform"`
	AccountName       string `json:"account_name,omitempty"`
	ExternalUserID    string `json:"external_user_id,omitempty"`
	ExternalUserEmail string `json:"external_user_email,omitempty"`
	Status            string `json:"status"`
	ConnectionType    string `json:"connection_type,omitempty"`
}

// AccountHealth represents the health status of a social account.
type AccountHealth struct {
	SocialAccountID      string     `json:"social_account_id"`
	Platform             string     `json:"platform"`
	Status               string     `json:"status"`
	LastSuccessfulPostAt *time.Time `json:"last_successful_post_at,omitempty"`
	TokenExpiresAt       *time.Time `json:"token_expires_at,omitempty"`
	LastError            JSONMap    `json:"last_error,omitempty"`
}

// ListAccountsParams are optional filters for listing accounts.
type ListAccountsParams struct {
	Platform       string
	ExternalUserID string
	Status         string
	ProfileID      string
}

// ConnectAccountParams configures a BYO-token account connection.
type ConnectAccountParams struct {
	ProfileID   string            `json:"profile_id,omitempty"`
	Platform    string            `json:"platform"`
	Credentials map[string]string `json:"credentials"`
}

// PaginatedAccounts wraps a list response with pagination metadata.
type PaginatedAccounts struct {
	Data []SocialAccount
	Meta PageMeta
}

// AccountsService handles social account operations.
type AccountsService struct {
	client *Client
}

// List returns connected social accounts (data slice only).
func (s *AccountsService) List(ctx context.Context, params *ListAccountsParams) ([]SocialAccount, error) {
	page, err := s.ListPage(ctx, params)
	if err != nil {
		return nil, err
	}
	return page.Data, nil
}

// ListPage returns connected social accounts with pagination metadata.
func (s *AccountsService) ListPage(ctx context.Context, params *ListAccountsParams) (*PaginatedAccounts, error) {
	query := map[string]string{}
	if params != nil {
		query["platform"] = params.Platform
		query["external_user_id"] = params.ExternalUserID
		query["status"] = params.Status
		query["profile_id"] = params.ProfileID
	}
	var env apiEnvelope[[]SocialAccount]
	if err := s.client.do(ctx, http.MethodGet, "/v1/accounts", query, nil, &env, nil); err != nil {
		return nil, err
	}
	return &PaginatedAccounts{Data: env.Data, Meta: pageMetaFromEnvelope(env)}, nil
}

// Get returns a single account by ID.
func (s *AccountsService) Get(ctx context.Context, accountID string) (*SocialAccount, error) {
	accounts, err := s.List(ctx, nil)
	if err != nil {
		return nil, err
	}
	for _, account := range accounts {
		if account.ID == accountID {
			a := account
			return &a, nil
		}
	}
	return nil, &APIError{Status: 404, Code: "not_found", NormalizedCode: "not_found", Message: "account not found"}
}

// Connect connects a BYO-token account.
func (s *AccountsService) Connect(ctx context.Context, params *ConnectAccountParams) (*SocialAccount, error) {
	var env apiEnvelope[SocialAccount]
	if err := s.client.do(ctx, http.MethodPost, "/v1/accounts/connect", nil, params, &env, nil); err != nil {
		return nil, err
	}
	return &env.Data, nil
}

// Disconnect removes a connected account.
func (s *AccountsService) Disconnect(ctx context.Context, accountID string) (JSONMap, error) {
	var env apiEnvelope[JSONMap]
	if err := s.client.do(ctx, http.MethodDelete, "/v1/accounts/"+accountID, nil, nil, &env, nil); err != nil {
		return nil, err
	}
	return env.Data, nil
}

// Capabilities returns the platform feature matrix for a connected account.
func (s *AccountsService) Capabilities(ctx context.Context, accountID string) (JSONMap, error) {
	var env apiEnvelope[JSONMap]
	if err := s.client.do(ctx, http.MethodGet, "/v1/accounts/"+accountID+"/capabilities", nil, nil, &env, nil); err != nil {
		return nil, err
	}
	return env.Data, nil
}

// Health returns the connection health for an account.
func (s *AccountsService) Health(ctx context.Context, accountID string) (*AccountHealth, error) {
	var env apiEnvelope[AccountHealth]
	if err := s.client.do(ctx, http.MethodGet, "/v1/accounts/"+accountID+"/health", nil, nil, &env, nil); err != nil {
		return nil, err
	}
	return &env.Data, nil
}

// TikTokCreatorInfo fetches TikTok creator info needed for posting.
func (s *AccountsService) TikTokCreatorInfo(ctx context.Context, accountID string) (JSONMap, error) {
	var env apiEnvelope[JSONMap]
	if err := s.client.do(ctx, http.MethodGet, "/v1/accounts/"+accountID+"/tiktok/creator-info", nil, nil, &env, nil); err != nil {
		return nil, err
	}
	return env.Data, nil
}

// FacebookPageInsights fetches Facebook page insights for an account.
func (s *AccountsService) FacebookPageInsights(ctx context.Context, accountID string) (JSONMap, error) {
	var env apiEnvelope[JSONMap]
	if err := s.client.do(ctx, http.MethodGet, "/v1/accounts/"+accountID+"/facebook/page-insights", nil, nil, &env, nil); err != nil {
		return nil, err
	}
	return env.Data, nil
}
