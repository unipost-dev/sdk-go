package unipost

import (
	"context"
	"net/http"
	"strconv"
	"time"
)

// PlatformResult represents a per-platform delivery result for a post.
type PlatformResult struct {
	ID              string   `json:"id,omitempty"`
	SocialAccountID string   `json:"social_account_id"`
	Platform        string   `json:"platform,omitempty"`
	AccountName     string   `json:"account_name,omitempty"`
	Caption         string   `json:"caption,omitempty"`
	Status          string   `json:"status"`
	ExternalID      string   `json:"external_id,omitempty"`
	URL             string   `json:"url,omitempty"`
	ErrorMessage    string   `json:"error_message,omitempty"`
	PublishedAt     string   `json:"published_at,omitempty"`
	Warnings        []string `json:"warnings,omitempty"`
}

// Post represents a social media post.
type Post struct {
	ID                 string           `json:"id"`
	Caption            *string          `json:"caption"`
	MediaURLs          []string         `json:"media_urls,omitempty"`
	Status             string           `json:"status"`
	ExecutionMode      string           `json:"execution_mode,omitempty"`
	QueuedResultsCount int              `json:"queued_results_count,omitempty"`
	ActiveJobCount     int              `json:"active_job_count,omitempty"`
	RetryingCount      int              `json:"retrying_count,omitempty"`
	DeadCount          int              `json:"dead_count,omitempty"`
	CreatedAt          time.Time        `json:"created_at"`
	ScheduledAt        *time.Time       `json:"scheduled_at,omitempty"`
	PublishedAt        *time.Time       `json:"published_at,omitempty"`
	Results            []PlatformResult `json:"results,omitempty"`
}

// PaginatedPosts wraps a posts list with pagination metadata.
type PaginatedPosts struct {
	Data       []Post
	Meta       PageMeta
	NextCursor string
}

// CreatePostPlatform represents per-platform post data.
type CreatePostPlatform struct {
	AccountID       string         `json:"account_id"`
	Caption         string         `json:"caption,omitempty"`
	MediaURLs       []string       `json:"media_urls,omitempty"`
	MediaIDs        []string       `json:"media_ids,omitempty"`
	ThreadPosition  int            `json:"thread_position,omitempty"`
	FirstComment    string         `json:"first_comment,omitempty"`
	InReplyTo       string         `json:"in_reply_to,omitempty"`
	PlatformOptions map[string]any `json:"platform_options,omitempty"`
}

// CreatePostParams configures a post creation.
type CreatePostParams struct {
	Caption        string
	AccountIDs     []string
	MediaURLs      []string
	MediaIDs       []string
	ScheduledAt    string
	Status         string
	IdempotencyKey string
	PlatformPosts  []CreatePostPlatform
}

// UpdatePostParams configures a post update.
type UpdatePostParams struct {
	Caption       *string
	AccountIDs    []string
	MediaURLs     []string
	MediaIDs      []string
	ScheduledAt   *string
	Status        *string
	Archived      *bool
	PlatformPosts []CreatePostPlatform
}

// ValidationIssue represents a single validation error or warning.
type ValidationIssue struct {
	PlatformPostIndex int    `json:"platform_post_index"`
	AccountID         string `json:"account_id,omitempty"`
	Platform          string `json:"platform,omitempty"`
	Field             string `json:"field"`
	Code              string `json:"code"`
	Message           string `json:"message"`
	Severity          string `json:"severity"`
}

// ValidationResult is returned by Posts.Validate.
type ValidationResult struct {
	Valid    bool              `json:"valid"`
	Errors   []ValidationIssue `json:"errors"`
	Warnings []ValidationIssue `json:"warnings"`
}

// ListPostsParams are optional filters for listing posts.
type ListPostsParams struct {
	Status   string
	Platform string
	From     string
	To       string
	Limit    int
	Cursor   string
}

// PostQueueSnapshot is the response of Posts.GetQueue.
type PostQueueSnapshot struct {
	Post Post          `json:"post"`
	Jobs []DeliveryJob `json:"jobs"`
}

