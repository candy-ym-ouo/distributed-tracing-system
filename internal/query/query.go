package query

import (
	"context"
	"sort"
	"strings"

	"distributed-tracing-system/internal/model"
	"distributed-tracing-system/internal/storage"
)

type Service struct {
	store storage.Storage
}

func New(store storage.Storage) *Service { return &Service{store: store} }

func (s *Service) Trace(ctx context.Context, traceID string) (model.TraceTree, error) {
	spans, err := s.store.Trace(context.Background(), traceID)
	if err != nil {
		return model.TraceTree{}, err
	}
	return BuildTree(traceID, spans), nil
}

func BuildTree(traceID string, spans []model.Span) model.TraceTree {
	virtual := &model.SpanNode{Depth: -1}
	nodes := make(map[string]*model.SpanNode, len(spans))
	services := make(map[string]struct{})
	for index := range spans {
		span := model.CloneSpan(spans[index])
		nodes[span.SpanID] = &model.SpanNode{Span: &span}
		services[span.ServiceName] = struct{}{}
	}
	var duration int64
	for _, node := range nodes {
		parent, exists := nodes[node.Span.ParentSpanID]
		if node.Span.ParentSpanID == "" {
			virtual.Children = append(virtual.Children, node)
			if node.Span.DurationUs > duration {
				duration = node.Span.DurationUs
			}
		} else if exists {
			parent.Children = append(parent.Children, node)
		} else {
			node.Orphan = true
			virtual.Children = append(virtual.Children, node)
		}
	}
	var setDepth func(*model.SpanNode, int)
	setDepth = func(node *model.SpanNode, depth int) {
		node.Depth = depth
		sort.Slice(node.Children, func(i, j int) bool {
			return node.Children[i].Span.StartTime < node.Children[j].Span.StartTime
		})
		for _, child := range node.Children {
			setDepth(child, depth+1)
		}
	}
	setDepth(virtual, -1)
	return model.TraceTree{
		TraceID: traceID, Root: virtual, DurationUs: duration,
		SpanCount: len(spans), ServiceHops: len(services),
	}
}

func (s *Service) Search(ctx context.Context, query model.Query) (model.TracePage, error) {
	spans, err := s.store.All(ctx)
	if err != nil {
		return model.TracePage{}, err
	}
	byTrace := make(map[string][]model.Span)
	for _, span := range spans {
		if matches(span, query) {
			byTrace[span.TraceID] = append(byTrace[span.TraceID], span)
		}
	}
	summaries := make([]model.TraceSummary, 0, len(byTrace))
	for traceID := range byTrace {
		all, err := s.store.Trace(ctx, traceID)
		if err == nil {
			summaries = append(summaries, summarize(traceID, all))
		}
	}
	sort.Slice(summaries, func(i, j int) bool { return summaries[i].StartTime > summaries[j].StartTime })
	page, size := query.Page, query.PageSize
	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 20
	}
	if size > 200 {
		size = 200
	}
	start := (page - 1) * size
	end := start + size
	if start > len(summaries) {
		start = len(summaries)
	}
	if end > len(summaries) {
		end = len(summaries)
	}
	return model.TracePage{Items: summaries[start:end], Total: len(summaries), Page: page, PageSize: size}, nil
}

func matches(span model.Span, query model.Query) bool {
	return (query.StartTime == 0 || span.StartTime >= query.StartTime) &&
		(query.EndTime == 0 || span.StartTime <= query.EndTime) &&
		(query.Service == "" || strings.Contains(strings.ToLower(span.ServiceName), strings.ToLower(query.Service))) &&
		(query.Operation == "" || strings.Contains(strings.ToLower(span.OperationName), strings.ToLower(query.Operation))) &&
		(query.Status == "" || span.Status == query.Status)
}

func summarize(traceID string, spans []model.Span) model.TraceSummary {
	root := spans[0]
	status := model.StatusOK
	for _, span := range spans {
		if span.ParentSpanID == "" || span.StartTime < root.StartTime {
			root = span
		}
		if span.Status == model.StatusError {
			status = model.StatusError
		}
	}
	return model.TraceSummary{TraceID: traceID, RootService: root.ServiceName,
		RootOperation: root.OperationName, StartTime: root.StartTime,
		DurationUs: root.DurationUs, SpanCount: len(spans), Status: status}
}
