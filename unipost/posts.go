package unipost

import (
	"context"
	"encoding/json"
	"net/url"
	"strconv"
)

// PostsService handles post operations.
type PostsService struct {
	http *httpClient
}

// Create creates a new post.
func (s *PostsService) Create(ctx context.Context, params *CreatePostParams) (*Post, error) {
	headers := map[string]string{}
	if params.IdempotencyKey != "" {
		headers["Idempotency-Key"] = params.IdempotencyKey
	}

	data, err := s.http.post(ctx, "/v1/social-posts", params, headers)
	if err != nil {
		return nil, err
	}

	var resp dataEnvelope[Post]
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}
	return &resp.Data, nil
}

// List returns posts matching the given filters.
func (s *PostsService) List(ctx context.Context, params *ListPostsParams) (*PaginatedResponse[Post], error) {
	q := url.Values{}
	if params != nil {
		if params.Status != "" {
			q.Set("status", params.Status)
		}
		if params.Platform != "" {
			q.Set("platform", params.Platform)
		}
		if params.From != "" {
			q.Set("from", params.From)
		}
		if params.To != "" {
			q.Set("to", params.To)
		}
		if params.Limit > 0 {
			q.Set("limit", strconv.Itoa(params.Limit))
		}
		if params.Cursor != "" {
			q.Set("cursor", params.Cursor)
		}
	}

	data, err := s.http.get(ctx, "/v1/social-posts", q)
	if err != nil {
		return nil, err
	}

	var resp PaginatedResponse[Post]
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// Get returns a single post by ID.
func (s *PostsService) Get(ctx context.Context, postID string) (*Post, error) {
	data, err := s.http.get(ctx, "/v1/social-posts/"+postID, nil)
	if err != nil {
		return nil, err
	}

	var resp dataEnvelope[Post]
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}
	return &resp.Data, nil
}

// Analytics returns analytics data for a post.
func (s *PostsService) Analytics(ctx context.Context, postID string) (*PostAnalytics, error) {
	data, err := s.http.get(ctx, "/v1/social-posts/"+postID+"/analytics", nil)
	if err != nil {
		return nil, err
	}

	var resp dataEnvelope[PostAnalytics]
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}
	return &resp.Data, nil
}

// Publish publishes a draft post.
func (s *PostsService) Publish(ctx context.Context, postID string) (*Post, error) {
	data, err := s.http.post(ctx, "/v1/social-posts/"+postID+"/publish", nil, nil)
	if err != nil {
		return nil, err
	}

	var resp dataEnvelope[Post]
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}
	return &resp.Data, nil
}

// Cancel cancels a scheduled post.
func (s *PostsService) Cancel(ctx context.Context, postID string) (*Post, error) {
	data, err := s.http.post(ctx, "/v1/social-posts/"+postID+"/cancel", nil, nil)
	if err != nil {
		return nil, err
	}

	var resp dataEnvelope[Post]
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}
	return &resp.Data, nil
}

// BulkCreate creates up to 50 posts at once.
func (s *PostsService) BulkCreate(ctx context.Context, posts []CreatePostParams) ([]Post, error) {
	data, err := s.http.post(ctx, "/v1/social-posts/bulk", posts, nil)
	if err != nil {
		return nil, err
	}

	var resp dataEnvelope[[]Post]
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}
	return resp.Data, nil
}
