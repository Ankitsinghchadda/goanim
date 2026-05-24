package direction_test

import (
	"testing"
	"time"

	"github.com/ankitsinghchadda/goanim/core/animation"
	"github.com/ankitsinghchadda/goanim/core/direction"
	"github.com/ankitsinghchadda/goanim/core/mobject"
)

// TestPauseDuration — a Pause reports the requested duration and no
// targets.
func TestPauseDuration(t *testing.T) {
	p := direction.Pause(750 * time.Millisecond)
	if got := p.Duration(); got != 750*time.Millisecond {
		t.Errorf("Pause.Duration() = %v, want 750ms", got)
	}
	if got := p.Targets(); got != nil {
		t.Errorf("Pause.Targets() = %v, want nil", got)
	}
	// Apply is a no-op; the test passes if it doesn't panic for any t.
	for _, tt := range []float64{-1, 0, 0.25, 0.5, 0.75, 1, 2} {
		p.Apply(tt)
	}
}

// TestPauseLabel — the optional label is preserved via the Labeled
// interface, while unlabeled Pauses return empty.
func TestPauseLabel(t *testing.T) {
	bare := direction.Pause(100 * time.Millisecond)
	if l, ok := bare.(direction.Labeled); ok && l.Label() != "" {
		t.Errorf("bare Pause.Label() = %q, want \"\"", l.Label())
	}
	labeled := direction.Pause(100*time.Millisecond, "after-server-intro")
	l, ok := labeled.(direction.Labeled)
	if !ok {
		t.Fatal("labeled Pause does not satisfy direction.Labeled")
	}
	if l.Label() != "after-server-intro" {
		t.Errorf("labeled Pause.Label() = %q, want %q", l.Label(), "after-server-intro")
	}
}

// TestSequencePauseDuration — Sequence(FadeIn, Pause, FadeIn) reports
// the correct combined duration.
func TestSequencePauseDuration(t *testing.T) {
	m1, m2 := dummyMob(1), dummyMob(2)
	seq := animation.Sequence(
		animation.FadeIn(m1, 400*time.Millisecond),
		direction.Pause(1*time.Second),
		animation.FadeIn(m2, 600*time.Millisecond),
	)
	want := 400*time.Millisecond + 1*time.Second + 600*time.Millisecond
	if got := seq.Duration(); got != want {
		t.Errorf("Sequence(FadeIn, Pause, FadeIn).Duration() = %v, want %v", got, want)
	}
}

// TestParallelPauseDuration — Parallel takes the MAX of its children's
// durations, so a long Pause alongside a short animation extends the
// scene.
func TestParallelPauseDuration(t *testing.T) {
	m := dummyMob(1)
	par := animation.Parallel(
		animation.FadeIn(m, 200*time.Millisecond),
		direction.Pause(2*time.Second),
	)
	if got, want := par.Duration(), 2*time.Second; got != want {
		t.Errorf("Parallel(FadeIn, Pause).Duration() = %v, want %v", got, want)
	}
}

// TestStaggerPauseDuration — a Pause in a Stagger contributes to the
// total like any other animation.
func TestStaggerPauseDuration(t *testing.T) {
	m1, m2 := dummyMob(1), dummyMob(2)
	stag := animation.Stagger(100*time.Millisecond,
		animation.FadeIn(m1, 400*time.Millisecond),
		direction.Pause(800*time.Millisecond),
		animation.FadeIn(m2, 400*time.Millisecond),
	)
	// 3 children, inter-start delay 100ms → starts at 0ms, 100ms, 200ms.
	// Final child ends at 200 + 400 = 600ms. Middle child (Pause) ends at
	// 100 + 800 = 900ms. Max = 900ms.
	if got, want := stag.Duration(), 900*time.Millisecond; got != want {
		t.Errorf("Stagger duration = %v, want %v", got, want)
	}
}

// TestParallelOfSequenceWithPause — the composite worst-case the
// prompt cited: Parallel of a Sequence containing a Pause, alongside
// a Stagger. Reports the correct duration and runs Apply across the
// full range without panicking.
func TestParallelOfSequenceWithPause(t *testing.T) {
	m1, m2, m3 := dummyMob(1), dummyMob(2), dummyMob(3)
	// Sequence: 400 + 600 (Pause) + 400 = 1400ms
	seq := animation.Sequence(
		animation.FadeIn(m1, 400*time.Millisecond),
		direction.Pause(600*time.Millisecond),
		animation.FadeIn(m2, 400*time.Millisecond),
	)
	// Stagger of 3 200ms FadeIns at 150ms delay: 300 + 200 = 500ms total.
	stag := animation.Stagger(150*time.Millisecond,
		animation.FadeIn(m1, 200*time.Millisecond),
		animation.FadeIn(m2, 200*time.Millisecond),
		animation.FadeIn(m3, 200*time.Millisecond),
	)
	par := animation.Parallel(seq, stag)
	if got, want := par.Duration(), 1400*time.Millisecond; got != want {
		t.Errorf("Parallel(Sequence, Stagger).Duration() = %v, want %v", got, want)
	}
	// Apply across the full range. We're verifying composition doesn't
	// panic — the Pause inside the Sequence should be hit cleanly.
	for _, tt := range []float64{0, 0.1, 0.3, 0.5, 0.7, 0.9, 1.0} {
		par.Apply(tt)
	}
}

// dummyMob produces a minimal Mobject so we can pass it to FadeIn
// without pulling in real shapes.
func dummyMob(seed int64) mobject.Mobject {
	return mobject.NewGroup(seed)
}
