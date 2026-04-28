package unipost

// JSONMap is a free-form JSON object.
type JSONMap = map[string]any

// PageMeta carries pagination metadata returned by list endpoints.
type PageMeta struct {
	Total      *int   `json:"total,omitempty"`
	Limit      *int   `json:"limit,omitempty"`
	HasMore    *bool  `json:"has_more,omitempty"`
	NextCursor string `json:"next_cursor,omitempty"`
}

type apiEnvelope[T any] struct {
	Data       T       `json:"data"`
	Meta       PageMeta `json:"meta"`
	NextCursor string  `json:"next_cursor,omitempty"`
}

func pageMetaFromEnvelope[T any](env apiEnvelope[T]) PageMeta {
	return env.Meta
}
