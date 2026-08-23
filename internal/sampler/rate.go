package sampler

import (
	"encoding/hex"
	"errors"
	"hash/fnv"
	"sync"
	"time"
)

type Rate struct {
	threshold uint64
}

func NewRate(probability float64) (*Rate, error) {
	if probability < 0 || probability > 1 {
		return nil, errors.New("sampling probability must be between 0 and 1")
	}
	return &Rate{threshold: uint64(probability * 1_000_000)}, nil
}

func (r *Rate) Sample(traceID string) bool {
	if r.threshold == 0 {
		return false
	}
	if r.threshold >= 1_000_000 {
		return true
	}
	h := fnv.New64a()
	decoded, err := hex.DecodeString(traceID)
	if err != nil {
		decoded = []byte(traceID)
	}
	_, _ = h.Write(decoded)
	return h.Sum64()%1_000_000 < r.threshold
}

type RateLimit struct {
	mu       sync.Mutex
	limit    int
	window   time.Time
	accepted int
	now      func() time.Time
}

func NewRateLimit(limit int) *RateLimit {
	if limit < 0 {
		limit = 0
	}
	return &RateLimit{limit: limit, now: time.Now}
}

func (r *RateLimit) Sample(string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	now := r.now()
	if r.window.IsZero() || now.Sub(r.window) >= time.Second {
		r.window = now
		r.accepted = 0
	}
	if r.accepted >= r.limit {
		return false
	}
	r.accepted++
	return true
}
