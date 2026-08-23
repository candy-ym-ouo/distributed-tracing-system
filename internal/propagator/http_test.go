package propagator

import (
	"net/http"
	"testing"
)

func TestHTTPRoundTrip(t *testing.T) {
	h := http.Header{}
	p := HTTP{B3: true}
	p.Inject(HeaderCarrier(h), Context{TraceID: "0123456789abcdef0123456789abcdef", ParentSpanID: "0123456789abcdef", Sampled: true})
	got, ok := p.Extract(HeaderCarrier(h))
	if !ok || got.TraceID == "" || !got.Sampled {
		t.Fatalf("got=%+v ok=%v", got, ok)
	}
}
