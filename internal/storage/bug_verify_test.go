package storage

import (
	"context"
	"distributed-tracing-system/internal/model"
	"errors"
	"testing"
)

func TestBug08PutPreservesCanceledContext(t *testing.T) {
	f, err := OpenFile(t.TempDir() + "/spans.jsonl")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err = f.Put(ctx, []model.Span{})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("error=%v", err)
	}
}
