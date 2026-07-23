package unipost

import (
	"encoding/json"
	"fmt"
	"strings"
)

// APIError is returned for any non-2xx response from the UniPost API.
//
// Inspect the Code (or NormalizedCode) field to branch on error kind:
//
//	if apiErr, ok := err.(*unipost.APIError); ok {
//	    switch apiErr.Code {
//	    case "unauthorized":
//	        // ...
//	    case "not_found":
//	        // ...
//	    }
//	}
type APIError struct {
	Status           int                 // HTTP status code
	Code             string              // Resolved code; Inbox reply errors preserve the raw server code
	NormalizedCode   string              // Lowercased canonical code from the API
	Message          string              // Human-readable message
	Errors           map[string][]string // 422 field-level errors, when applicable
	Platform         string              // Set on platform_error responses
	RetryAfter       int                 // Set on rate_limit responses, in seconds
	RequestID        string              // Server-assigned request id, for support tickets
	ErrorSource      ErrorSource         // Source of the failure, when classified by the API
	ErrorTemporality ErrorTemporality    // Whether retrying may be useful, when classified by the API
	ProviderError    *ProviderError      // Sanitized platform metadata, when available
	RetryPolicy      *RetryPolicy        // Best-effort retry queue snapshot, when available
}

func (e *APIError) Error() string {
	if e == nil {
		return ""
	}
	code := firstNonEmpty(e.NormalizedCode, e.Code)
	if code == "" {
		return fmt.Sprintf("unipost api error (%d): %s", e.Status, e.Message)
	}
	return fmt.Sprintf("unipost api error (%d %s): %s", e.Status, code, e.Message)
}

type errorEnvelope struct {
	Error struct {
		Code             string              `json:"code"`
		NormalizedCode   string              `json:"normalized_code"`
		Message          string              `json:"message"`
		Errors           map[string][]string `json:"errors"`
		Platform         string              `json:"platform"`
		RetryAfter       int                 `json:"retry_after"`
		ErrorSource      ErrorSource         `json:"error_source"`
		ErrorTemporality ErrorTemporality    `json:"error_temporality"`
		ProviderError    *ProviderError      `json:"provider_error"`
		RetryPolicy      *RetryPolicy        `json:"retry_policy"`
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
	return &APIError{
		Status:           status,
		Code:             firstNonEmpty(env.Error.NormalizedCode, env.Error.Code),
		NormalizedCode:   env.Error.NormalizedCode,
		Message:          msg,
		Errors:           env.Error.Errors,
		Platform:         env.Error.Platform,
		RetryAfter:       env.Error.RetryAfter,
		RequestID:        env.RequestID,
		ErrorSource:      env.Error.ErrorSource,
		ErrorTemporality: env.Error.ErrorTemporality,
		ProviderError:    env.Error.ProviderError,
		RetryPolicy:      env.Error.RetryPolicy,
	}
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}
