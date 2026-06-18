package unipost

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"
)

// LogEntry represents a workspace developer log row.
type LogEntry struct {
	ID              int64   `json:"id"`
	WorkspaceID     string  `json:"workspace_id"`
	Timestamp       string  `json:"ts"`
	Level           string  `json:"level"`
	Status          string  `json:"status"`
	Category        string  `json:"category"`
	Action          string  `json:"action"`
	Source          string  `json:"source"`
	Message         string  `json:"message,omitempty"`
	RequestID       string  `json:"request_id,omitempty"`
	Platform        string  `json:"platform,omitempty"`
	ProfileID       string  `json:"profile_id,omitempty"`
	SocialAccountID string  `json:"social_account_id,omitempty"`
	PostID          string  `json:"post_id,omitempty"`
	ErrorCode       string  `json:"error_code,omitempty"`
	Metadata        JSONMap `json:"metadata,omitempty"`
	RequestPayload  JSONMap `json:"request_payload,omitempty"`
	ResponsePayload JSONMap `json:"response_payload,omitempty"`
}

// LogListParams are optional filters for listing logs.
type LogListParams struct {
	Category        string
	Action          string
	Source          string
	Level           string
	Status          string
	Platform        string
	ProfileID       string
	SocialAccountID string
	PostID          string
	RequestID       string
	ErrorCode       string
	Query           string
	From            string
	To              string
	Limit           int
	Cursor          string
}

// LogStreamParams are optional filters for streaming logs.
type LogStreamParams struct {
	Category        string
	Level           string
	Status          string
	Platform        string
	ProfileID       string
	SocialAccountID string
	PostID          string
	RequestID       string
	ErrorCode       string
	AfterID         int64
	LastEventID     string
}

// PaginatedLogs wraps a logs list response with cursor metadata.
type PaginatedLogs struct {
	Data       []LogEntry
	Meta       PageMeta
	NextCursor string
}

// LogsService handles developer-log operations.
type LogsService struct {
	client *Client
}

func logListQuery(p *LogListParams) map[string]string {
	q := map[string]string{}
	if p == nil {
		return q
	}
	q["category"] = p.Category
	q["action"] = p.Action
	q["source"] = p.Source
	q["level"] = p.Level
	q["status"] = p.Status
	q["platform"] = p.Platform
	q["profile_id"] = p.ProfileID
	q["social_account_id"] = p.SocialAccountID
	q["post_id"] = p.PostID
	q["request_id"] = p.RequestID
	q["error_code"] = p.ErrorCode
	q["q"] = p.Query
	q["from"] = p.From
	q["to"] = p.To
	q["cursor"] = p.Cursor
	if p.Limit > 0 {
		q["limit"] = strconv.Itoa(p.Limit)
	}
	return q
}

func logStreamQuery(p *LogStreamParams) map[string]string {
	q := map[string]string{}
	if p == nil {
		return q
	}
	q["category"] = p.Category
	q["level"] = p.Level
	q["status"] = p.Status
	q["platform"] = p.Platform
	q["profile_id"] = p.ProfileID
	q["social_account_id"] = p.SocialAccountID
	q["post_id"] = p.PostID
	q["request_id"] = p.RequestID
	q["error_code"] = p.ErrorCode
	if p.AfterID > 0 {
		q["after_id"] = strconv.FormatInt(p.AfterID, 10)
	}
	return q
}

func (s *LogsService) List(ctx context.Context, params *LogListParams) (*PaginatedLogs, error) {
	var env apiEnvelope[[]LogEntry]
	if err := s.client.do(ctx, http.MethodGet, "/v1/logs", logListQuery(params), nil, &env, nil); err != nil {
		return nil, err
	}
	out := &PaginatedLogs{Data: env.Data, Meta: env.Meta}
	if env.Meta.NextCursor != "" {
		out.NextCursor = env.Meta.NextCursor
	} else {
		out.NextCursor = env.NextCursor
	}
	return out, nil
}

func (s *LogsService) Get(ctx context.Context, logID int64) (*LogEntry, error) {
	var env apiEnvelope[LogEntry]
	if err := s.client.do(ctx, http.MethodGet, "/v1/logs/"+strconv.FormatInt(logID, 10), nil, nil, &env, nil); err != nil {
		return nil, err
	}
	return &env.Data, nil
}

func (s *LogsService) Stream(ctx context.Context, params *LogStreamParams) (*LogStream, error) {
	headers := map[string]string{}
	if params != nil && strings.TrimSpace(params.LastEventID) != "" {
		headers["Last-Event-ID"] = params.LastEventID
	}
	resp, err := s.client.doStream(ctx, "/v1/logs/stream", logStreamQuery(params), headers)
	if err != nil {
		return nil, err
	}
	return &LogStream{body: resp.Body, reader: bufio.NewReader(resp.Body)}, nil
}

// LogStream reads log.created events from a Server-Sent Events response.
type LogStream struct {
	body      io.ReadCloser
	reader    *bufio.Reader
	event     LogEntry
	eventName string
	id        string
	err       error
}

func (s *LogStream) Next() bool {
	if s == nil || s.err != nil {
		return false
	}

	var eventName string
	var eventID string
	var dataLines []string

	for {
		line, err := s.reader.ReadString('\n')
		if err != nil && len(line) == 0 {
			if err != io.EOF {
				s.err = err
			}
			return false
		}
		line = strings.TrimRight(line, "\r\n")
		if line == "" {
			if len(dataLines) == 0 {
				if err == io.EOF {
					return false
				}
				continue
			}
			if eventName != "" && eventName != "log.created" {
				dataLines = nil
				eventName = ""
				eventID = ""
				continue
			}
			var event LogEntry
			if decodeErr := json.Unmarshal([]byte(strings.Join(dataLines, "\n")), &event); decodeErr != nil {
				s.err = decodeErr
				return false
			}
			s.event = event
			s.eventName = eventName
			s.id = eventID
			return true
		}
		if strings.HasPrefix(line, ":") {
			if err == io.EOF {
				return false
			}
			continue
		}
		field, value, _ := strings.Cut(line, ":")
		value = strings.TrimPrefix(value, " ")
		switch field {
		case "event":
			eventName = value
		case "id":
			eventID = value
		case "data":
			dataLines = append(dataLines, value)
		}
		if err == io.EOF {
			return false
		}
	}
}

func (s *LogStream) Event() LogEntry {
	return s.event
}

func (s *LogStream) EventName() string {
	return s.eventName
}

func (s *LogStream) ID() string {
	return s.id
}

func (s *LogStream) Err() error {
	return s.err
}

func (s *LogStream) Close() error {
	if s == nil || s.body == nil {
		return nil
	}
	return s.body.Close()
}
