package unipost

import (
	"context"
	"net/http"
	"time"
)

// PlatformCredential represents a stored OAuth app credential.
type PlatformCredential struct {
	Platform  string    `json:"platform"`
	ClientID  string    `json:"client_id"`
	CreatedAt time.Time `json:"created_at"`
}

// CreatePlatformCredentialParams configures a credential creation.
type CreatePlatformCredentialParams struct {
	Platform     string `json:"platform"`
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}

// PaginatedPlatformCredentials wraps a list response.
type PaginatedPlatformCredentials struct {
	Data []PlatformCredential
	Meta PageMeta
}

// PlatformCredentialsService handles BYO-OAuth credential operations.
type PlatformCredentialsService struct {
	client *Client
}

func (s *PlatformCredentialsService) Create(ctx context.Context, params *CreatePlatformCredentialParams) (*PlatformCredential, error) {
	var env apiEnvelope[PlatformCredential]
	if err := s.client.do(ctx, http.MethodPost, "/v1/platform-credentials", nil, params, &env, nil); err != nil {
		return nil, err
	}
	return &env.Data, nil
}

func (s *PlatformCredentialsService) List(ctx context.Context) (*PaginatedPlatformCredentials, error) {
	var env apiEnvelope[[]PlatformCredential]
	if err := s.client.do(ctx, http.MethodGet, "/v1/platform-credentials", nil, nil, &env, nil); err != nil {
		return nil, err
	}
	return &PaginatedPlatformCredentials{Data: env.Data, Meta: pageMetaFromEnvelope(env)}, nil
}

func (s *PlatformCredentialsService) Delete(ctx context.Context, platform string) error {
	return s.client.do(ctx, http.MethodDelete, "/v1/platform-credentials/"+platform, nil, nil, nil, nil)
}
