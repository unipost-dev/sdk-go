package unipost

import "time"

// ErrorSource identifies where a failure originated.
type ErrorSource string

const (
	ErrorSourceUnipost  ErrorSource = "unipost"
	ErrorSourcePlatform ErrorSource = "platform"
	ErrorSourceWorker   ErrorSource = "worker"
	ErrorSourceUnknown  ErrorSource = "unknown"
)

// ErrorTemporality identifies whether retrying may be useful.
type ErrorTemporality string

const (
	ErrorTemporalityTemporary ErrorTemporality = "temporary"
	ErrorTemporalityPermanent ErrorTemporality = "permanent"
	ErrorTemporalityUnknown   ErrorTemporality = "unknown"
)

// RetryState is a best-effort snapshot of UniPost's retry queue state.
type RetryState string

const (
	RetryStateNotRetriable RetryState = "not_retriable"
	RetryStateScheduled    RetryState = "scheduled"
	RetryStateRunning      RetryState = "running"
	RetryStateExhausted    RetryState = "exhausted"
	RetryStateManualOnly   RetryState = "manual_only"
	RetryStateUnknown      RetryState = "unknown"
)

// ProviderError contains sanitized, allowlisted platform error metadata.
type ProviderError struct {
	Provider      string `json:"provider,omitempty"`
	HTTPStatus    int    `json:"http_status,omitempty"`
	Code          string `json:"code,omitempty"`
	Subcode       string `json:"subcode,omitempty"`
	Type          string `json:"type,omitempty"`
	Reason        string `json:"reason,omitempty"`
	Domain        string `json:"domain,omitempty"`
	QuotaLimit    string `json:"quota_limit,omitempty"`
	QuotaLocation string `json:"quota_location,omitempty"`
	IsTransient   *bool  `json:"is_transient,omitempty"`
}

// RetryPolicy describes whether UniPost will retry automatically and whether
// a manual retry is currently allowed.
type RetryPolicy struct {
	IsRetriable        bool       `json:"is_retriable"`
	WillRetry          bool       `json:"will_retry"`
	RetryState         RetryState `json:"retry_state"`
	NextRunAt          *time.Time `json:"next_run_at,omitempty"`
	AttemptsMade       *int       `json:"attempts_made,omitempty"`
	MaxAttempts        *int       `json:"max_attempts,omitempty"`
	AttemptsRemaining  *int       `json:"attempts_remaining,omitempty"`
	ManualRetryAllowed bool       `json:"manual_retry_allowed"`
	Reason             string     `json:"reason,omitempty"`
}
