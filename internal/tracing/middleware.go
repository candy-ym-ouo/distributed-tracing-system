package tracing

import (
	"fmt"
	"net/http"
	"strconv"

	"distributed-tracing-system/internal/model"
	"distributed-tracing-system/internal/propagator"
	"distributed-tracing-system/internal/tracer"
)

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(status int) {
	if w.status != 0 {
		return
	}
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func (w *statusWriter) Write(body []byte) (int, error) {
	if w.status == 0 {
		w.WriteHeader(http.StatusOK)
	}
	return w.ResponseWriter.Write(body)
}

func ServerMiddleware(t *tracer.Tracer, propagation propagator.HTTP) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			remote, ok := propagation.Extract(propagator.HeaderCarrier(r.Header))
			options := tracer.Options{
				Kind: model.KindServer,
				Tags: map[string]string{
					"http.method": r.Method,
					"http.path":   r.URL.Path,
				},
			}
			if ok {
				options.Remote = &remote
			}
			ctx, span := t.Start(r.Context(), r.Method+" "+r.URL.Path, options)
			writer := &statusWriter{ResponseWriter: w}
			next.ServeHTTP(writer, r.WithContext(ctx))
			if writer.status == 0 {
				writer.status = http.StatusOK
			}
			span.SetTag("http.status_code", strconv.Itoa(writer.status))
			if writer.status >= 500 {
				span.Finish(fmt.Errorf("http status %d", writer.status))
			} else {
				span.Finish(nil)
			}
		})
	}
}

func TraceContext(r *http.Request) (propagator.Context, bool) {
	span, ok := tracer.FromContext(r.Context())
	if !ok {
		return propagator.Context{}, false
	}
	snapshot := span.Snapshot()
	return propagator.Context{
		TraceID:      snapshot.TraceID,
		ParentSpanID: snapshot.SpanID,
		Sampled:      snapshot.Sampled,
	}, true
}
