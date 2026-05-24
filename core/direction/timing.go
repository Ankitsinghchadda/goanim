package direction

import (
	"time"

	"github.com/ankitsinghchadda/goanim/core/animation"
	"github.com/ankitsinghchadda/goanim/core/mobject"
)

// Pause produces a zero-change animation that holds the scene for the
// given duration. Every frame during a Pause is identical to the last
// rendered frame.
//
// Pause is implemented as an Animation whose Apply is a no-op and
// whose Duration is the requested time, so it composes inside
// Sequence / Parallel / Stagger like any other animation:
//
//	scene.Play(sink, animation.Sequence(
//	    animation.FadeIn(server, 600*time.Millisecond),
//	    direction.Pause(2*time.Second),
//	    animation.FadeIn(database, 600*time.Millisecond),
//	))
//
// The optional variadic label is metadata for a future interactive
// presentation mode (advance-on-keypress at labeled pauses). In
// file-render mode it's ignored — but it survives the timeline so a
// later phase can pick it up. Only the first label is honored when
// multiple are passed (the variadic form is a convenience over a
// separate constructor).
//
// Implementation note: the scene's frame loop currently re-renders
// every Pause frame even though the output is identical. This is
// safe — rough geometry is cached, so re-rendering is cheap — but a
// frame-byte-cache optimization (hash the resolved mobject state and
// emit the previous frame's bytes when unchanged) is documented as
// a follow-up. The scene API doesn't change either way.
func Pause(duration time.Duration, label ...string) animation.Animation {
	p := &pauseAnim{dur: duration}
	if len(label) > 0 {
		p.label = label[0]
	}
	return p
}

// Labeled is implemented by Pause animations carrying a label. The
// scene player can use this to record pause-points in the timeline.
//
// Other Animation types may also implement Labeled in the future; it's
// kept separate from animation.Animation to avoid bloating that
// interface with a metadata concept.
type Labeled interface {
	Label() string
}

type pauseAnim struct {
	dur   time.Duration
	label string
}

func (p *pauseAnim) Apply(float64)              {} // intentional no-op
func (p *pauseAnim) Duration() time.Duration    { return p.dur }
func (p *pauseAnim) Targets() []mobject.Mobject { return nil }
func (p *pauseAnim) Label() string              { return p.label }

// IsConstantAnim marks Pause as a zero-change animation. The scene
// player checks for this and renders one frame, then writes it N
// times to the sink rather than re-rendering.
func (p *pauseAnim) IsConstantAnim() {}

// HoldFor is an alias for Pause that reads better at certain call
// sites — e.g. after building up a scene, "scene.Play(HoldFor(2s))"
// reads as "show this for 2 seconds." Functionally identical.
func HoldFor(duration time.Duration) animation.Animation {
	return Pause(duration)
}
