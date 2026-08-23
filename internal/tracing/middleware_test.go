package tracing

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"distributed-tracing-system/internal/model"
	"distributed-tracing-system/internal/propagator"
	"distributed-tracing-system/internal/sampler"
	"distributed-tracing-system/internal/tracer"
)

type spanSink struct{ spans []model.Span }

func (s *spanSink) Add(span model.Span) bool { s.spans = append(s.spans, span); return true }

func TestServerMiddlewareKeepsFirstStatus(t *testing.T) {
	sink := &spanSink{}
	tr, err := tracer.New("svc", sampler.Always{}, sink)
	if err != nil {
		t.Fatal(err)
	}
	handler := ServerMiddleware(tr, propagator.HTTP{})(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.WriteHeader(http.StatusOK)
	}))
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/x", nil))
	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d", recorder.Code)
	}
	if len(sink.spans) != 1 || sink.spans[0].Status != model.StatusError {
		t.Fatalf("spans = %+v", sink.spans)
	}
}
