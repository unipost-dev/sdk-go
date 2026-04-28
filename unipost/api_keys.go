package unipost

import (
	"context"
	"net/http"
	"time"
)

// APIKey represents an API key record (without the secret).
type APIKey struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Prefix      string     `json:"prefix"`
	Environment string     `json:"environment"`
	CreatedAt   time.Time  `json:"created_at"`
	LastUsedAt  *time.Time `json:"last_used_at,omitempty"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
}

// CreatedAPIKey is returned only at creation time and includes the plaintext key.
type CreatedAPIKey struct {
	APIKey
	Key string `json:"key"`
}

// CreateAPIKeyParams configures an API key creation.
type CreateAPIKeyParams struct {
	Name        string `json:"name"`
	Environment string `json:"environment,omitempty"`
	ExpiresAt   string `json:"expires_at,omitempty"`
}

// PaginatedAPIKeys wraps a list response.
type PaginatedAPIKeys struct {
	Data []APIKey
	Meta PageMeta
}

// APIKeysService handles API key CRUD operations.
type APIKeysService struct {
	client *Client
}

func (s *APIKeysService) List(ctx context.Context) (*PaginatedAPIKeys, error) {
	var env apiEnvelope[[]APIKey]
	if err := s.client.do(ctx, http.MethodGet, "/v1/api-keys", nil, nil, &env, nil); err != nil {
		return nil, err
	}
	return &PaginatedAPIKeys{Data: env.Data, Meta: pageMetaFromEnvelope(env)}, nil
}

func (s *APIKeysService) Create(ctx context.Context, params *CreateAPIKeyParams) (*CreatedAPIKey, error) {
	var env apiEnvelope[CreatedAPIKey]
	if err := s.client.do(ctx, http.MethodPost, "/v1/api-keys", nil, params, &env, nil); err != nil {
		return nil, err
	}
	return &env.Data, nil
}

func (s *APIKeysService) Revoke(ctx context.Context, keyID string) error {
	return s.client.do(ctx, http.MethodDelete, "/v1/api-keys/"+keyID, nil, nil, nil, nil)
}
