package unipost

import (
	"encoding/json"
	"fmt"
)

// UniPostError is the base error type for all API errors.
type UniPostError struct {
	Status         int    // HTTP status code
	Code           string // Original error code from the API
	NormalizedCode string // Lowercased canonical code (e.g. "unauthorized")
	Message        string // Human-readable message
}

func (e *UniPostError) Error() string {
	if e == nil {
		return ""
	}
	code := e.NormalizedCode
	if code == "" {
		code = e.Code
	}
	if code == "" {
		return fmt.Sprintf("unipost: %s (status=%d)", e.Message, e.Status)
	}
	return fmt.Sprintf("unipost: %s (status=%d, code=%s)", e.Message, e.Status, code)
}

// AuthError represents a 401 authentication failure.
type AuthError struct{ UniPostError }

// NotFoundError represents a 404 not found.
type NotFoundError struct{ UniPostError }

// ValidationError represents a 422 validation failure.
type ValidationError struct {
	UniPostError
	Errors map[string][]string
}

// RateLimitError represents a 429 rate limit exceeded.
type RateLimitError struct {
	UniPostError
	RetryAfter int
}

// PlatformError represents a 502 platform-side error.
type PlatformError struct {
	UniPostError
	Platform string
}

// QuotaError represents a 403 quota exceeded.
type QuotaError struct{ UniPostError }

type errorEnvelope struct {
	Error struct {
		Code           string              `json:"code"`
		NormalizedCode string              `json:"normalized_code"`
		Message        string              `json:"message"`
		Errors         map[string][]string `json:"errors"`
		Platform       string              `json:"platform"`
		RetryAfter     int                 `json:"retry_after"`
	} `json:"error"`
	RequestID string `json:"request_id"`
}

func parseAPIError(status int, body []byte) error {
	var env errorEnvelope
	if len(body) > 0 {
		_ = json.Unmarshal(body, &env)
	}

	msg := env.Error.Message
	if msg == "" {
		msg = fmt.Sprintf("HTTP %d", status)
	}
	code := env.Error.Code
	if code == "" {
		code = env.Error.NormalizedCode
	}
	base := UniPostError{
		Status:         status,
		Code:           code,
		NormalizedCode: env.Error.NormalizedCode,
		Message:        msg,
	}

	switch status {
	case 401:
		return &AuthError{base}
	case 404:
		return &NotFoundError{base}
	case 422:
		return &ValidationError{UniPostError: base, Errors: env.Error.Errors}
	case 429:
		return &RateLimitError{UniPostError: base, RetryAfter: env.Error.RetryAfter}
	case 403:
		if env.Error.NormalizedCode == "quota_exceeded" || env.Error.Code == "quota_exceeded" {
			return &QuotaError{base}
		}
	case 502:
		if env.Error.Platform != "" {
			return &PlatformError{UniPostError: base, Platform: env.Error.Platform}
		}
	}
	return &base
}
