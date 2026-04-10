package unipost

import (
	"context"
	"encoding/json"
)

// UsersService handles managed user operations.
type UsersService struct {
	http *httpClient
}

// List returns all managed users.
func (s *UsersService) List(ctx context.Context) ([]ManagedUser, error) {
	data, err := s.http.get(ctx, "/v1/users", nil)
	if err != nil {
		return nil, err
	}

	var resp PaginatedResponse[ManagedUser]
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}
	return resp.Data, nil
}

// Get returns a single managed user by external_user_id.
func (s *UsersService) Get(ctx context.Context, externalUserID string) (*ManagedUser, error) {
	data, err := s.http.get(ctx, "/v1/users/"+externalUserID, nil)
	if err != nil {
		return nil, err
	}

	var resp dataEnvelope[ManagedUser]
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}
	return &resp.Data, nil
}
