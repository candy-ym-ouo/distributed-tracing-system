package propagator

import (
	"net/http"
	"testing"
)

func TestBug09StandardHeadersWinOverB3(t *testing.T) {
	h := http.Header{"X-Trace-Id": []string{"0123456789abcdef0123456789abcdef"}, "X-Span-Id": []string{"0123456789abcdef"}, "X-Sampled": []string{"1"}, "X-B3-Traceid": []string{"fedcba9876543210fedcba9876543210"}, "X-B3-Spanid": []string{"fedcba9876543210"}, "X-B3-Sampled": []string{"0"}}
	got, ok := (HTTP{B3: true}).Extract(HeaderCarrier(h))
	if !ok || got.TraceID != "0123456789abcdef0123456789abcdef" || got.ParentSpanID != "0123456789abcdef" || !got.Sampled {
		t.Fatalf("got=%+v ok=%v", got, ok)
	}
}
