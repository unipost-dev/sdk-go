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
	Status         int                 // HTTP status code
	Code           string              // Resolved error code (prefers normalized_code)
	NormalizedCode string              // Lowercased canonical code from the API
	Message        string              // Human-readable message
	Errors         map[string][]string // 422 field-level errors, when applicable
	Platform       string              // Set on platform_error responses
	RetryAfter     int                 // Set on rate_limit responses, in seconds
	RequestID      string              // Server-assigned request id, for support tickets
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
	return &APIError{
		Status:         status,
		Code:           firstNonEmpty(env.Error.NormalizedCode, env.Error.Code),
		NormalizedCode: env.Error.NormalizedCode,
		Message:        msg,
		Errors:         env.Error.Errors,
		Platform:       env.Error.Platform,
		RetryAfter:     env.Error.RetryAfter,
		RequestID:      env.RequestID,
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
