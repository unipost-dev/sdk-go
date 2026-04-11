package unipost

import (
	"context"
	"encoding/json"
	"net/url"
)

// AccountsService handles social account operations.
type AccountsService struct {
	http *httpClient
}

// ListAccountsParams are optional filters for listing accounts.
type ListAccountsParams struct {
	Platform       string
	ExternalUserID string
	Status         string
	ProfileID      string
}

// List returns all connected social accounts.
func (s *AccountsService) List(ctx context.Context, params *ListAccountsParams) ([]SocialAccount, error) {
	q := url.Values{}
	if params != nil {
		if params.Platform != "" {
			q.Set("platform", params.Platform)
		}
		if params.ExternalUserID != "" {
			q.Set("external_user_id", params.ExternalUserID)
		}
		if params.Status != "" {
			q.Set("status", params.Status)
		}
		if params.ProfileID != "" {
			q.Set("profile_id", params.ProfileID)
		}
	}

	data, err := s.http.get(ctx, "/v1/social-accounts", q)
	if err != nil {
		return nil, err
	}

	var resp PaginatedResponse[SocialAccount]
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}
	return resp.Data, nil
}

// Get returns a single account by ID.
func (s *AccountsService) Get(ctx context.Context, accountID string) (*SocialAccount, error) {
	data, err := s.http.get(ctx, "/v1/social-accounts/"+accountID, nil)
	if err != nil {
		return nil, err
	}

	var resp dataEnvelope[SocialAccount]
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}
	return &resp.Data, nil
}

// Health returns the health status of an account.
func (s *AccountsService) Health(ctx context.Context, accountID string) (*AccountHealth, error) {
	data, err := s.http.get(ctx, "/v1/social-accounts/"+accountID+"/health", nil)
	if err != nil {
		return nil, err
	}

	var resp dataEnvelope[AccountHealth]
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}
	return &resp.Data, nil
}
