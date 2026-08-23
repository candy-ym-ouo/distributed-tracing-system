package aggregation

import (
	"context"
	"sort"

	"distributed-tracing-system/internal/model"
	"distributed-tracing-system/internal/storage"
)

type Aggregator struct{ store storage.Storage }

func New(store storage.Storage) *Aggregator { return &Aggregator{store: store} }

func (a *Aggregator) spans(ctx context.Context, start, end int64) ([]model.Span, error) {
	all, err := a.store.All(ctx)
	if err != nil {
		return nil, err
	}
	result := all[:0]
	for _, span := range all {
		if (start == 0 || span.StartTime >= start) && (end == 0 || span.StartTime <= end) {
			result = append(result, span)
		}
	}
	return result, nil
}

func (a *Aggregator) Services(ctx context.Context, start, end int64) ([]model.ServiceStat, error) {
	spans, err := a.spans(ctx, start, end)
	if err != nil {
		return nil, err
	}
	type bucket struct {
		durations []int64
		errors    int64
		total     int64
	}
	buckets := make(map[string]*bucket)
	for _, span := range spans {
		b := buckets[span.ServiceName]
		if b == nil {
			b = &bucket{}
			buckets[span.ServiceName] = b
		}
		b.total++
		b.durations = append(b.durations, span.DurationUs)
		if span.Status == model.StatusError {
			b.errors++
		}
	}
	result := make([]model.ServiceStat, 0, len(buckets))
	for name, b := range buckets {
		sort.Slice(b.durations, func(i, j int) bool { return b.durations[i] < b.durations[j] })
		result = append(result, model.ServiceStat{ServiceName: name, SpanCount: b.total,
			ErrorCount: b.errors, ErrorRate: float64(b.errors) / float64(b.total),
			AvgDuration: average(b.durations), P50: percentile(b.durations, .50),
			P95: percentile(b.durations, .95), P99: percentile(b.durations, .99)})
	}
	sort.Slice(result, func(i, j int) bool { return result[i].SpanCount > result[j].SpanCount })
	return result, nil
}

func (a *Aggregator) Operations(ctx context.Context, start, end int64, limit int) ([]model.OperationStat, error) {
	spans, err := a.spans(ctx, start, end)
	if err != nil {
		return nil, err
	}
	type bucket struct {
		service, operation string
		durations          []int64
	}
	buckets := make(map[string]*bucket)
	for _, span := range spans {
		key := span.ServiceName + "\x00" + span.OperationName
		b := buckets[key]
		if b == nil {
			b = &bucket{service: span.ServiceName, operation: span.OperationName}
			buckets[key] = b
		}
		b.durations = append(b.durations, span.DurationUs)
	}
	result := make([]model.OperationStat, 0, len(buckets))
	for _, b := range buckets {
		sort.Slice(b.durations, func(i, j int) bool { return b.durations[i] < b.durations[j] })
		result = append(result, model.OperationStat{ServiceName: b.service, OperationName: b.operation,
			Count: int64(len(b.durations)), AvgDuration: average(b.durations), P99: percentile(b.durations, .99)})
	}
	sort.Slice(result, func(i, j int) bool { return result[i].P99 > result[j].P99 })
	if limit < 1 {
		limit = 10
	}
	if len(result) > limit {
		result = result[:limit]
	}
	return result, nil
}

func (a *Aggregator) Topology(ctx context.Context, start, end int64) ([]model.EdgeStat, error) {
	spans, err := a.spans(ctx, start, end)
	if err != nil {
		return nil, err
	}
	byTrace := make(map[string]map[string]model.Span)
	for _, span := range spans {
		if byTrace[span.TraceID] == nil {
			byTrace[span.TraceID] = make(map[string]model.Span)
		}
		byTrace[span.TraceID][span.SpanID] = span
	}
	type edge struct{ count, duration int64 }
	edges := make(map[string]*edge)
	for _, trace := range byTrace {
		for _, child := range trace {
			parent, ok := trace[child.ParentSpanID]
			if !ok || parent.ServiceName == child.ServiceName {
				continue
			}
			key := parent.ServiceName + "\x00" + child.ServiceName
			e := edges[key]
			if e == nil {
				e = &edge{}
				edges[key] = e
			}
			e.count++
			e.duration += child.DurationUs
		}
	}
	result := make([]model.EdgeStat, 0, len(edges))
	for key, e := range edges {
		for index := range key {
			if key[index] == 0 {
				result = append(result, model.EdgeStat{From: key[:index], To: key[index+1:],
					CallCount: e.count, AvgDuration: e.duration / e.count})
				break
			}
		}
	}
	sort.Slice(result, func(i, j int) bool { return result[i].CallCount > result[j].CallCount })
	return result, nil
}

func average(values []int64) int64 {
	if len(values) == 0 {
		return 0
	}
	var total int64
	for _, value := range values {
		total += value
	}
	return total / int64(len(values))
}

func percentile(values []int64, p float64) int64 {
	if len(values) == 0 {
		return 0
	}
	index := int(float64(len(values)-1)*p + .5)
	if index >= len(values) {
		index = len(values) - 1
	}
	return values[index]
}
