package query

import (
	"context"
	"distributed-tracing-system/internal/model"
	"distributed-tracing-system/internal/storage"
	"errors"
	"testing"
)

type canceledStore struct{}

func (canceledStore) Put(context.Context, []model.Span) error { return nil }
func (canceledStore) Trace(ctx context.Context, _ string) ([]model.Span, error) {
	return nil, ctx.Err()
}
func (canceledStore) All(context.Context) ([]model.Span, error) { return nil, nil }
func (canceledStore) Stats() storage.Stats                      { return storage.Stats{} }
func (canceledStore) Close() error                              { return nil }
func TestBug01QueryPropagatesCanceledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := New(canceledStore{}).Trace(ctx, "0123456789abcdef0123456789abcdef")
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("error=%v", err)
	}
}