// PostAnalyticsItem is a per-platform analytics entry for a post.
type PostAnalyticsItem struct {
	PostID              string  `json:"post_id"`
	SocialAccountID     string  `json:"social_account_id"`
	Platform            string  `json:"platform"`
	ExternalID          string  `json:"external_id"`
	Impressions         int64   `json:"impressions"`
	Reach               int64   `json:"reach"`
	Likes               int64   `json:"likes"`
	Comments            int64   `json:"comments"`
	Shares              int64   `json:"shares"`
	Saves               int64   `json:"saves"`
	Clicks              int64   `json:"clicks"`
	VideoViews          int64   `json:"video_views"`
	Views               int64   `json:"views"`
	EngagementRate      float64 `json:"engagement_rate"`
	ConsecutiveFailures int32   `json:"consecutive_failures"`
	LastFailureReason   string  `json:"last_failure_reason,omitempty"`
}

// PostPreviewLink is the response of Posts.PreviewLink.
type PostPreviewLink struct {
	URL       string    `json:"url"`
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
}

// BulkError describes a single failed item in a bulk response.
type BulkError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// BulkPostResult is one entry in the bulk-create response array.
type BulkPostResult struct {
	Status int        `json:"status"`
	Data   *Post      `json:"data,omitempty"`
	Error  *BulkError `json:"error,omitempty"`
}

// PostsService handles post operations.
type PostsService struct {
	client *Client
}

func marshalCreatePostBody(p *CreatePostParams) map[string]any {
	body := map[string]any{}
	if p == nil {
		return body
	}
	if p.Caption != "" {
		body["caption"] = p.Caption
	}
	if len(p.AccountIDs) > 0 {
		body["account_ids"] = p.AccountIDs
	}
	if len(p.MediaURLs) > 0 {
		body["media_urls"] = p.MediaURLs
	}
	if len(p.MediaIDs) > 0 {
		body["media_ids"] = p.MediaIDs
	}
	if p.ScheduledAt != "" {
		body["scheduled_at"] = p.ScheduledAt
	}
	if p.Status != "" {
		body["status"] = p.Status
	}
	if len(p.PlatformPosts) > 0 {
		body["platform_posts"] = p.PlatformPosts
	}
	return body
}

func marshalUpdatePostBody(p *UpdatePostParams) map[string]any {
	body := map[string]any{}
	if p == nil {
		return body
	}
	if p.Caption != nil {
		body["caption"] = *p.Caption
	}
	if len(p.AccountIDs) > 0 {
		body["account_ids"] = p.AccountIDs
	}
	if len(p.MediaURLs) > 0 {
		body["media_urls"] = p.MediaURLs
	}
	if len(p.MediaIDs) > 0 {
		body["media_ids"] = p.MediaIDs
	}
	if p.ScheduledAt != nil {
		body["scheduled_at"] = *p.ScheduledAt
	}
	if p.Status != nil {
		body["status"] = *p.Status
	}
	if p.Archived != nil {
		body["archived"] = *p.Archived
	}
	if len(p.PlatformPosts) > 0 {
		body["platform_posts"] = p.PlatformPosts
	}
	return body
}

func (s *PostsService) List(ctx context.Context, params *ListPostsParams) (*PaginatedPosts, error) {
	query := map[string]string{}
	if params != nil {
		query["status"] = params.Status
		query["platform"] = params.Platform
		query["from"] = params.From
		query["to"] = params.To
		query["cursor"] = params.Cursor
		if params.Limit > 0 {
			query["limit"] = strconv.Itoa(params.Limit)
		}
	}
	var env apiEnvelope[[]Post]
	if err := s.client.do(ctx, http.MethodGet, "/v1/posts", query, nil, &env, nil); err != nil {
		return nil, err
	}
	out := &PaginatedPosts{Data: env.Data, Meta: env.Meta}
	if env.Meta.NextCursor != "" {
		out.NextCursor = env.Meta.NextCursor
	} else {
		out.NextCursor = env.NextCursor
	}
	return out, nil
}

func (s *PostsService) Get(ctx context.Context, postID string) (*Post, error) {
	var env apiEnvelope[Post]
	if err := s.client.do(ctx, http.MethodGet, "/v1/posts/"+postID, nil, nil, &env, nil); err != nil {
		return nil, err
	}
	return &env.Data, nil
}

func (s *PostsService) GetQueue(ctx context.Context, postID string) (*PostQueueSnapshot, error) {
	var env apiEnvelope[PostQueueSnapshot]
	if err := s.client.do(ctx, http.MethodGet, "/v1/posts/"+postID+"/queue", nil, nil, &env, nil); err != nil {
		return nil, err
	}
	return &env.Data, nil
}

