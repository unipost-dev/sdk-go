package unipost

import (
	"context"
	"net/http"
)

// OAuthConnectResponse is returned by OAuth.Connect.
type OAuthConnectResponse struct {
	AuthURL string `json:"auth_url"`
}

// OAuthService initiates dashboard-style OAuth flows for connecting accounts.
type OAuthService struct {
	client *Client
}

// Connect returns the auth URL to redirect a user to for the given platform.
func (s *OAuthService) Connect(ctx context.Context, platform, redirectURL string) (*OAuthConnectResponse, error) {
	q := map[string]string{}
	if redirectURL != "" {
		q["redirect_url"] = redirectURL
	}
	var env apiEnvelope[OAuthConnectResponse]
	if err := s.client.do(ctx, http.MethodGet, "/v1/oauth/connect/"+platform, q, nil, &env, nil); err != nil {
		return nil, err
	}
	return &env.Data, nil
}
