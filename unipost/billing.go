package unipost

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

type XCreditsAllowance struct {
	Mode                    string  `json:"mode"`
	PlanID                  string  `json:"plan_id"`
	MonthlyAllowance        int     `json:"monthly_allowance"`
	MonthlyUsed             int     `json:"monthly_used"`
	MonthlyFinalized        int     `json:"monthly_finalized"`
	MonthlyPending          int     `json:"monthly_pending"`
	MonthlyEffective        int     `json:"monthly_effective"`
	MonthlyRemaining        int     `json:"monthly_remaining"`
	BillingPeriodStart      string  `json:"billing_period_start"`
	BillingPeriodEnd        string  `json:"billing_period_end"`
	CatalogVersion          string  `json:"catalog_version"`
	InboundDailyUsage       int     `json:"inbound_daily_usage"`
	InboundDailyLimit       int     `json:"inbound_daily_limit"`
	InboundEventsAccepted   int     `json:"inbound_events_accepted"`
	InboundEventsSuppressed int     `json:"inbound_events_suppressed"`
	InboundDailyResetAt     string  `json:"inbound_daily_reset_at"`
	InboundDailyPercent     float64 `json:"inbound_daily_percent"`
	PausePaidSources        bool    `json:"pause_paid_sources"`
	InboundPauseReason      string  `json:"inbound_pause_reason,omitempty"`
	ConnectionModeNote      string  `json:"connection_mode_note"`
}

type XCreditsAllowanceResponse struct {
	Data      XCreditsAllowance `json:"data"`
	RequestID string            `json:"request_id"`
}

type XCreditEvent struct {
	OperationID    string  `json:"operation_id"`
	AccountID      *string `json:"account_id,omitempty"`
	ExternalUserID *string `json:"external_user_id,omitempty"`
	Operation      string  `json:"operation"`
	CatalogVersion string  `json:"catalog_version"`
	Estimated      int     `json:"estimated"`
	Reserved       int     `json:"reserved"`
	Charged        int     `json:"charged"`
	Released       int     `json:"released"`
	Status         string  `json:"status"`
	CreatedAt      string  `json:"created_at"`
	UpdatedAt      string  `json:"updated_at"`
	FinalizedAt    *string `json:"finalized_at,omitempty"`
	ExpiresAt      *string `json:"expires_at,omitempty"`
}

type XCreditEventsMeta struct {
	NextCursor string `json:"next_cursor,omitempty"`
	HasMore    bool   `json:"has_more"`
	Limit      int    `json:"limit"`
}

type XCreditEventsResponse struct {
	Data      []XCreditEvent    `json:"data"`
	Meta      XCreditEventsMeta `json:"meta"`
	RequestID string            `json:"request_id"`
}

type ListXCreditEventsParams struct {
	AccountID      string
	ExternalUserID string
	Operation      string
	Status         string
	StartTime      string
	EndTime        string
	Cursor         string
	Limit          int
}

type BillingService struct {
	client *Client
}

func (s *BillingService) GetXCredits(ctx context.Context) (*XCreditsAllowanceResponse, error) {
	result, err := s.client.doResponseOnce(
		ctx,
		http.MethodGet,
		"/v1/billing/x-credits",
		nil,
		nil,
		nil,
	)
	if err != nil {
		return nil, err
	}
	if result.StatusCode < 200 || result.StatusCode >= 300 {
		return nil, apiErrorWithResponseHeaders(result)
	}
	var response XCreditsAllowanceResponse
	if err := json.Unmarshal(result.Body, &response); err != nil {
		return nil, fmt.Errorf("unipost: decode X Credits allowance response: %w", err)
	}
	return &response, nil
}

func (s *BillingService) ListXCreditEvents(
	ctx context.Context,
	params *ListXCreditEventsParams,
) (*XCreditEventsResponse, error) {
	if params == nil {
		params = &ListXCreditEventsParams{}
	}
	if params.Limit < 0 || params.Limit > 100 {
		return nil, fmt.Errorf("unipost: Limit must be between 1 and 100 when provided")
	}
	if err := validateXReadTimeRange(params.StartTime, params.EndTime); err != nil {
		return nil, err
	}
	query := map[string]string{
		"account_id":       params.AccountID,
		"external_user_id": params.ExternalUserID,
		"operation":        params.Operation,
		"status":           params.Status,
		"start_time":       params.StartTime,
		"end_time":         params.EndTime,
		"cursor":           params.Cursor,
	}
	if params.Limit > 0 {
		query["limit"] = strconv.Itoa(params.Limit)
	}
	result, err := s.client.doResponseOnce(
		ctx,
		http.MethodGet,
		"/v1/billing/x-credits/events",
		query,
		nil,
		nil,
	)
	if err != nil {
		return nil, err
	}
	if result.StatusCode < 200 || result.StatusCode >= 300 {
		return nil, apiErrorWithResponseHeaders(result)
	}
	var response XCreditEventsResponse
	if err := json.Unmarshal(result.Body, &response); err != nil {
		return nil, fmt.Errorf("unipost: decode X Credit events response: %w", err)
	}
	return &response, nil
}

func apiErrorWithResponseHeaders(result *responseAwareHTTPResult) error {
	err := parseAPIError(result.StatusCode, result.Body)
	if apiErr, ok := err.(*APIError); ok {
		if retryAfter := retryAfterSeconds(result.Header); retryAfter > 0 {
			apiErr.RetryAfter = retryAfter
		}
	}
	return err
}
