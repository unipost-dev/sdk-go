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
	"time"
)

const (
	defaultBaseURL = "https://api.unipost.dev"
	defaultTimeout = 30 * time.Second
	maxRetries     = 2
	userAgent      = "unipost-go/0.1.0"
)

type httpClient struct {
	apiKey  string
	baseURL string
	client  *http.Client
}

func newHTTPClient(apiKey, baseURL string, timeout time.Duration) *httpClient {
	return &httpClient{
		apiKey:  apiKey,
		baseURL: baseURL,
		client:  &http.Client{Timeout: timeout},
	}
}

func (h *httpClient) request(ctx context.Context, method, path string, body interface{}, query url.Values, headers map[string]string) ([]byte, error) {
	u, err := url.Parse(h.baseURL + path)
	if err != nil {
		return nil, fmt.Errorf("unipost: invalid URL: %w", err)
	}
	if query != nil {
		u.RawQuery = query.Encode()
	}

	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("unipost: marshal error: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	var lastErr error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		req, err := http.NewRequestWithContext(ctx, method, u.String(), bodyReader)
		if err != nil {
			return nil, fmt.Errorf("unipost: request error: %w", err)
		}

		req.Header.Set("Authorization", "Bearer "+h.apiKey)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("User-Agent", userAgent)
		for k, v := range headers {
			req.Header.Set(k, v)
		}

		resp, err := h.client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("unipost: HTTP error: %w", err)
		}

		respBody, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("unipost: read error: %w", err)
		}

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return respBody, nil
		}

		// Rate limit retry
		if resp.StatusCode == 429 && attempt < maxRetries {
			retryAfter := 1
			if ra := resp.Header.Get("Retry-After"); ra != "" {
				if v, err := strconv.Atoi(ra); err == nil {
					retryAfter = v
				}
			}
			time.Sleep(time.Duration(retryAfter) * time.Second)

			// Reset body reader for retry
			if body != nil {
				data, _ := json.Marshal(body)
				bodyReader = bytes.NewReader(data)
			}

			var errBody map[string]interface{}
			json.Unmarshal(respBody, &errBody)
			lastErr = parseAPIError(resp.StatusCode, errBody)
			continue
		}

		var errBody map[string]interface{}
		json.Unmarshal(respBody, &errBody)
		return nil, parseAPIError(resp.StatusCode, errBody)
	}

	if lastErr != nil {
		return nil, lastErr
	}
	return nil, fmt.Errorf("unipost: request failed after retries")
}

func (h *httpClient) get(ctx context.Context, path string, query url.Values) ([]byte, error) {
	return h.request(ctx, http.MethodGet, path, nil, query, nil)
}

func (h *httpClient) post(ctx context.Context, path string, body interface{}, headers map[string]string) ([]byte, error) {
	return h.request(ctx, http.MethodPost, path, body, nil, headers)
}

func (h *httpClient) delete(ctx context.Context, path string) ([]byte, error) {
	return h.request(ctx, http.MethodDelete, path, nil, nil, nil)
}
