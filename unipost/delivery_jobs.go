package unipost

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// DeliveryJob represents a single platform-delivery job in the queue.
type DeliveryJob struct {
	ID                 string     `json:"id"`
	PostID             string     `json:"post_id"`
	SocialPostResultID string     `json:"social_post_result_id"`
	SocialAccountID    string     `json:"social_account_id"`
	Platform           string     `json:"platform"`
	Kind               string     `json:"kind"`
	State              string     `json:"state"`
	Attempts           int32      `json:"attempts"`
	MaxAttempts        int32      `json:"max_attempts"`
	FailureStage       *string    `json:"failure_stage,omitempty"`
	ErrorCode          *string    `json:"error_code,omitempty"`
	PlatformErrorCode  *string    `json:"platform_error_code,omitempty"`
	LastError          *string    `json:"last_error,omitempty"`
	NextRunAt          *time.Time `json:"next_run_at,omitempty"`
	LastAttemptAt      *time.Time `json:"last_attempt_at,omitempty"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
}

// ListDeliveryJobsParams are optional filters for listing delivery jobs.
type ListDeliveryJobsParams struct {
	Limit  int
	Offset int
	States []string // e.g. ["queued", "retrying"]
}

// DeliveryJobsService handles delivery-job operations.
type DeliveryJobsService struct {
	client *Client
}

func (s *DeliveryJobsService) List(ctx context.Context, params *ListDeliveryJobsParams) ([]DeliveryJob, error) {
	query := map[string]string{}
	if params != nil {
		if params.Limit > 0 {
			query["limit"] = strconv.Itoa(params.Limit)
		}
		if params.Offset > 0 {
			query["offset"] = strconv.Itoa(params.Offset)
		}
		if len(params.States) > 0 {
			query["states"] = strings.Join(params.States, ",")
		}
	}
	var env apiEnvelope[[]DeliveryJob]
	if err := s.client.do(ctx, http.MethodGet, "/v1/post-delivery-jobs", query, nil, &env, nil); err != nil {
		return nil, err
	}
	return env.Data, nil
}

func (s *DeliveryJobsService) Summary(ctx context.Context) (JSONMap, error) {
	var env apiEnvelope[JSONMap]
	if err := s.client.do(ctx, http.MethodGet, "/v1/post-delivery-jobs/summary", nil, nil, &env, nil); err != nil {
		return nil, err
	}
	return env.Data, nil
}

func (s *DeliveryJobsService) Retry(ctx context.Context, jobID string) (*DeliveryJob, error) {
	var env apiEnvelope[DeliveryJob]
	if err := s.client.do(ctx, http.MethodPost, "/v1/post-delivery-jobs/"+jobID+"/retry", nil, nil, &env, nil); err != nil {
		return nil, err
	}
	return &env.Data, nil
}

func (s *DeliveryJobsService) Cancel(ctx context.Context, jobID string) (*DeliveryJob, error) {
	var env apiEnvelope[DeliveryJob]
	if err := s.client.do(ctx, http.MethodPost, "/v1/post-delivery-jobs/"+jobID+"/cancel", nil, nil, &env, nil); err != nil {
		return nil, err
	}
	return &env.Data, nil
}
