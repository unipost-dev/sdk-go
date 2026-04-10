package unipost

import (
	"context"
	"encoding/json"
	"net/url"
)

// AnalyticsService handles analytics operations.
type AnalyticsService struct {
	http *httpClient
}

// Rollup returns aggregated analytics data.
func (s *AnalyticsService) Rollup(ctx context.Context, params *AnalyticsRollupParams) (*AnalyticsRollup, error) {
	q := url.Values{}
	q.Set("from", params.From)
	q.Set("to", params.To)
	if params.Granularity != "" {
		q.Set("granularity", params.Granularity)
	}
	if params.GroupBy != "" {
		q.Set("group_by", params.GroupBy)
	}

	data, err := s.http.get(ctx, "/v1/analytics/rollup", q)
	if err != nil {
		return nil, err
	}

	var resp dataEnvelope[AnalyticsRollup]
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}
	return &resp.Data, nil
}
