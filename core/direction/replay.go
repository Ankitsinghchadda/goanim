package direction

import (
	"image/color"
	"time"

	"github.com/ankitsinghchadda/goanim/core/animation"
	"github.com/ankitsinghchadda/goanim/core/geometry"
	"github.com/ankitsinghchadda/goanim/core/mobject"
	"github.com/ankitsinghchadda/goanim/core/render"
	"github.com/ankitsinghchadda/goanim/core/style"
)

// Replay re-runs a captured animation at the given speed factor. A
// speed factor < 1 slows it down; > 1 speeds it up. During the replay
// a small "◀◀ replay" badge appears in the top-right corner of the
// frame so the viewer recognizes they're watching a re-run.
//
// IMPORTANT idempotency assumption: Replay requires the wrapped
// animation to be IDEMPOTENT — i.e., calling Apply(t) twice with the
// same t produces the same end state regardless of how it was
// reached. Goanim's built-in animations (FadeIn, MoveTo, MoveAlong,
// DrawOn) all satisfy this because they interpolate between
// captured start/end states. Composed animations (Sequence, Parallel,
// Stagger) inherit idempotency from their children.
//
// What does NOT work: replaying a sub-animation whose initial state
// was already set by a previous animation in the same scene. In that
// case the replay sees the post-animation state as the "start"
// state and the result will be a no-op. To replay a multi-step
// flow, wrap the whole flow in a Sequence and Replay that.
//
//	flow := animation.MoveAlong(packet, requestPath, 2*time.Second)
//	scene.Play(flow)
//	scene.Pause(1*time.Second)
//	scene.Play(direction.Replay(flow, 0.5))  // half-speed replay
func Replay(anim animation.Animation, speedFactor float64) animation.Animation {
	if speedFactor <= 0 {
		speedFactor = 1
	}
	originalDur := anim.Duration()
	newDur := time.Duration(float64(originalDur) / speedFactor)
	badge := newReplayBadge()
	return &replayAnim{
		inner: anim,
		dur:   newDur,
		badge: badge,
	}
}

type replayAnim struct {
	inner animation.Animation
	dur   time.Duration
	badge *replayBadge
}

func (r *replayAnim) Apply(t float64) {
	t = clamp01(t)
	r.inner.Apply(t)
	// Badge fades in at the start, holds, fades out at the end so it
	// doesn't pop sharply.
	switch {
	case t < 0.1:
		r.badge.opacity = t / 0.1
	case t > 0.9:
		r.badge.opacity = (1 - t) / 0.1
	default:
		r.badge.opacity = 1
	}
}
func (r *replayAnim) Duration() time.Duration { return r.dur }
func (r *replayAnim) Targets() []mobject.Mobject {
	// Inner targets + the badge so the scene auto-adds it.
	return append(r.inner.Targets(), r.badge)
}

// replayBadge is the "◀◀ replay" indicator that appears during a
// Replay. Frame-fixed (camera-independent) and styled to be subtle.
type replayBadge struct {
	*mobject.Group
	opacity float64
}

func newReplayBadge() *replayBadge {
	return &replayBadge{Group: mobject.NewGroup(0xDEADBEEF)}
}

func (b *replayBadge) Bounds() geometry.Rect {
	// Position in top-right of a 1920×1080 frame. The actual frame
	// dimensions aren't known at construction; Render reads ctx.BgColor
	// existence as a proxy and positions relative to a "default" frame.
	// For now we hardcode a reasonable spot.
	return geometry.RectFromCenter(geometry.Pt(840, 480), 220, 70)
}
func (b *replayBadge) Children() []mobject.Mobject { return nil }
func (b *replayBadge) SetReveal(t float64)         { b.opacity = clamp01(t) }
func (b *replayBadge) Render(rd render.Renderer, ctx style.Context) {
	if b.opacity <= 0 {
		return
	}
	// Top-right of frame: x = +780, y = +470 (assuming 1920×1080).
	const cx, cy = 780.0, 470.0
	const w, h = 240.0, 70.0
	bg := color.RGBA{0x1E, 0x1E, 0x1E, 0xE0} // semi-opaque dark
	rd.DrawPath(
		geometry.RectanglePath(cx-w/2, cy-h/2, w, h, 14),
		render.PathStyle{
			Fill:         style.ApplyOpacity(bg, b.opacity*0.9),
			IgnoreCamera: true,
		},
	)
	face := ctx.FontFace(style.FontSans)
	if face == nil {
		face = ctx.FontFace(style.FontHandDrawn)
	}
	rd.DrawText("◀◀  replay", cx, cy, render.TextStyle{
		Face:         face,
		Size:         32,
		Color:        style.ApplyOpacity(color.RGBA{0xFF, 0xFF, 0xFF, 0xFF}, b.opacity),
		Align:        render.AlignCenter,
		Baseline:     render.BaselineMiddle,
		IgnoreCamera: true,
	})
}
