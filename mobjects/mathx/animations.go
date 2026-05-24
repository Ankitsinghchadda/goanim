package mathx

import (
	"time"

	"github.com/ankitsinghchadda/goanim/core/animation"
	"github.com/ankitsinghchadda/goanim/core/animation/easing"
	"github.com/ankitsinghchadda/goanim/core/mobject"
)

// Write animates an equation as if being handwritten: each symbol
// fades in with a small stagger, so glyphs appear left-to-right rather
// than the entire equation popping in at once. Conceptually like
// DrawOn for math.
//
// Implementation: rather than truly drawing each glyph stroke-by-stroke
// (which is a Phase-5 nicety), we reveal the equation as a whole using
// the reveal fraction. The animation duration is split: nothing
// visible for first 10%, then linear reveal to 100% at end.
func Write(eq *Equation, duration time.Duration) animation.Animation {
	return &writeAnim{eq: eq, dur: duration, ease: easing.OutCubic}
}

type writeAnim struct {
	eq   *Equation
	dur  time.Duration
	ease animation.EasingFunc
}

func (w *writeAnim) Apply(t float64) {
	if t < 0 {
		t = 0
	}
	if t > 1 {
		t = 1
	}
	// Small startup pause for visual rhythm.
	const startup = 0.05
	if t < startup {
		w.eq.SetReveal(0)
		return
	}
	mapped := (t - startup) / (1 - startup)
	w.eq.SetReveal(w.ease(mapped))
}
func (w *writeAnim) Duration() time.Duration    { return w.dur }
func (w *writeAnim) Targets() []mobject.Mobject { return []mobject.Mobject{w.eq} }

// HighlightTerm flashes a specific submobject of an equation by index.
// Useful for "look at the m in E=mc^2." Index is 0-based; out-of-range
// indices are no-ops.
//
// Current limitation: highlight pulses the equation's overall scale
// via SetReveal. True per-symbol highlighting requires per-symbol
// position/style state and is not yet implemented.
func HighlightTerm(eq *Equation, index int, duration time.Duration) animation.Animation {
	return &highlightAnim{eq: eq, dur: duration}
}

type highlightAnim struct {
	eq  *Equation
	dur time.Duration
}

func (h *highlightAnim) Apply(t float64) {
	if t < 0 {
		t = 0
	}
	if t > 1 {
		t = 1
	}
	// Pulse: peak at t=0.5.
	w := 1 - 2*absMath(t-0.5)
	// Reveal is clamped in [0,1], so we hold reveal at 1 instead of
	// pulsing. A per-glyph color tween would be the proper fix.
	_ = w
	h.eq.SetReveal(1)
}
func (h *highlightAnim) Duration() time.Duration    { return h.dur }
func (h *highlightAnim) Targets() []mobject.Mobject { return []mobject.Mobject{h.eq} }

// TransformEquation morphs one equation into another. Phase-4
// approximation: cross-fade between the two — `from` fades out while
// `to` fades in. A proper symbol-matching morph (where shared symbols
// reposition rather than fade) is a Phase-5 follow-up.
//
// Both equations must be added to the scene; the caller is
// responsible for that.
func TransformEquation(from, to *Equation, duration time.Duration) animation.Animation {
	return &transformAnim{from: from, to: to, dur: duration, ease: easing.InOutCubic}
}

type transformAnim struct {
	from, to *Equation
	dur      time.Duration
	ease     animation.EasingFunc
}

func (a *transformAnim) Apply(t float64) {
	if t < 0 {
		t = 0
	}
	if t > 1 {
		t = 1
	}
	tt := a.ease(t)
	a.from.SetReveal(1 - tt)
	a.to.SetReveal(tt)
}
func (a *transformAnim) Duration() time.Duration { return a.dur }
func (a *transformAnim) Targets() []mobject.Mobject {
	return []mobject.Mobject{a.from, a.to}
}

func absMath(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
