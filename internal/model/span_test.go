package model

import (
	"errors"
	"testing"
	"time"
)

func TestSpanFinish(t *testing.T) {
	s := Span{TraceID: "0123456789abcdef0123456789abcdef", SpanID: "0123456789abcdef", StartTime: time.Now().UnixNano()}
	s.Finish(time.Unix(0, s.StartTime+2*time.Millisecond.Nanoseconds()), nil)
	if s.DurationUs != 2000 || s.Status != StatusOK {
		t.Fatalf("unexpected span: %+v", s)
	}
}
func TestSpanValidation(t *testing.T) {
	s := Span{TraceID: "bad", SpanID: "bad"}
	if errors.Is(s.Validate(), nil) {
		t.Fatal("expected invalid span")
	}
}
