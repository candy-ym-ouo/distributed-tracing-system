package storage

import (
	"context"
	"errors"

	"distributed-tracing-system/internal/model"
)

var ErrNotFound = errors.New("trace not found")

type Stats struct {
	SpanCount  int   `json:"span_count"`
	TraceCount int   `json:"trace_count"`
	FileBytes  int64 `json:"file_bytes"`
}

type Storage interface {
	Put(context.Context, []model.Span) error
	Trace(context.Context, string) ([]model.Span, error)
	All(context.Context) ([]model.Span, error)
	Stats() Stats
	Close() error
}

type NopCloser struct {
	Storage
}

func (NopCloser) Close() error { return nil }
