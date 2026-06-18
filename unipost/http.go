package unipost

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const maxRetries = 2

func (c *Client) do(ctx context.Context, method, path string, query map[string]string, body any, out any, headers map[string]string) error {
	if c.apiKey == "" {
		return fmt.Errorf("unipost: API key is required (set UNIPOST_API_KEY or use WithAPIKey)")
	}

	full, err := url.Parse(c.baseURL + path)
	if err != nil {
		return fmt.Errorf("unipost: invalid URL: %w", err)
	}
	if len(query) > 0 {
		values := full.Query()
		for k, v := range query {
			if strings.TrimSpace(v) != "" {
				values.Set(k, v)
			}
		}
		full.RawQuery = values.Encode()
	}

	var raw []byte
	if body != nil {
		raw, err = json.Marshal(body)
		if err != nil {
			return fmt.Errorf("unipost: marshal error: %w", err)
		}
	}

	var lastErr error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		var payload io.Reader
		if raw != nil {
			payload = bytes.NewReader(raw)
		}
		req, err := http.NewRequestWithContext(ctx, method, full.String(), payload)
		if err != nil {
			return fmt.Errorf("unipost: request error: %w", err)
		}
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
		req.Header.Set("User-Agent", userAgent)
		if raw != nil {
			req.Header.Set("Content-Type", "application/json")
		}
		for k, v := range headers {
			req.Header.Set(k, v)
		}

		resp, err := c.http.Do(req)
		if err != nil {
			return fmt.Errorf("unipost: HTTP error: %w", err)
		}

		respBody, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return fmt.Errorf("unipost: read error: %w", err)
		}

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			if out == nil || len(respBody) == 0 || resp.StatusCode == http.StatusNoContent {
				return nil
			}
			if err := json.Unmarshal(respBody, out); err != nil {
				return fmt.Errorf("unipost: decode error: %w", err)
			}
			return nil
		}

		if resp.StatusCode == 429 && attempt < maxRetries {
			retryAfter := 1
			if ra := resp.Header.Get("Retry-After"); ra != "" {
				if v, err := strconv.Atoi(ra); err == nil && v > 0 {
					retryAfter = v
				}
			}
			time.Sleep(time.Duration(retryAfter) * time.Second)
			lastErr = parseAPIError(resp.StatusCode, respBody)
			continue
		}

		return parseAPIError(resp.StatusCode, respBody)
	}

	if lastErr != nil {
		return lastErr
	}
	return fmt.Errorf("unipost: request failed after retries")
}

func (c *Client) doText(ctx context.Context, method, path string, query map[string]string, headers map[string]string) (string, error) {
	if c.apiKey == "" {
		return "", fmt.Errorf("unipost: API key is required (set UNIPOST_API_KEY or use WithAPIKey)")
	}

	full, err := url.Parse(c.baseURL + path)
	if err != nil {
		return "", fmt.Errorf("unipost: invalid URL: %w", err)
	}
	if len(query) > 0 {
		values := full.Query()
		for k, v := range query {
			if strings.TrimSpace(v) != "" {
				values.Set(k, v)
			}
		}
		full.RawQuery = values.Encode()
	}

	var lastErr error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		req, err := http.NewRequestWithContext(ctx, method, full.String(), nil)
		if err != nil {
			return "", fmt.Errorf("unipost: request error: %w", err)
		}
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
		req.Header.Set("User-Agent", userAgent)
		for k, v := range headers {
			req.Header.Set(k, v)
		}

		resp, err := c.http.Do(req)
		if err != nil {
			return "", fmt.Errorf("unipost: HTTP error: %w", err)
		}

		respBody, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return "", fmt.Errorf("unipost: read error: %w", err)
		}

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return string(respBody), nil
		}

		if resp.StatusCode == 429 && attempt < maxRetries {
			retryAfter := 1
			if ra := resp.Header.Get("Retry-After"); ra != "" {
				if v, err := strconv.Atoi(ra); err == nil && v > 0 {
					retryAfter = v
				}
			}
			time.Sleep(time.Duration(retryAfter) * time.Second)
			lastErr = parseAPIError(resp.StatusCode, respBody)
			continue
		}

		return "", parseAPIError(resp.StatusCode, respBody)
	}

	if lastErr != nil {
		return "", lastErr
	}
	return "", fmt.Errorf("unipost: request failed after retries")
}

func (c *Client) doStream(ctx context.Context, path string, query map[string]string, headers map[string]string) (*http.Response, error) {
	if c.apiKey == "" {
		return nil, fmt.Errorf("unipost: API key is required (set UNIPOST_API_KEY or use WithAPIKey)")
	}

	full, err := url.Parse(c.baseURL + path)
	if err != nil {
		return nil, fmt.Errorf("unipost: invalid URL: %w", err)
	}
	if len(query) > 0 {
		values := full.Query()
		for k, v := range query {
			if strings.TrimSpace(v) != "" {
				values.Set(k, v)
			}
		}
		full.RawQuery = values.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, full.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("unipost: request error: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "text/event-stream")
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("unipost: HTTP error: %w", err)
	}
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return resp, nil
	}

	body, readErr := io.ReadAll(resp.Body)
	resp.Body.Close()
	if readErr != nil {
		return nil, fmt.Errorf("unipost: read error: %w", readErr)
	}
	return nil, parseAPIError(resp.StatusCode, body)
}
