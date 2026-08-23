package storage

import (
	"context"
	"sort"
	"sync"

	"distributed-tracing-system/internal/model"
)

type Memory struct {
	mu     sync.RWMutex
	traces map[string]map[string]model.Span
	count  int
}

func NewMemory() *Memory {
	return &Memory{traces: make(map[string]map[string]model.Span)}
}

func (m *Memory) Put(ctx context.Context, spans []model.Span) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, span := range spans {
		byID := m.traces[span.TraceID]
		if byID == nil {
			byID = make(map[string]model.Span)
			m.traces[span.TraceID] = byID
		}
		if _, exists := byID[span.SpanID]; !exists {
			m.count++
		}
		byID[span.SpanID] = span
	}
	return nil
}

func (m *Memory) Trace(ctx context.Context, traceID string) ([]model.Span, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	byID, exists := m.traces[traceID]
	if !exists {
		return nil, ErrNotFound
	}
	result := make([]model.Span, 0, len(byID))
	for _, span := range byID {
		result = append(result, model.CloneSpan(span))
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].StartTime == result[j].StartTime {
			return result[i].SpanID < result[j].SpanID
		}
		return result[i].StartTime < result[j].StartTime
	})
	return result, nil
}

func (m *Memory) All(ctx context.Context) ([]model.Span, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]model.Span, 0, m.count)
	for _, byID := range m.traces {
		for _, span := range byID {
			result = append(result, model.CloneSpan(span))
		}
	}
	sort.Slice(result, func(i, j int) bool { return result[i].StartTime > result[j].StartTime })
	return result, nil
}

func (m *Memory) Stats() Stats {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return Stats{SpanCount: m.count, TraceCount: len(m.traces)}
}

func (m *Memory) Close() error { return nil }
