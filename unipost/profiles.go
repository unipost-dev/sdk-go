package unipost

import (
	"context"
	"encoding/json"
)

// ProfilesService handles profile operations.
type ProfilesService struct {
	http *httpClient
}

// List returns all profiles.
func (s *ProfilesService) List(ctx context.Context) (*ListProfilesResponse, error) {
	data, err := s.http.get(ctx, "/v1/profiles", nil)
	if err != nil {
		return nil, err
	}

	var resp ListProfilesResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// Get returns a single profile by ID.
func (s *ProfilesService) Get(ctx context.Context, profileID string) (*Profile, error) {
	data, err := s.http.get(ctx, "/v1/profiles/"+profileID, nil)
	if err != nil {
		return nil, err
	}

	var resp dataEnvelope[Profile]
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}
	return &resp.Data, nil
}
