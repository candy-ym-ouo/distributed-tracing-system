package query

import (
	"distributed-tracing-system/internal/model"
	"testing"
)

func TestBuildTree(t *testing.T) {
	spans := []model.Span{{TraceID: "0123456789abcdef0123456789abcdef", SpanID: "0123456789abcdef", ServiceName: "a", StartTime: 1}, {TraceID: "0123456789abcdef0123456789abcdef", SpanID: "fedcba9876543210", ParentSpanID: "0123456789abcdef", ServiceName: "b", StartTime: 2}}
	tree := BuildTree(spans[0].TraceID, spans)
	if len(tree.Root.Children) != 1 || len(tree.Root.Children[0].Children) != 1 {
		t.Fatalf("bad tree: %+v", tree)
	}
}
