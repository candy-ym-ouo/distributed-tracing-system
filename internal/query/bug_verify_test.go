package query

import (
	"context"
	"distributed-tracing-system/internal/storage"
	"errors"
	"testing"
)

func TestBug06MissingTraceReturnsNotFound(t *testing.T) {
	_, err := New(storage.NewMemory()).Trace(context.Background(), "0123456789abcdef0123456789abcdef")
	if !errors.Is(err, storage.ErrNotFound) {
		t.Fatalf("error=%v", err)
	}
}
