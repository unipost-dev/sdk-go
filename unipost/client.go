// Package unipost provides the official UniPost API client for Go.
package unipost

import (
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	defaultBaseURL = "https://api.unipost.dev"
	defaultTimeout = 30 * time.Second
	sdkVersion     = "0.2.5"
	userAgent      = "unipost-go/" + sdkVersion
)

// Option configures a Client.
type Option func(*Client)

// WithAPIKey sets the API key explicitly. By default the client reads
// UNIPOST_API_KEY from the environment.
func WithAPIKey(apiKey string) Option {
	return func(c *Client) { c.apiKey = apiKey }
}

// WithBaseURL overrides the API base URL (for self-hosted deployments).
func WithBaseURL(baseURL string) Option {
	return func(c *Client) { c.baseURL = strings.TrimRight(baseURL, "/") }
}

// WithHTTPClient sets a custom *http.Client (e.g. with custom transport).
func WithHTTPClient(client *http.Client) Option {
	return func(c *Client) { c.http = client }
}

// WithTimeout sets the HTTP request timeout. Ignored if WithHTTPClient is also set.
func WithTimeout(d time.Duration) Option {
	return func(c *Client) {
		if c.http == nil {
			c.http = &http.Client{Timeout: d}
		}
	}
}

// Client is the UniPost API client.
type Client struct {
	apiKey  string
	baseURL string
	http    *http.Client

	Workspace           *WorkspaceService
	Profiles            *ProfilesService
	Accounts            *AccountsService
	Platforms           *PlatformsService
	Plans               *PlansService
	PlatformCredentials *PlatformCredentialsService
	APIKeys             *APIKeysService
	Posts               *PostsService
	DeliveryJobs        *DeliveryJobsService
	Media               *MediaService
	Analytics           *AnalyticsService
	Connect             *ConnectService
	Users               *UsersService
	Webhooks            *WebhooksService
	OAuth               *OAuthService
	Usage               *UsageService
}

// NewClient creates a new UniPost API client.
//
// By default it reads the UNIPOST_API_KEY environment variable.
// Use WithAPIKey to override it.
//
//	client := unipost.NewClient()
//	client := unipost.NewClient(unipost.WithAPIKey("up_live_xxx"))
func NewClient(opts ...Option) *Client {
	c := &Client{
		apiKey:  os.Getenv("UNIPOST_API_KEY"),
		baseURL: defaultBaseURL,
		http:    &http.Client{Timeout: defaultTimeout},
	}
	for _, opt := range opts {
		opt(c)
	}

	c.Workspace = &WorkspaceService{client: c}
	c.Profiles = &ProfilesService{client: c}
	c.Accounts = &AccountsService{client: c}
	c.Platforms = &PlatformsService{client: c}
	c.Plans = &PlansService{client: c}
	c.PlatformCredentials = &PlatformCredentialsService{client: c}
	c.APIKeys = &APIKeysService{client: c}
	c.Posts = &PostsService{client: c}
	c.DeliveryJobs = &DeliveryJobsService{client: c}
	c.Media = &MediaService{client: c}
	c.Analytics = &AnalyticsService{client: c}
	c.Connect = &ConnectService{client: c}
	c.Users = &UsersService{client: c}
	c.Webhooks = &WebhooksService{client: c}
	c.OAuth = &OAuthService{client: c}
	c.Usage = &UsageService{client: c}
	return c
}
