package collector

import (
	"context"
	"sync"
	"testing"

	"distributed-tracing-system/internal/model"
	"distributed-tracing-system/internal/storage"
)

func TestCloseConcurrentSubmitDoesNotPanic(t *testing.T) {
	store := storage.NewMemory()
	c := New(store, 1, 8)
	span := model.Span{TraceID: "0123456789abcdef0123456789abcdef", SpanID: "0123456789abcdef", ServiceName: "svc", OperationName: "op", StartTime: 1}
	var wg sync.WaitGroup
	for i := 0; i < 16; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				_ = c.Submit([]model.Span{span})
			}
		}()
	}
	_ = c.Close(context.Background())
	wg.Wait()
}
