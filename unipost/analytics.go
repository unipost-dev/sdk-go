package unipost

import (
	"context"
	"net/http"
)

// AnalyticsQueryParams are common filters for the summary/trend/by-platform endpoints.
type AnalyticsQueryParams struct {
	From      string
	To        string
	ProfileID string
	Platform  string
	Status    string
}

// AnalyticsRollupParams configures the rollup endpoint.
type AnalyticsRollupParams struct {
	From        string
	To          string
	Granularity string // "day" | "week" | "month"
	GroupBy     string // "platform" | "social_account_id" | "status" | "external_user_id"
}

// AnalyticsRollup is the rollup response shape.
type AnalyticsRollup struct {
	Granularity string    `json:"granularity"`
	GroupBy     []string  `json:"group_by"`
	Series      []JSONMap `json:"series"`
}

// AnalyticsService handles analytics operations.
type AnalyticsService struct {
	client *Client
}

func analyticsQuery(p *AnalyticsQueryParams) map[string]string {
	q := map[string]string{}
	if p == nil {
		return q
	}
	q["from"] = p.From
	q["to"] = p.To
	q["profile_id"] = p.ProfileID
	q["platform"] = p.Platform
	q["status"] = p.Status
	return q
}

func (s *AnalyticsService) Summary(ctx context.Context, params *AnalyticsQueryParams) (JSONMap, error) {
	var env apiEnvelope[JSONMap]
	if err := s.client.do(ctx, http.MethodGet, "/v1/analytics/summary", analyticsQuery(params), nil, &env, nil); err != nil {
		return nil, err
	}
	return env.Data, nil
}

func (s *AnalyticsService) Trend(ctx context.Context, params *AnalyticsQueryParams) (JSONMap, error) {
	var env apiEnvelope[JSONMap]
	if err := s.client.do(ctx, http.MethodGet, "/v1/analytics/trend", analyticsQuery(params), nil, &env, nil); err != nil {
		return nil, err
	}
	return env.Data, nil
}

func (s *AnalyticsService) ByPlatform(ctx context.Context, params *AnalyticsQueryParams) ([]JSONMap, error) {
	var env apiEnvelope[[]JSONMap]
	if err := s.client.do(ctx, http.MethodGet, "/v1/analytics/by-platform", analyticsQuery(params), nil, &env, nil); err != nil {
		return nil, err
	}
	return env.Data, nil
}

func (s *AnalyticsService) Rollup(ctx context.Context, params *AnalyticsRollupParams) (*AnalyticsRollup, error) {
	q := map[string]string{}
	if params != nil {
		q["from"] = params.From
		q["to"] = params.To
		q["granularity"] = params.Granularity
		q["group_by"] = params.GroupBy
	}
	var env apiEnvelope[AnalyticsRollup]
	if err := s.client.do(ctx, http.MethodGet, "/v1/analytics/rollup", q, nil, &env, nil); err != nil {
		return nil, err
	}
	return &env.Data, nil
}
