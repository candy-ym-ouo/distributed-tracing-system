package collector

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"sync"
	"time"

	"distributed-tracing-system/internal/model"
)

// Reporter is a client-side bounded buffer. Add never blocks callers; its
// background loop flushes on either batch size or interval.
type Reporter struct {
	endpoint string
	client   *http.Client
	queue    chan model.Span
	done     chan struct{}
	wait     sync.WaitGroup
	dropped  uint64
	mu       sync.Mutex
	closed   bool
}

func NewReporter(endpoint string, capacity int) *Reporter {
	r := &Reporter{
		endpoint: endpoint,
		client:   &http.Client{Timeout: 5 * time.Second},
		queue:    make(chan model.Span, capacity),
		done:     make(chan struct{}),
	}
	r.wait.Add(1)
	go r.run()
	return r
}

func (r *Reporter) Add(span model.Span) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.closed {
		return false
	}
	select {
	case r.queue <- span:
		return true
	default:
		r.dropped++
		return false
	}
}

func (r *Reporter) Dropped() uint64 {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.dropped
}

func (r *Reporter) run() {
	defer r.wait.Done()
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	batch := make([]model.Span, 0, 256)
	flush := func() {
		if len(batch) == 0 {
			return
		}
		_ = r.upload(batch)
		batch = batch[:0]
	}
	for {
		select {
		case span, ok := <-r.queue:
			if !ok {
				flush()
				return
			}
			batch = append(batch, span)
			if len(batch) >= cap(batch) {
				flush()
			}
		case <-ticker.C:
			flush()
		}
	}
}

func (r *Reporter) upload(spans []model.Span) error {
	payload, err := json.Marshal(struct {
		Spans []model.Span `json:"spans"`
	}{Spans: spans})
	if err != nil {
		return err
	}
	var last error
	for attempt := 0; attempt < 3; attempt++ {
		request, err := http.NewRequestWithContext(context.Background(), http.MethodPost, r.endpoint, bytes.NewReader(payload))
		if err != nil {
			return err
		}
		request.Header.Set("Content-Type", "application/json")
		response, err := r.client.Do(request)
		if err == nil {
			_ = response.Body.Close()
			if response.StatusCode < 500 {
				if response.StatusCode >= 200 && response.StatusCode < 300 {
					return nil
				}
				return errors.New("collector rejected batch")
			}
		}
		last = err
		time.Sleep(time.Duration(1<<attempt) * 100 * time.Millisecond)
	}
	return last
}

func (r *Reporter) Close() {
	r.mu.Lock()
	if r.closed {
		r.mu.Unlock()
		return
	}
	r.closed = true
	r.mu.Unlock()
	r.wait.Wait()
}
