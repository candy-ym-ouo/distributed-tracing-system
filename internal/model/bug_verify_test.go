package model

import "testing"

func TestBug04CloneSpanDoesNotShareTags(t *testing.T) {
	a := Span{Tags: map[string]string{"k": "v"}}
	b := CloneSpan(a)
	b.Tags["k"] = "changed"
	if a.Tags["k"] != "v" {
		t.Fatal("tag mutation polluted source span")
	}
}
