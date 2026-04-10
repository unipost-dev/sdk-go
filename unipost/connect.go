package unipost

import (
	"context"
	"encoding/json"
)

// ConnectService handles OAuth Connect sessions.
type ConnectService struct {
	http *httpClient
}

// CreateSession creates a Connect session for end-user OAuth.
func (s *ConnectService) CreateSession(ctx context.Context, params *CreateConnectSessionParams) (*ConnectSession, error) {
	data, err := s.http.post(ctx, "/v1/connect/sessions", params, nil)
	if err != nil {
		return nil, err
	}

	var resp dataEnvelope[ConnectSession]
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}
	return &resp.Data, nil
}

// GetSession returns the status of a Connect session.
func (s *ConnectService) GetSession(ctx context.Context, sessionID string) (*ConnectSession, error) {
	data, err := s.http.get(ctx, "/v1/connect/sessions/"+sessionID, nil)
	if err != nil {
		return nil, err
	}

	var resp dataEnvelope[ConnectSession]
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}
	return &resp.Data, nil
}
