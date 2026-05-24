// Package animation provides goanim's animation primitives: typed
// interpolators (FadeIn, MoveTo, DrawOn, ...) and composite
// combinators (Sequence, Parallel, Stagger). The Scene timeline in
// core/scene drives them.
//
// An Animation is invoked at normalized time t ∈ [0, 1] and mutates
// its target mobject(s) to the corresponding intermediate state. The
// concrete easing is selected by the animation, defaulting to one
// that feels good for the kind of motion (we generally avoid Linear
// and EaseInOut as defaults — they feel mechanical).
package animation

import (
	"time"

	"github.com/ankitsinghchadda/goanim/core/animation/easing"
	"github.com/ankitsinghchadda/goanim/core/mobject"
)

// Animation is the core interface. Implementations must be safe to
// call repeatedly with non-monotonic t (the timeline may seek); the
// expected behavior is "render the state at t" rather than "advance
// by dt."
type Animation interface {
	// Apply mutates the target(s) to the state at normalized time t.
	// Implementations must handle t outside [0, 1] gracefully (clamp).
	Apply(t float64)

	// Duration is how long the animation runs in scene-clock seconds.
	Duration() time.Duration

	// Targets returns the mobjects affected. Used by the scene to
	// auto-add them if missing.
	Targets() []mobject.Mobject
}

// EasingFunc is an alias kept in this package so callers don't have
// to import the sub-package for the common case.
type EasingFunc = easing.Func

// ConstantAnim is an optional capability marker for animations whose
// Apply(t) produces zero scene state change across t — i.e., Pause.
// The scene player checks for this at frame-loop construction time
// and, when present, renders one frame and writes the SAME pixel
// data N times to the sink instead of re-rendering identical frames.
//
// Implementations only need a marker method; the contract is "if you
// promise this, runFrames will skip per-tick Apply / Render". A
// non-Pause animation that wrongly claims this would freeze its
// targets during play, so don't.
type ConstantAnim interface {
	Animation
	IsConstantAnim()
}
