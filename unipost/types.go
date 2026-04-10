// Package unipost provides the official UniPost API client for Go.
package unipost

// Platform represents a supported social media platform.
type Platform string

const (
	PlatformTwitter   Platform = "twitter"
	PlatformLinkedIn  Platform = "linkedin"
	PlatformInstagram Platform = "instagram"
	PlatformThreads   Platform = "threads"
	PlatformTikTok    Platform = "tiktok"
	PlatformYouTube   Platform = "youtube"
	PlatformBluesky   Platform = "bluesky"
)

// SocialAccount represents a connected social media account.
type SocialAccount struct {
	ID                string `json:"id"`
	Platform          string `json:"platform"`
	AccountName       string `json:"account_name"`
	ExternalUserID    string `json:"external_user_id,omitempty"`
	ExternalUserEmail string `json:"external_user_email,omitempty"`
	ConnectedAt       string `json:"connected_at"`
	Status            string `json:"status"`
	ConnectionType    string `json:"connection_type"`
}

// AccountHealth represents the health status of a social account.
type AccountHealth struct {
	AccountID     string `json:"account_id"`
	Status        string `json:"status"` // "ok", "degraded", "disconnected"
	LastCheckedAt string `json:"last_checked_at"`
	Error         string `json:"error,omitempty"`
}

// Post represents a social media post.
type Post struct {
	ID          string           `json:"id"`
	Caption     string           `json:"caption"`
	Status      string           `json:"status"`
	ScheduledAt string           `json:"scheduled_at,omitempty"`
	CreatedAt   string           `json:"created_at"`
	PublishedAt string           `json:"published_at,omitempty"`
	Results     []PlatformResult `json:"results,omitempty"`
}

// PlatformResult represents the publish result for a specific platform.
type PlatformResult struct {
	SocialAccountID string `json:"social_account_id"`
	Platform        string `json:"platform,omitempty"`
	AccountName     string `json:"account_name,omitempty"`
	Status          string `json:"status"`
	ExternalID      string `json:"external_id,omitempty"`
	ErrorMessage    string `json:"error_message,omitempty"`
	PublishedAt     string `json:"published_at,omitempty"`
}

// CreatePostParams are the parameters for creating a new post.
type CreatePostParams struct {
	Caption        string                `json:"caption,omitempty"`
	AccountIDs     []string              `json:"account_ids,omitempty"`
	PlatformPosts  []CreatePlatformPost  `json:"platform_posts,omitempty"`
	MediaURLs      []string              `json:"media_urls,omitempty"`
	ScheduledAt    string                `json:"scheduled_at,omitempty"`
	Status         string                `json:"status,omitempty"`
	IdempotencyKey string                `json:"-"` // sent as header, not body
}

// CreatePlatformPost represents per-platform post data.
type CreatePlatformPost struct {
	AccountID      string `json:"account_id"`
	Caption        string `json:"caption,omitempty"`
	ThreadPosition int    `json:"thread_position,omitempty"`
	FirstComment   string `json:"first_comment,omitempty"`
	MediaIDs       []string `json:"media_ids,omitempty"`
}

// ListPostsParams are the parameters for listing posts.
type ListPostsParams struct {
	Status   string `json:"status,omitempty"`
	Platform string `json:"platform,omitempty"`
	From     string `json:"from,omitempty"`
	To       string `json:"to,omitempty"`
	Limit    int    `json:"limit,omitempty"`
	Cursor   string `json:"cursor,omitempty"`
}

// PostAnalytics represents analytics data for a post.
type PostAnalytics struct {
	PostID      string                       `json:"post_id"`
	Impressions int64                        `json:"impressions"`
	Engagements int64                        `json:"engagements"`
	Likes       int64                        `json:"likes"`
	Comments    int64                        `json:"comments"`
	Shares      int64                        `json:"shares"`
	Clicks      int64                        `json:"clicks"`
	Results     map[string]map[string]int64  `json:"results,omitempty"`
}

// ConnectSession represents an OAuth Connect session.
type ConnectSession struct {
	ID             string `json:"id"`
	URL            string `json:"url"`
	Status         string `json:"status"` // "pending", "completed", "expired"
	ExpiresAt      string `json:"expires_at"`
	Platform       string `json:"platform"`
	ExternalUserID string `json:"external_user_id"`
}

// CreateConnectSessionParams are the parameters for creating a Connect session.
type CreateConnectSessionParams struct {
	Platform          string `json:"platform"`
	ExternalUserID    string `json:"external_user_id"`
	ExternalUserEmail string `json:"external_user_email,omitempty"`
	ReturnURL         string `json:"return_url"`
}

// ManagedUser represents a user managed through Connect.
type ManagedUser struct {
	ExternalUserID    string          `json:"external_user_id"`
	ExternalUserEmail string          `json:"external_user_email,omitempty"`
	Accounts          []SocialAccount `json:"accounts"`
	CreatedAt         string          `json:"created_at"`
}

// MediaUploadRequest are the parameters for requesting a presigned upload URL.
type MediaUploadRequest struct {
	Filename    string `json:"filename"`
	ContentType string `json:"content_type"`
	SizeBytes   int64  `json:"size_bytes"`
}

// MediaUploadResponse contains the presigned URL and media ID.
type MediaUploadResponse struct {
	MediaID   string `json:"media_id"`
	UploadURL string `json:"upload_url"`
}

// AnalyticsRollupParams are the parameters for analytics rollup.
type AnalyticsRollupParams struct {
	From        string `json:"from"`
	To          string `json:"to"`
	Granularity string `json:"granularity,omitempty"` // "day", "week", "month"
	GroupBy     string `json:"group_by,omitempty"`    // "platform", "social_account_id", "status"
}

// AnalyticsRollup represents an analytics rollup response.
type AnalyticsRollup struct {
	From        string            `json:"from"`
	To          string            `json:"to"`
	Granularity string            `json:"granularity"`
	Buckets     []AnalyticsBucket `json:"buckets"`
}

// AnalyticsBucket represents a single analytics data point.
type AnalyticsBucket struct {
	Key         string `json:"key"`
	Impressions int64  `json:"impressions"`
	Engagements int64  `json:"engagements"`
	Likes       int64  `json:"likes"`
	Comments    int64  `json:"comments"`
	Shares      int64  `json:"shares"`
	Clicks      int64  `json:"clicks"`
}

// PaginatedResponse wraps paginated API responses.
type PaginatedResponse[T any] struct {
	Data       []T    `json:"data"`
	NextCursor string `json:"next_cursor,omitempty"`
}

// dataEnvelope wraps the { "data": T } API response pattern.
type dataEnvelope[T any] struct {
	Data T `json:"data"`
}
