package sampler

import "testing"

func TestRateBounds(t *testing.T) {
	zero, _ := NewRate(0)
	one, _ := NewRate(1)
	if zero.Sample("x") || !one.Sample("x") {
		t.Fatal("bounds failed")
	}
}
