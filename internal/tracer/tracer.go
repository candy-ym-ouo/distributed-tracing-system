package tracer

import (
	"context"
	"errors"
	"sync"
	"time"

	"distributed-tracing-system/internal/idgen"
	"distributed-tracing-system/internal/model"
	"distributed-tracing-system/internal/propagator"
	"distributed-tracing-system/internal/sampler"
)

type Reporter interface {
	Add(model.Span) bool
}

type Options struct {
	Kind   model.SpanKind
	Tags   map[string]string
	Remote *propagator.Context
}

type Tracer struct {
	service  string
	sampler  sampler.Sampler
	reporter Reporter
	now      func() time.Time
}

type ActiveSpan struct {
	mu       sync.Mutex
	span     model.Span
	tracer   *Tracer
	finished bool
}

type spanKey struct{}

func New(service string, sampling sampler.Sampler, reporter Reporter) (*Tracer, error) {
	if service == "" {
		return nil, errors.New("service name is required")
	}
	if sampling == nil {
		sampling = sampler.Always{}
	}
	return &Tracer{service: service, sampler: sampling, reporter: reporter, now: time.Now}, nil
}

func (t *Tracer) Start(ctx context.Context, operation string, options Options) (context.Context, *ActiveSpan) {
	traceID := ""
	parentID := ""
	sampled := false
	decisionKnown := false
	if options.Remote != nil {
		traceID = options.Remote.TraceID
		parentID = options.Remote.ParentSpanID
		sampled = options.Remote.Sampled
		decisionKnown = true
	} else if parent, ok := FromContext(ctx); ok {
		parentSnapshot := parent.Snapshot()
		traceID = parentSnapshot.TraceID
		parentID = parentSnapshot.SpanID
		sampled = parentSnapshot.Sampled
		decisionKnown = true
	}
	if traceID == "" {
		traceID = idgen.MustTraceID()
	}
	if !decisionKnown {
		sampled = t.sampler.Sample(traceID)
	}
	now := t.now()
	span := model.Span{
		TraceID:       traceID,
		SpanID:        idgen.MustSpanID(),
		ParentSpanID:  parentID,
		ServiceName:   t.service,
		OperationName: operation,
		Kind:          options.Kind,
		StartTime:     now.UnixNano(),
		Status:        model.StatusOK,
		Sampled:       sampled,
		Tags:          cloneTags(options.Tags),
	}
	active := &ActiveSpan{span: span, tracer: t}
	return ContextWithSpan(ctx, active), active
}

func (s *ActiveSpan) SetTag(key, value string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.finished {
		s.span.SetTag(key, value)
	}
}

func (s *ActiveSpan) Finish(err error) model.Span {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.finished {
		return model.CloneSpan(s.span)
	}
	s.finished = true
	s.span.Finish(s.tracer.now(), err)
	if (s.span.Sampled || s.span.Forced) && s.tracer.reporter != nil {
		s.tracer.reporter.Add(model.CloneSpan(s.span))
	}
	return model.CloneSpan(s.span)
}

func (s *ActiveSpan) Snapshot() model.Span {
	s.mu.Lock()
	defer s.mu.Unlock()
	return model.CloneSpan(s.span)
}

func ContextWithSpan(ctx context.Context, span *ActiveSpan) context.Context {
	return context.WithValue(ctx, spanKey{}, span)
}

func FromContext(ctx context.Context) (*ActiveSpan, bool) {
	span, ok := ctx.Value(spanKey{}).(*ActiveSpan)
	return span, ok
}

func cloneTags(tags map[string]string) map[string]string {
	if len(tags) == 0 {
		return nil
	}
	result := make(map[string]string, len(tags))
	for key, value := range tags {
		result[key] = value
	}
	return result
}
