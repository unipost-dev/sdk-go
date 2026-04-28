package unipost

import (
	"context"
	"net/http"
	"time"
)

// Workspace represents a UniPost workspace.
type Workspace struct {
	ID                     string    `json:"id"`
	Name                   string    `json:"name"`
	PerAccountMonthlyLimit *int32    `json:"per_account_monthly_limit,omitempty"`
	UsageModes             []string  `json:"usage_modes,omitempty"`
	CreatedAt              time.Time `json:"created_at"`
	UpdatedAt              time.Time `json:"updated_at"`
}

// UpdateWorkspaceParams configures a workspace update.
type UpdateWorkspaceParams struct {
	Name                   *string `json:"name,omitempty"`
	PerAccountMonthlyLimit *int32  `json:"per_account_monthly_limit,omitempty"`
}

// WorkspaceService handles workspace operations.
type WorkspaceService struct {
	client *Client
}

// Get returns the workspace bound to the authenticated caller.
func (s *WorkspaceService) Get(ctx context.Context) (*Workspace, error) {
	var env apiEnvelope[Workspace]
	if err := s.client.do(ctx, http.MethodGet, "/v1/workspace", nil, nil, &env, nil); err != nil {
		return nil, err
	}
	return &env.Data, nil
}

// Update modifies the authenticated workspace.
func (s *WorkspaceService) Update(ctx context.Context, params *UpdateWorkspaceParams) (*Workspace, error) {
	body := map[string]any{}
	if params != nil {
		if params.Name != nil {
			body["name"] = *params.Name
		}
		if params.PerAccountMonthlyLimit != nil {
			body["per_account_monthly_limit"] = *params.PerAccountMonthlyLimit
		}
	}
	var env apiEnvelope[Workspace]
	if err := s.client.do(ctx, http.MethodPatch, "/v1/workspace", nil, body, &env, nil); err != nil {
		return nil, err
	}
	return &env.Data, nil
}
