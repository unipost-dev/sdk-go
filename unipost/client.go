package unipost

import (
	"fmt"
	"os"
	"time"
)

// ClientOption configures a Client.
type ClientOption func(*clientConfig)

type clientConfig struct {
	apiKey  string
	baseURL string
	timeout time.Duration
}

// WithAPIKey sets the API key explicitly.
func WithAPIKey(key string) ClientOption {
	return func(c *clientConfig) { c.apiKey = key }
}

// WithBaseURL overrides the API base URL (for private deployments).
func WithBaseURL(url string) ClientOption {
	return func(c *clientConfig) { c.baseURL = url }
}

// WithTimeout sets the HTTP request timeout.
func WithTimeout(d time.Duration) ClientOption {
	return func(c *clientConfig) { c.timeout = d }
}

// Client is the UniPost API client.
type Client struct {
	Accounts  *AccountsService
	Posts     *PostsService
	Media     *MediaService
	Analytics *AnalyticsService
	Connect   *ConnectService
	Users     *UsersService
	Profiles  *ProfilesService
}

// NewClient creates a new UniPost API client.
//
// By default it reads the UNIPOST_API_KEY environment variable.
// Use WithAPIKey to provide it explicitly.
//
//	client := unipost.NewClient()
//	client := unipost.NewClient(unipost.WithAPIKey("up_live_xxx"))
func NewClient(opts ...ClientOption) *Client {
	cfg := &clientConfig{
		baseURL: defaultBaseURL,
		timeout: defaultTimeout,
	}
	for _, opt := range opts {
		opt(cfg)
	}

	if cfg.apiKey == "" {
		cfg.apiKey = os.Getenv("UNIPOST_API_KEY")
	}
	if cfg.apiKey == "" {
		panic(fmt.Sprintf("unipost: API key is required. Set UNIPOST_API_KEY or use WithAPIKey()"))
	}

	h := newHTTPClient(cfg.apiKey, cfg.baseURL, cfg.timeout)

	return &Client{
		Accounts:  &AccountsService{http: h},
		Posts:     &PostsService{http: h},
		Media:     &MediaService{http: h},
		Analytics: &AnalyticsService{http: h},
		Connect:   &ConnectService{http: h},
		Users:     &UsersService{http: h},
		Profiles:  &ProfilesService{http: h},
	}
}
