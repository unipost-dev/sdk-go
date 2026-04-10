package unipost

import "fmt"

// UniPostError is the base error type for all API errors.
type UniPostError struct {
	Message string
	Status  int
	Code    string
}

func (e *UniPostError) Error() string {
	return fmt.Sprintf("unipost: %s (status=%d, code=%s)", e.Message, e.Status, e.Code)
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

// PlatformError represents a platform-side error.
type PlatformError struct {
	UniPostError
	Platform string
}

// QuotaError represents a quota exceeded error.
type QuotaError struct{ UniPostError }

// parseAPIError converts an HTTP status and error body into a typed error.
func parseAPIError(status int, body map[string]interface{}) error {
	errObj, _ := body["error"].(map[string]interface{})
	msg, _ := errObj["message"].(string)
	code, _ := errObj["code"].(string)
	if msg == "" {
		msg = "Unknown API error"
	}
	if code == "" {
		code = "unknown"
	}

	base := UniPostError{Message: msg, Status: status, Code: code}

	switch status {
	case 401:
		return &AuthError{base}
	case 404:
		return &NotFoundError{base}
	case 422:
		ve := &ValidationError{UniPostError: base}
		if errs, ok := errObj["errors"].(map[string]interface{}); ok {
			ve.Errors = make(map[string][]string)
			for k, v := range errs {
				if arr, ok := v.([]interface{}); ok {
					for _, item := range arr {
						if s, ok := item.(string); ok {
							ve.Errors[k] = append(ve.Errors[k], s)
						}
					}
				}
			}
		}
		return ve
	case 429:
		return &RateLimitError{UniPostError: base}
	case 403:
		if code == "quota_exceeded" {
			return &QuotaError{base}
		}
	case 502:
		if platform, ok := errObj["platform"].(string); ok {
			return &PlatformError{UniPostError: base, Platform: platform}
		}
	}

	return &base
}
