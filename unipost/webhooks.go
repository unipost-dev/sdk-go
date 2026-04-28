package unipost

import (
	"context"
	"net/http"
	"time"
)

// WebhookSubscription represents a configured webhook endpoint.
type WebhookSubscription struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	URL           string    `json:"url"`
	Events        []string  `json:"events"`
	Active        bool      `json:"active"`
	Secret        string    `json:"secret,omitempty"`
	SecretPreview string    `json:"secret_preview"`
	CreatedAt     time.Time `json:"created_at"`
}

// CreateWebhookParams configures a webhook subscription creation.
type CreateWebhookParams struct {
	Name   string   `json:"name"`
	URL    string   `json:"url"`
	Events []string `json:"events"`
	Active *bool    `json:"active,omitempty"`
	Secret string   `json:"secret,omitempty"`
}

// UpdateWebhookParams configures a webhook subscription update.
type UpdateWebhookParams struct {
	Name   *string  `json:"name,omitempty"`
	URL    *string  `json:"url,omitempty"`
	Events []string `json:"events,omitempty"`
	Active *bool    `json:"active,omitempty"`
}

// PaginatedWebhookSubscriptions wraps a list response.
type PaginatedWebhookSubscriptions struct {
	Data []WebhookSubscription
	Meta PageMeta
}

// WebhooksService handles webhook subscription CRUD.
type WebhooksService struct {
	client *Client
}

func (s *WebhooksService) Create(ctx context.Context, params *CreateWebhookParams) (*WebhookSubscription, error) {
	var env apiEnvelope[WebhookSubscription]
	if err := s.client.do(ctx, http.MethodPost, "/v1/webhooks", nil, params, &env, nil); err != nil {
		return nil, err
	}
	return &env.Data, nil
}

func (s *WebhooksService) List(ctx context.Context) ([]WebhookSubscription, error) {
	page, err := s.ListPage(ctx)
	if err != nil {
		return nil, err
	}
	return page.Data, nil
}

func (s *WebhooksService) ListPage(ctx context.Context) (*PaginatedWebhookSubscriptions, error) {
	var env apiEnvelope[[]WebhookSubscription]
	if err := s.client.do(ctx, http.MethodGet, "/v1/webhooks", nil, nil, &env, nil); err != nil {
		return nil, err
	}
	return &PaginatedWebhookSubscriptions{Data: env.Data, Meta: pageMetaFromEnvelope(env)}, nil
}

func (s *WebhooksService) Get(ctx context.Context, webhookID string) (*WebhookSubscription, error) {
	var env apiEnvelope[WebhookSubscription]
	if err := s.client.do(ctx, http.MethodGet, "/v1/webhooks/"+webhookID, nil, nil, &env, nil); err != nil {
		return nil, err
	}
	return &env.Data, nil
}

func (s *WebhooksService) Update(ctx context.Context, webhookID string, params *UpdateWebhookParams) (*WebhookSubscription, error) {
	var env apiEnvelope[WebhookSubscription]
	if err := s.client.do(ctx, http.MethodPatch, "/v1/webhooks/"+webhookID, nil, params, &env, nil); err != nil {
		return nil, err
	}
	return &env.Data, nil
}

func (s *WebhooksService) Rotate(ctx context.Context, webhookID string) (*WebhookSubscription, error) {
	var env apiEnvelope[WebhookSubscription]
	if err := s.client.do(ctx, http.MethodPost, "/v1/webhooks/"+webhookID+"/rotate", nil, nil, &env, nil); err != nil {
		return nil, err
	}
	return &env.Data, nil
}

func (s *WebhooksService) Delete(ctx context.Context, webhookID string) error {
	return s.client.do(ctx, http.MethodDelete, "/v1/webhooks/"+webhookID, nil, nil, nil, nil)
}
