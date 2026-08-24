package collector

import (
	"testing"
	"time"
)

func TestBug07ReporterCloseStopsWorker(t *testing.T) {
	r := NewReporter("http://127.0.0.1:1", 1)
	done := make(chan struct{})
	go func() { r.Close(); close(done) }()
	select {
	case <-done:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("reporter close did not stop worker")
	}
}