func (s *PostsService) Analytics(ctx context.Context, postID string, refresh bool) ([]PostAnalyticsItem, error) {
	query := map[string]string{}
	if refresh {
		query["refresh"] = "true"
	}
	var env apiEnvelope[[]PostAnalyticsItem]
	if err := s.client.do(ctx, http.MethodGet, "/v1/posts/"+postID+"/analytics", query, nil, &env, nil); err != nil {
		return nil, err
	}
	if env.Data == nil {
		return []PostAnalyticsItem{}, nil
	}
	return env.Data, nil
}

func (s *PostsService) Create(ctx context.Context, params *CreatePostParams) (*Post, error) {
	headers := map[string]string{}
	if params != nil && params.IdempotencyKey != "" {
		headers["Idempotency-Key"] = params.IdempotencyKey
	}
	var env apiEnvelope[Post]
	if err := s.client.do(ctx, http.MethodPost, "/v1/posts", nil, marshalCreatePostBody(params), &env, headers); err != nil {
		return nil, err
	}
	return &env.Data, nil
}

func (s *PostsService) Validate(ctx context.Context, params *CreatePostParams) (*ValidationResult, error) {
	var env apiEnvelope[ValidationResult]
	if err := s.client.do(ctx, http.MethodPost, "/v1/posts/validate", nil, marshalCreatePostBody(params), &env, nil); err != nil {
		return nil, err
	}
	return &env.Data, nil
}

func (s *PostsService) Publish(ctx context.Context, postID string) (*Post, error) {
	var env apiEnvelope[Post]
	if err := s.client.do(ctx, http.MethodPost, "/v1/posts/"+postID+"/publish", nil, nil, &env, nil); err != nil {
		return nil, err
	}
	return &env.Data, nil
}

func (s *PostsService) Update(ctx context.Context, postID string, params *UpdatePostParams) (*Post, error) {
	var env apiEnvelope[Post]
	if err := s.client.do(ctx, http.MethodPatch, "/v1/posts/"+postID, nil, marshalUpdatePostBody(params), &env, nil); err != nil {
		return nil, err
	}
	return &env.Data, nil
}

func (s *PostsService) Archive(ctx context.Context, postID string) (*Post, error) {
	var env apiEnvelope[Post]
	if err := s.client.do(ctx, http.MethodPost, "/v1/posts/"+postID+"/archive", nil, nil, &env, nil); err != nil {
		return nil, err
	}
	return &env.Data, nil
}

func (s *PostsService) Restore(ctx context.Context, postID string) (*Post, error) {
	var env apiEnvelope[Post]
	if err := s.client.do(ctx, http.MethodPost, "/v1/posts/"+postID+"/restore", nil, nil, &env, nil); err != nil {
		return nil, err
	}
	return &env.Data, nil
}

func (s *PostsService) Cancel(ctx context.Context, postID string) (*Post, error) {
	var env apiEnvelope[Post]
	if err := s.client.do(ctx, http.MethodPost, "/v1/posts/"+postID+"/cancel", nil, nil, &env, nil); err != nil {
		return nil, err
	}
	return &env.Data, nil
}

func (s *PostsService) Delete(ctx context.Context, postID string) error {
	return s.client.do(ctx, http.MethodDelete, "/v1/posts/"+postID, nil, nil, nil, nil)
}

func (s *PostsService) PreviewLink(ctx context.Context, postID string) (*PostPreviewLink, error) {
	var env apiEnvelope[PostPreviewLink]
	if err := s.client.do(ctx, http.MethodPost, "/v1/posts/"+postID+"/preview-link", nil, nil, &env, nil); err != nil {
		return nil, err
	}
	return &env.Data, nil
}

func (s *PostsService) RetryResult(ctx context.Context, postID, resultID string) (*PlatformResult, error) {
	var env apiEnvelope[PlatformResult]
	if err := s.client.do(ctx, http.MethodPost, "/v1/posts/"+postID+"/results/"+resultID+"/retry", nil, nil, &env, nil); err != nil {
		return nil, err
	}
	return &env.Data, nil
}

func (s *PostsService) BulkCreate(ctx context.Context, posts []*CreatePostParams) ([]BulkPostResult, error) {
	body := make([]map[string]any, 0, len(posts))
	for _, p := range posts {
		body = append(body, marshalCreatePostBody(p))
	}
	payload := map[string]any{"posts": body}
	var env apiEnvelope[[]BulkPostResult]
	if err := s.client.do(ctx, http.MethodPost, "/v1/posts/bulk", nil, payload, &env, nil); err != nil {
		return nil, err
	}
	return env.Data, nil
}
