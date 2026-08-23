package tracing

import (
	"fmt"
	"net/http"

	"distributed-tracing-system/internal/model"
	"distributed-tracing-system/internal/propagator"
	"distributed-tracing-system/internal/tracer"
)

type Transport struct {
	Base       http.RoundTripper
	Tracer     *tracer.Tracer
	Propagator propagator.HTTP
}

func (t *Transport) RoundTrip(request *http.Request) (*http.Response, error) {
	base := t.Base
	if base == nil {
		base = http.DefaultTransport
	}
	ctx, span := t.Tracer.Start(request.Context(), request.Method+" "+request.URL.Host, tracer.Options{
		Kind: model.KindClient,
		Tags: map[string]string{
			"http.method":  request.Method,
			"http.url":     request.URL.String(),
			"peer.service": request.URL.Host,
		},
	})
	copy := request.Clone(ctx)
	copy.Header = request.Header.Clone()
	snapshot := span.Snapshot()
	t.Propagator.Inject(propagator.HeaderCarrier(copy.Header), propagator.Context{
		TraceID:      snapshot.TraceID,
		ParentSpanID: snapshot.SpanID,
		Sampled:      snapshot.Sampled,
	})
	response, err := base.RoundTrip(copy)
	if response != nil {
		span.SetTag("http.status_code", fmt.Sprint(response.StatusCode))
		if response.StatusCode >= 500 && err == nil {
			err = fmt.Errorf("http status %d", response.StatusCode)
		}
	}
	span.Finish(err)
	return response, err
}

func NewClient(t *tracer.Tracer) *http.Client {
	return &http.Client{Transport: &Transport{
		Tracer:     t,
		Propagator: propagator.HTTP{B3: true},
	}}
}
