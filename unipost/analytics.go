package unipost

import (
	"context"
	"net/http"
	"strconv"
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

// AnalyticsPostsParams filters post-level analytics explorer rows.
type AnalyticsPostsParams struct {
	From      string
	To        string
	ProfileID string
	Platform  string
	Status    string
	AccountID string
	PostID    string
	Sort      string
	Limit     int
	Cursor    string
}

// AnalyticsPostRow is one post-level analytics explorer row.
type AnalyticsPostRow struct {
	PostID              string  `json:"post_id"`
	SocialPostResultID  string  `json:"social_post_result_id"`
	SocialAccountID     string  `json:"social_account_id"`
	ProfileID           string  `json:"profile_id"`
	Platform            string  `json:"platform"`
	ExternalID          string  `json:"external_id,omitempty"`
	ExternalUserID      string  `json:"external_user_id,omitempty"`
	ResultStatus        string  `json:"result_status"`
	PostStatus          string  `json:"post_status"`
	Caption             string  `json:"caption,omitempty"`
	URL                 string  `json:"url,omitempty"`
	CreatedAt           string  `json:"created_at"`
	PublishedAt         string  `json:"published_at,omitempty"`
	Impressions         int64   `json:"impressions"`
	Reach               int64   `json:"reach"`
	Likes               int64   `json:"likes"`
	Comments            int64   `json:"comments"`
	Shares              int64   `json:"shares"`
	Saves               int64   `json:"saves"`
	Clicks              int64   `json:"clicks"`
	VideoViews          int64   `json:"video_views"`
	EngagementRate      float64 `json:"engagement_rate"`
	PlatformSpecific    JSONMap `json:"platform_specific,omitempty"`
	FetchedAt           string  `json:"fetched_at,omitempty"`
	ConsecutiveFailures int32   `json:"consecutive_failures"`
	LastFailureReason   string  `json:"last_failure_reason,omitempty"`
}

// PaginatedAnalyticsPosts wraps analytics rows with pagination metadata.
type PaginatedAnalyticsPosts struct {
	Data       []AnalyticsPostRow
	Meta       PageMeta
	NextCursor string
}

// AnalyticsPlatformParams filters platform availability/detail endpoints.
type AnalyticsPlatformParams struct {
	From      string
	To        string
	ProfileID string
}

// AnalyticsPlatformAvailability describes analytics support and health for a platform.
type AnalyticsPlatformAvailability struct {
	Platform              string   `json:"platform"`
	SupportedMetrics      []string `json:"supported_metrics"`
	RefreshSupported      bool     `json:"refresh_supported"`
	AccountCount          int64    `json:"account_count"`
	ActiveAccountCount    int64    `json:"active_account_count"`
	NeedsReconnectCount   int64    `json:"needs_reconnect_count"`
	AnalyticsRowCount     int64    `json:"analytics_row_count"`
	LastSuccessfulFetchAt string   `json:"last_successful_fetch_at,omitempty"`
	LastFailureReason     string   `json:"last_failure_reason,omitempty"`
	Health                string   `json:"health"`
	Notes                 []string `json:"notes,omitempty"`
}

// AnalyticsPlatformSummary is the aggregate metrics block for a platform detail.
type AnalyticsPlatformSummary struct {
	Posts          int64   `json:"posts"`
	Accounts       int64   `json:"accounts"`
	Impressions    int64   `json:"impressions"`
	Reach          int64   `json:"reach"`
	Likes          int64   `json:"likes"`
	Comments       int64   `json:"comments"`
	Shares         int64   `json:"shares"`
	Saves          int64   `json:"saves"`
	Clicks         int64   `json:"clicks"`
	VideoViews     int64   `json:"video_views"`
	EngagementRate float64 `json:"engagement_rate"`
}

// AnalyticsPlatformTrendRow is one daily platform analytics bucket.
type AnalyticsPlatformTrendRow struct {
	Date        string `json:"date"`
	Posts       int64  `json:"posts"`
	Impressions int64  `json:"impressions"`
	Reach       int64  `json:"reach"`
	Likes       int64  `json:"likes"`
	Comments    int64  `json:"comments"`
	Shares      int64  `json:"shares"`
	Saves       int64  `json:"saves"`
	Clicks      int64  `json:"clicks"`
	VideoViews  int64  `json:"video_views"`
}

// AnalyticsAccountAvailability describes analytics health for one connected account.
type AnalyticsAccountAvailability struct {
	SocialAccountID       string `json:"social_account_id"`
	ProfileID             string `json:"profile_id"`
	AccountName           string `json:"account_name,omitempty"`
	ExternalUserID        string `json:"external_user_id,omitempty"`
	Status                string `json:"status"`
	PostCount             int64  `json:"post_count"`
	LastSuccessfulFetchAt string `json:"last_successful_fetch_at,omitempty"`
	LastFailureReason     string `json:"last_failure_reason,omitempty"`
}

// AnalyticsPlatformDetail is the platform detail explorer response.
type AnalyticsPlatformDetail struct {
	Platform     string                         `json:"platform"`
	Period       JSONMap                        `json:"period"`
	Availability AnalyticsPlatformAvailability  `json:"availability"`
	Summary      AnalyticsPlatformSummary       `json:"summary"`
	Trend        []AnalyticsPlatformTrendRow    `json:"trend"`
	Accounts     []AnalyticsAccountAvailability `json:"accounts"`
	TopPosts     []AnalyticsPostRow             `json:"top_posts"`
}

// AnalyticsRefreshParams marks matching analytics rows stale for worker refresh.
type AnalyticsRefreshParams struct {
	Platform  string `json:"platform,omitempty"`
	ProfileID string `json:"profile_id,omitempty"`
	AccountID string `json:"account_id,omitempty"`
	PostID    string `json:"post_id,omitempty"`
	From      string `json:"from,omitempty"`
	To        string `json:"to,omitempty"`
	Limit     int    `json:"limit,omitempty"`
}

// AnalyticsRefreshResponse is returned when refresh work is accepted.
type AnalyticsRefreshResponse struct {
	Status         string                 `json:"status"`
	MatchedCount   int64                  `json:"matched_count"`
	RequestedCount int64                  `json:"requested_count"`
	Limit          int                    `json:"limit"`
	ProcessedBy    string                 `json:"processed_by,omitempty"`
	Filters        AnalyticsRefreshParams `json:"filters"`
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

func analyticsPostsQuery(p *AnalyticsPostsParams) map[string]string {
	q := map[string]string{}
	if p == nil {
		return q
	}
	q["from"] = p.From
	q["to"] = p.To
	q["profile_id"] = p.ProfileID
	q["platform"] = p.Platform
	q["status"] = p.Status
	q["account_id"] = p.AccountID
	q["post_id"] = p.PostID
	q["sort"] = p.Sort
	q["cursor"] = p.Cursor
	if p.Limit > 0 {
		q["limit"] = strconv.Itoa(p.Limit)
	}
	return q
}

func analyticsPlatformQuery(p *AnalyticsPlatformParams) map[string]string {
	q := map[string]string{}
	if p == nil {
		return q
	}
	q["from"] = p.From
	q["to"] = p.To
	q["profile_id"] = p.ProfileID
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

func (s *AnalyticsService) Posts(ctx context.Context, params *AnalyticsPostsParams) (*PaginatedAnalyticsPosts, error) {
	var env apiEnvelope[[]AnalyticsPostRow]
	if err := s.client.do(ctx, http.MethodGet, "/v1/analytics/posts", analyticsPostsQuery(params), nil, &env, nil); err != nil {
		return nil, err
	}
	out := &PaginatedAnalyticsPosts{Data: env.Data, Meta: env.Meta}
	if env.Meta.NextCursor != "" {
		out.NextCursor = env.Meta.NextCursor
	} else {
		out.NextCursor = env.NextCursor
	}
	return out, nil
}

func (s *AnalyticsService) ExportPostsCSV(ctx context.Context, params *AnalyticsPostsParams) (string, error) {
	return s.client.doText(ctx, http.MethodGet, "/v1/analytics/posts/export", analyticsPostsQuery(params), nil)
}

func (s *AnalyticsService) Platforms(ctx context.Context, params *AnalyticsPlatformParams) ([]AnalyticsPlatformAvailability, error) {
	var env apiEnvelope[[]AnalyticsPlatformAvailability]
	if err := s.client.do(ctx, http.MethodGet, "/v1/analytics/platforms", analyticsPlatformQuery(params), nil, &env, nil); err != nil {
		return nil, err
	}
	if env.Data == nil {
		return []AnalyticsPlatformAvailability{}, nil
	}
	return env.Data, nil
}

func (s *AnalyticsService) Platform(ctx context.Context, platform string, params *AnalyticsPlatformParams) (*AnalyticsPlatformDetail, error) {
	var env apiEnvelope[AnalyticsPlatformDetail]
	if err := s.client.do(ctx, http.MethodGet, "/v1/analytics/platforms/"+platform, analyticsPlatformQuery(params), nil, &env, nil); err != nil {
		return nil, err
	}
	return &env.Data, nil
}

func (s *AnalyticsService) Refresh(ctx context.Context, params *AnalyticsRefreshParams) (*AnalyticsRefreshResponse, error) {
	var env apiEnvelope[AnalyticsRefreshResponse]
	if err := s.client.do(ctx, http.MethodPost, "/v1/analytics/refresh", nil, params, &env, nil); err != nil {
		return nil, err
	}
	return &env.Data, nil
}
