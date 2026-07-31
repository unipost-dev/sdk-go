package unipost

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// XAccountCreditsReceipt describes the reservation and settlement for one X read.
type XAccountCreditsReceipt struct {
	OperationID       string `json:"operation_id"`
	Status            string `json:"status"`
	AccountingEnabled bool   `json:"accounting_enabled"`
	BillingMode       string `json:"billing_mode"`
	BypassReason      string `json:"bypass_reason,omitempty"`
	Operation         string `json:"operation"`
	Estimated         int    `json:"estimated"`
	Reserved          int    `json:"reserved"`
	Charged           int    `json:"charged"`
	Released          int    `json:"released"`
	CatalogVersion    string `json:"catalog_version"`
}

type XAccountPublicMetrics struct {
	Followers int `json:"followers"`
	Following int `json:"following"`
	Posts     int `json:"posts"`
	Listed    int `json:"listed"`
}

type XAccountProfile struct {
	AccountID         string                `json:"account_id"`
	Platform          string                `json:"platform"`
	ExternalAccountID string                `json:"external_account_id"`
	Username          string                `json:"username"`
	DisplayName       string                `json:"display_name"`
	Description       string                `json:"description"`
	ProfileImageURL   string                `json:"profile_image_url"`
	Location          *string               `json:"location,omitempty"`
	WebsiteURL        *string               `json:"website_url,omitempty"`
	AccountCreatedAt  string                `json:"account_created_at"`
	Verified          bool                  `json:"verified"`
	PublicMetrics     XAccountPublicMetrics `json:"public_metrics"`
	RetrievedAt       string                `json:"retrieved_at"`
}

type XAccountProfileMeta struct {
	Credits  XAccountCreditsReceipt `json:"credits"`
	Replayed *bool                  `json:"replayed,omitempty"`
}

type XAccountProfileResponse struct {
	Data      XAccountProfile     `json:"data"`
	Meta      XAccountProfileMeta `json:"meta"`
	RequestID string              `json:"request_id"`
}

type XAccountPostMedia struct {
	Type string `json:"type"`
}

type XAccountPostPublicMetrics struct {
	Likes       int `json:"likes"`
	Replies     int `json:"replies"`
	Reposts     int `json:"reposts"`
	Quotes      int `json:"quotes"`
	Impressions int `json:"impressions"`
}

type XAccountPostThread struct {
	ThreadID string `json:"thread_id"`
}

type XAccountPost struct {
	AccountID             string                    `json:"account_id"`
	ExternalPostID        string                    `json:"external_post_id"`
	Text                  string                    `json:"text"`
	CreatedAt             string                    `json:"created_at"`
	Language              *string                   `json:"language,omitempty"`
	ConversationID        string                    `json:"conversation_id"`
	ContentType           string                    `json:"content_type"`
	ReplyToExternalPostID *string                   `json:"reply_to_external_post_id,omitempty"`
	IsReply               bool                      `json:"is_reply"`
	IsSelfReply           bool                      `json:"is_self_reply"`
	IsRepost              bool                      `json:"is_repost"`
	IsQuote               bool                      `json:"is_quote"`
	Media                 []XAccountPostMedia       `json:"media,omitempty"`
	PublicMetrics         XAccountPostPublicMetrics `json:"public_metrics"`
	Thread                XAccountPostThread        `json:"thread"`
}

type XAccountPostsMeta struct {
	Limit           int                    `json:"limit"`
	ScannedCount    int                    `json:"scanned_count"`
	ReturnedCount   int                    `json:"returned_count"`
	HasMore         bool                   `json:"has_more"`
	NextCursor      string                 `json:"next_cursor,omitempty"`
	CursorExpiresAt string                 `json:"cursor_expires_at,omitempty"`
	Credits         XAccountCreditsReceipt `json:"credits"`
	Replayed        *bool                  `json:"replayed,omitempty"`
}

type XAccountPostsResponse struct {
	Data      []XAccountPost    `json:"data"`
	Meta      XAccountPostsMeta `json:"meta"`
	RequestID string            `json:"request_id"`
}

type XAccountProfileParams struct {
	ExternalUserID string
	IdempotencyKey string
}

type XAccountPostsParams struct {
	ExternalUserID         string
	IdempotencyKey         string
	Limit                  int
	Cursor                 string
	StartTime              string
	EndTime                string
	ExcludeReposts         bool
	ExcludeRepliesToOthers bool
}

// XAccountReadError adds X-read recovery metadata without changing APIError.
type XAccountReadError struct {
	APIError    *APIError
	Details     JSONMap
	IsRetriable *bool
	RetryAfter  int
}

func (e *XAccountReadError) Error() string {
	if e == nil || e.APIError == nil {
		return ""
	}
	return e.APIError.Error()
}

func (e *XAccountReadError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.APIError
}

