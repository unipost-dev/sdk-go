package unipost

import (
	"context"
	"net/http"
	"time"
)

// ConnectSession represents an OAuth Connect session.
type ConnectSession struct {
	ID                       string     `json:"id"`
	URL                      string     `json:"url"`
	Status                   string     `json:"status"` // "pending" | "completed" | "expired"
	ExpiresAt                time.Time  `json:"expires_at"`
	Platform                 string     `json:"platform"`
	ExternalUserID           string     `json:"external_user_id"`
	ExternalUserEmail        string     `json:"external_user_email,omitempty"`
	ReturnURL                string     `json:"return_url,omitempty"`
	CreatedAt                time.Time  `json:"created_at"`
	CompletedAt              *time.Time `json:"completed_at,omitempty"`
	CompletedSocialAccountID string     `json:"completed_social_account_id,omitempty"`
}

// CreateConnectSessionParams configures a Connect session.
type CreateConnectSessionParams struct {
	Platform          string `json:"platform"`
	ProfileID         string `json:"profile_id,omitempty"`
	ExternalUserID    string `json:"external_user_id"`
	ExternalUserEmail string `json:"external_user_email,omitempty"`
	ReturnURL         string `json:"return_url,omitempty"`
}

// ConnectService handles Connect (managed OAuth) operations.
type ConnectService struct {
	client *Client
}

func (s *ConnectService) CreateSession(ctx context.Context, params *CreateConnectSessionParams) (*ConnectSession, error) {
	var env apiEnvelope[ConnectSession]
	if err := s.client.do(ctx, http.MethodPost, "/v1/connect/sessions", nil, params, &env, nil); err != nil {
		return nil, err
	}
	return &env.Data, nil
}

func (s *ConnectService) GetSession(ctx context.Context, sessionID string) (*ConnectSession, error) {
	var env apiEnvelope[ConnectSession]
	if err := s.client.do(ctx, http.MethodGet, "/v1/connect/sessions/"+sessionID, nil, nil, &env, nil); err != nil {
		return nil, err
	}
	return &env.Data, nil
}
