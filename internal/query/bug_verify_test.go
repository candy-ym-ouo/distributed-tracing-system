package query

import (
	"context"
	"distributed-tracing-system/internal/model"
	"distributed-tracing-system/internal/storage"
	"errors"
	"testing"
)

type missingStore struct{}

func (missingStore) Put(context.Context, []model.Span) error { return nil }
func (missingStore) Trace(context.Context, string) ([]model.Span, error) {
	return nil, storage.ErrNotFound
}
func (missingStore) All(context.Context) ([]model.Span, error) { return nil, nil }
func (missingStore) Stats() storage.Stats                      { return storage.Stats{} }
func (missingStore) Close() error                              { return nil }
func TestBug02NotFoundErrorIsPreserved(t *testing.T) {
	_, err := New(missingStore{}).Trace(context.Background(), "0123456789abcdef0123456789abcdef")
	if !errors.Is(err, storage.ErrNotFound) {
		t.Fatalf("error=%v", err)
	}
}
