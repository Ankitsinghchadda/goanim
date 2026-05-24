package easing

import (
	"math"
	"testing"
)

func nearly(a, b float64) bool { return math.Abs(a-b) < 1e-6 }

func TestBoundaryConditions(t *testing.T) {
	cases := map[string]Func{
		"Linear":     Linear,
		"InQuad":     InQuad,
		"OutQuad":    OutQuad,
		"InOutQuad":  InOutQuad,
		"InCubic":    InCubic,
		"OutCubic":   OutCubic,
		"InOutCubic": InOutCubic,
		"InQuart":    InQuart,
		"OutQuart":   OutQuart,
		"InOutQuart": InOutQuart,
		"InExpo":     InExpo,
		"OutExpo":    OutExpo,
		"InOutExpo":  InOutExpo,
	}
	for name, f := range cases {
		if got := f(0); !nearly(got, 0) {
			t.Errorf("%s(0) = %v, want 0", name, got)
		}
		if got := f(1); !nearly(got, 1) {
			t.Errorf("%s(1) = %v, want 1", name, got)
		}
	}
}

func TestBackOvershoots(t *testing.T) {
	// OutBack reaches just over 1 mid-curve, then settles to 1 at t=1.
	if !nearly(OutBack(1), 1) {
		t.Errorf("OutBack(1) = %v, want 1", OutBack(1))
	}
	if v := OutBack(0.5); v < 0.7 {
		t.Errorf("OutBack(0.5) = %v, want > 0.7 (it should be on the way up)", v)
	}
}

func TestSpringDecays(t *testing.T) {
	f := Spring(180, 12, 1)
	if !nearly(f(0), 0) {
		t.Errorf("Spring at 0 should be 0, got %v", f(0))
	}
	// At t=1 the spring should have damped close to 1.
	if v := f(1); math.Abs(v-1) > 0.1 {
		t.Errorf("Spring at 1 should settle near 1, got %v", v)
	}
}

func TestClampedBeyondRange(t *testing.T) {
	if got := InCubic(2); !nearly(got, 1) {
		t.Errorf("InCubic(2) should clamp to 1, got %v", got)
	}
	if got := OutCubic(-1); !nearly(got, 0) {
		t.Errorf("OutCubic(-1) should clamp to 0, got %v", got)
	}
}