// Profile reads the live profile for a connected X account.
func (s *AccountsService) Profile(
	ctx context.Context,
	accountID string,
	params *XAccountProfileParams,
) (*XAccountProfileResponse, error) {
	if err := validateXAccountProfileParams(accountID, params); err != nil {
		return nil, err
	}
	result, err := s.client.doResponseOnce(
		ctx,
		http.MethodGet,
		"/v1/accounts/"+url.PathEscape(accountID)+"/profile",
		map[string]string{"external_user_id": params.ExternalUserID},
		nil,
		map[string]string{"Idempotency-Key": params.IdempotencyKey},
	)
	if err != nil {
		return nil, err
	}
	if result.StatusCode < 200 || result.StatusCode >= 300 {
		return nil, newXAccountReadError(result)
	}
	var response XAccountProfileResponse
	if err := json.Unmarshal(result.Body, &response); err != nil {
		return nil, fmt.Errorf("unipost: decode X profile response: %w", err)
	}
	return &response, nil
}

// ListPosts reads one cursor-paginated authored-post page for a connected X account.
func (s *AccountsService) ListPosts(
	ctx context.Context,
	accountID string,
	params *XAccountPostsParams,
) (*XAccountPostsResponse, error) {
	if err := validateXAccountPostsParams(accountID, params); err != nil {
		return nil, err
	}
	query := map[string]string{
		"external_user_id": params.ExternalUserID,
		"limit":            strconv.Itoa(params.Limit),
		"cursor":           params.Cursor,
		"start_time":       params.StartTime,
		"end_time":         params.EndTime,
	}
	if params.ExcludeReposts {
		query["exclude_reposts"] = "true"
	}
	if params.ExcludeRepliesToOthers {
		query["exclude_replies_to_others"] = "true"
	}
	result, err := s.client.doResponseOnce(
		ctx,
		http.MethodGet,
		"/v1/accounts/"+url.PathEscape(accountID)+"/posts",
		query,
		nil,
		map[string]string{"Idempotency-Key": params.IdempotencyKey},
	)
	if err != nil {
		return nil, err
	}
	if result.StatusCode < 200 || result.StatusCode >= 300 {
		return nil, newXAccountReadError(result)
	}
	var response XAccountPostsResponse
	if err := json.Unmarshal(result.Body, &response); err != nil {
		return nil, fmt.Errorf("unipost: decode X posts response: %w", err)
	}
	return &response, nil
}

func validateXAccountProfileParams(accountID string, params *XAccountProfileParams) error {
	if strings.TrimSpace(accountID) == "" {
		return fmt.Errorf("unipost: accountID is required")
	}
	if params == nil {
		return fmt.Errorf("unipost: X account profile params are required")
	}
	if strings.TrimSpace(params.ExternalUserID) == "" {
		return fmt.Errorf("unipost: ExternalUserID is required")
	}
	if strings.TrimSpace(params.IdempotencyKey) == "" {
		return fmt.Errorf("unipost: IdempotencyKey is required")
	}
	return nil
}

func validateXAccountPostsParams(accountID string, params *XAccountPostsParams) error {
	if params == nil {
		return fmt.Errorf("unipost: X account posts params are required")
	}
	if err := validateXAccountProfileParams(accountID, &XAccountProfileParams{
		ExternalUserID: params.ExternalUserID,
		IdempotencyKey: params.IdempotencyKey,
	}); err != nil {
		return err
	}
	if params.Limit < 5 || params.Limit > 100 {
		return fmt.Errorf("unipost: Limit must be between 5 and 100")
	}
	return validateXReadTimeRange(params.StartTime, params.EndTime)
}

func validateXReadTimeRange(start, end string) error {
	var startTime time.Time
	var err error
	if start != "" {
		startTime, err = time.Parse(time.RFC3339, start)
		if err != nil {
			return fmt.Errorf("unipost: StartTime must be a valid RFC 3339 timestamp")
		}
	}
	var endTime time.Time
	if end != "" {
		endTime, err = time.Parse(time.RFC3339, end)
		if err != nil {
			return fmt.Errorf("unipost: EndTime must be a valid RFC 3339 timestamp")
		}
	}
	if !startTime.IsZero() && !endTime.IsZero() && !endTime.After(startTime) {
		return fmt.Errorf("unipost: EndTime must be after StartTime")
	}
	return nil
}

func newXAccountReadError(result *responseAwareHTTPResult) error {
	base, ok := parseAPIError(result.StatusCode, result.Body).(*APIError)
	if !ok {
		return parseAPIError(result.StatusCode, result.Body)
	}
	var envelope struct {
		Error struct {
			Details     JSONMap `json:"details"`
			IsRetriable *bool   `json:"is_retriable"`
		} `json:"error"`
	}
	_ = json.Unmarshal(result.Body, &envelope)
	return &XAccountReadError{
		APIError:    base,
		Details:     envelope.Error.Details,
		IsRetriable: envelope.Error.IsRetriable,
		RetryAfter:  retryAfterSeconds(result.Header),
	}
}

func retryAfterSeconds(header http.Header) int {
	raw := strings.TrimSpace(header.Get("Retry-After"))
	value, err := strconv.Atoi(raw)
	if err != nil || value < 0 {
		return 0
	}
	return value
}
