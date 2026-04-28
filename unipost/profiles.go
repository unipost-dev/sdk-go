package unipost

import (
	"context"
	"net/http"
	"time"
)

// Profile represents a workspace profile.
type Profile struct {
	ID                   string    `json:"id"`
	WorkspaceID          string    `json:"workspace_id"`
	Name                 string    `json:"name"`
	AccountCount         int       `json:"account_count"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
	BrandingLogoURL      *string   `json:"branding_logo_url,omitempty"`
	BrandingDisplayName  *string   `json:"branding_display_name,omitempty"`
	BrandingPrimaryColor *string   `json:"branding_primary_color,omitempty"`
}

// CreateProfileParams configures a profile creation.
type CreateProfileParams struct {
	Name                 string  `json:"name"`
	BrandingLogoURL      *string `json:"branding_logo_url,omitempty"`
	BrandingDisplayName  *string `json:"branding_display_name,omitempty"`
	BrandingPrimaryColor *string `json:"branding_primary_color,omitempty"`
}

// UpdateProfileParams configures a profile update.
type UpdateProfileParams struct {
	Name                 *string `json:"name,omitempty"`
	BrandingLogoURL      *string `json:"branding_logo_url,omitempty"`
	BrandingDisplayName  *string `json:"branding_display_name,omitempty"`
	BrandingPrimaryColor *string `json:"branding_primary_color,omitempty"`
}

// PaginatedProfiles is a list response.
type PaginatedProfiles struct {
	Data []Profile
	Meta PageMeta
}

// ProfilesService handles profile operations.
type ProfilesService struct {
	client *Client
}

func (s *ProfilesService) List(ctx context.Context) (*PaginatedProfiles, error) {
	var env apiEnvelope[[]Profile]
	if err := s.client.do(ctx, http.MethodGet, "/v1/profiles", nil, nil, &env, nil); err != nil {
		return nil, err
	}
	return &PaginatedProfiles{Data: env.Data, Meta: pageMetaFromEnvelope(env)}, nil
}

func (s *ProfilesService) Create(ctx context.Context, params *CreateProfileParams) (*Profile, error) {
	body := map[string]any{}
	if params != nil {
		body["name"] = params.Name
		if params.BrandingLogoURL != nil {
			body["branding_logo_url"] = *params.BrandingLogoURL
		}
		if params.BrandingDisplayName != nil {
			body["branding_display_name"] = *params.BrandingDisplayName
		}
		if params.BrandingPrimaryColor != nil {
			body["branding_primary_color"] = *params.BrandingPrimaryColor
		}
	}
	var env apiEnvelope[Profile]
	if err := s.client.do(ctx, http.MethodPost, "/v1/profiles", nil, body, &env, nil); err != nil {
		return nil, err
	}
	return &env.Data, nil
}

func (s *ProfilesService) Get(ctx context.Context, profileID string) (*Profile, error) {
	var env apiEnvelope[Profile]
	if err := s.client.do(ctx, http.MethodGet, "/v1/profiles/"+profileID, nil, nil, &env, nil); err != nil {
		return nil, err
	}
	return &env.Data, nil
}

func (s *ProfilesService) Update(ctx context.Context, profileID string, params *UpdateProfileParams) (*Profile, error) {
	body := map[string]any{}
	if params != nil {
		if params.Name != nil {
			body["name"] = *params.Name
		}
		if params.BrandingLogoURL != nil {
			body["branding_logo_url"] = *params.BrandingLogoURL
		}
		if params.BrandingDisplayName != nil {
			body["branding_display_name"] = *params.BrandingDisplayName
		}
		if params.BrandingPrimaryColor != nil {
			body["branding_primary_color"] = *params.BrandingPrimaryColor
		}
	}
	var env apiEnvelope[Profile]
	if err := s.client.do(ctx, http.MethodPatch, "/v1/profiles/"+profileID, nil, body, &env, nil); err != nil {
		return nil, err
	}
	return &env.Data, nil
}

func (s *ProfilesService) Delete(ctx context.Context, profileID string) error {
	return s.client.do(ctx, http.MethodDelete, "/v1/profiles/"+profileID, nil, nil, nil, nil)
}
