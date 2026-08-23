package propagator

import (
	"encoding/hex"
	"net/http"
	"strings"
)

const (
	TraceHeader   = "X-Trace-Id"
	SpanHeader    = "X-Span-Id"
	SampledHeader = "X-Sampled"
)

type HeaderCarrier http.Header

func (h HeaderCarrier) Get(key string) string { return http.Header(h).Get(key) }
func (h HeaderCarrier) Set(key, value string) { http.Header(h).Set(key, value) }

type HTTP struct {
	B3 bool
}

func (p HTTP) Inject(carrier Carrier, value Context) {
	carrier.Set(TraceHeader, value.TraceID)
	carrier.Set(SpanHeader, value.ParentSpanID)
	carrier.Set(SampledHeader, boolString(value.Sampled))
	if p.B3 {
		carrier.Set("X-B3-TraceId", value.TraceID)
		carrier.Set("X-B3-SpanId", value.ParentSpanID)
		carrier.Set("X-B3-Sampled", boolString(value.Sampled))
	}
}

func (p HTTP) Extract(carrier Carrier) (Context, bool) {
	traceID := carrier.Get(TraceHeader)
	spanID := carrier.Get(SpanHeader)
	sampled := carrier.Get(SampledHeader)
	if traceID == "" && p.B3 {
		traceID = carrier.Get("X-B3-TraceId")
		spanID = carrier.Get("X-B3-SpanId")
		sampled = carrier.Get("X-B3-Sampled")
	}
	traceID = strings.ToLower(traceID)
	spanID = strings.ToLower(spanID)
	if !validHex(traceID, 32) || (spanID != "" && !validHex(spanID, 16)) {
		return Context{}, false
	}
	return Context{
		TraceID:      traceID,
		ParentSpanID: spanID,
		Sampled:      sampled == "1" || strings.EqualFold(sampled, "true"),
		Remote:       true,
	}, true
}

func validHex(value string, size int) bool {
	if len(value) != size {
		return false
	}
	_, err := hex.DecodeString(value)
	return err == nil
}

func boolString(value bool) string {
	if value {
		return "1"
	}
	return "0"
}
