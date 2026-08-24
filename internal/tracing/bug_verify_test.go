package tracing

import (
	"distributed-tracing-system/internal/model"
	"distributed-tracing-system/internal/propagator"
	"distributed-tracing-system/internal/sampler"
	"distributed-tracing-system/internal/tracer"
	"io"
	"net/http"
	"net/url"
	"testing"
)

type captureRT struct{ headers http.Header }

func (c *captureRT) RoundTrip(r *http.Request) (*http.Response, error) {
	c.headers = r.Header
	return &http.Response{StatusCode: 200, Body: io.NopCloser(nil), Header: make(http.Header), Request: r}, nil
}

type captureSink struct{ spans []model.Span }

func (c *captureSink) Add(s model.Span) bool { c.spans = append(c.spans, s); return true }
func TestBug03ClientInjectsCurrentSpan(t *testing.T) {
	sink := &captureSink{}
	tr, _ := tracer.New("svc", sampler.Always{}, sink)
	rootCtx, _ := tr.Start(t.Context(), "root", tracer.Options{})
	cap := &captureRT{}
	u, _ := url.Parse("http://peer")
	client := &http.Client{Transport: &Transport{Base: cap, Tracer: tr, Propagator: propagator.HTTP{B3: true}}}
	_, _ = client.Do((&http.Request{Method: "GET", URL: u, Header: make(http.Header), Body: io.NopCloser(nil)}).WithContext(rootCtx))
	if len(sink.spans) == 0 || cap.headers.Get("X-Span-Id") != sink.spans[0].SpanID {
		t.Fatalf("span header=%q client=%+v", cap.headers.Get("X-Span-Id"), sink.spans)
	}
}
