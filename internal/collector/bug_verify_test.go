package collector

import (
	"distributed-tracing-system/internal/model"
	"distributed-tracing-system/internal/storage"
	"testing"
)

func TestBug05QueueFullIsReported(t *testing.T) {
	c := New(storage.NewMemory(), 0, 1)
	defer c.Close(t.Context())
	s := model.Span{TraceID: "0123456789abcdef0123456789abcdef", SpanID: "0123456789abcdef", ServiceName: "s", OperationName: "o", StartTime: 1}
	if err := c.Submit([]model.Span{s}); err != nil {
		t.Fatal(err)
	}
	if err := c.Submit([]model.Span{s}); err != ErrQueueFull {
		t.Fatalf("error=%v", err)
	}
}
