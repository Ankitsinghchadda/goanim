package animation

import (
	"time"

	"github.com/ankitsinghchadda/goanim/core/animation/easing"
	"github.com/ankitsinghchadda/goanim/core/mobject"
	"github.com/ankitsinghchadda/goanim/core/style"
)

// Positioner is implemented by mobjects whose position can be moved
// by an animation. Systemdesign types (Client, Server, Database, the
// packet) all implement this.
type Positioner interface {
	mobject.Mobject
	Position() (float64, float64)
	SetPosition(x, y float64)
}

// FadeIn animates opacity 0 → 1 over duration. Default easing: OutCubic.
func FadeIn(target mobject.Mobject, duration time.Duration) Animation {
	return &opacityAnim{target: target, from: 0, to: 1, dur: duration, ease: easing.OutCubic}
}

// FadeOut animates opacity 1 → 0 over duration. Default easing: InCubic.
func FadeOut(target mobject.Mobject, duration time.Duration) Animation {
	return &opacityAnim{target: target, from: 1, to: 0, dur: duration, ease: easing.InCubic}
}

type opacityAnim struct {
	target   mobject.Mobject
	from, to float64
	dur      time.Duration
	ease     EasingFunc
}

func (a *opacityAnim) Apply(t float64) {
	t = clamp01(t)
	op := a.from + (a.to-a.from)*a.ease(t)
	st := a.target.Style()
	st.Opacity = &op
}
func (a *opacityAnim) Duration() time.Duration    { return a.dur }
func (a *opacityAnim) Targets() []mobject.Mobject { return []mobject.Mobject{a.target} }

// MoveTo translates the target to absolute (x, y).
//
// Crucial property for temporal stability: position is updated via
// SetPosition, NOT by regenerating the mobject. The rough geometry
// cached inside the mobject is reused as-is, just translated to the
// new center.
func MoveTo(target Positioner, x, y float64, duration time.Duration) Animation {
	fromX, fromY := target.Position()
	return &moveAnim{
		target: target, fromX: fromX, fromY: fromY,
		toX: x, toY: y, dur: duration, ease: easing.OutCubic,
	}
}

// Shift translates the target by (dx, dy).
func Shift(target Positioner, dx, dy float64, duration time.Duration) Animation {
	fromX, fromY := target.Position()
	return MoveTo(target, fromX+dx, fromY+dy, duration)
}

// MoveAlong moves the target along an arbitrary path. Useful for
// packet animations following arrows. The path is sampled by arc
// length so the motion has constant visual speed.
//
// Default easing: Linear (motion-along usually wants steady speed).
func MoveAlong(target Positioner, path *geometryPath, duration time.Duration) Animation {
	return &alongAnim{target: target, path: path, dur: duration, ease: easing.Linear}
}

// geometryPath is a tiny indirection so this file doesn't pull in
// geometry directly (avoid an import cycle if animation is ever
// imported by mobject).
type geometryPath = pathArg

// pathArg is a path container used by MoveAlong.
type pathArg struct {
	length float64
	sample func(arc float64) (x, y float64)
}

// NewPathFromSamples wraps an explicit arc-length sampler and total
// arc length in the form MoveAlong consumes.
func NewPathFromSamples(length float64, sample func(arc float64) (x, y float64)) *pathArg {
	return &pathArg{length: length, sample: sample}
}

type alongAnim struct {
	target Positioner
	path   *geometryPath
	dur    time.Duration
	ease   EasingFunc
}

func (a *alongAnim) Apply(t float64) {
	t = clamp01(t)
	arc := a.path.length * a.ease(t)
	x, y := a.path.sample(arc)
	a.target.SetPosition(x, y)
}
func (a *alongAnim) Duration() time.Duration    { return a.dur }
func (a *alongAnim) Targets() []mobject.Mobject { return []mobject.Mobject{a.target} }

type moveAnim struct {
	target       Positioner
	fromX, fromY float64
	toX, toY     float64
	dur          time.Duration
	ease         EasingFunc
}

func (a *moveAnim) Apply(t float64) {
	t = clamp01(t)
	p := a.ease(t)
	x := a.fromX + (a.toX-a.fromX)*p
	y := a.fromY + (a.toY-a.fromY)*p
	a.target.SetPosition(x, y)
}
func (a *moveAnim) Duration() time.Duration    { return a.dur }
func (a *moveAnim) Targets() []mobject.Mobject { return []mobject.Mobject{a.target} }

// Revealer is implemented by mobjects whose outline progress can be
// controlled (stroke-reveal). At Reveal(1.0) the mobject is fully
// drawn; at 0, completely hidden.
type Revealer interface {
	mobject.Mobject
	SetReveal(t float64)
}

// DrawOn reveals a mobject progressively along its outline. The
// underlying mechanism is path truncation: at progress p, the outline
// is rendered to its first p*length units of arc length.
//
// Default easing is OutCubic (the line slows down as it reaches the
// destination — feels natural for sketching).
func DrawOn(target Revealer, duration time.Duration) Animation {
	target.SetReveal(0)
	return &revealAnim{target: target, from: 0, to: 1, dur: duration, ease: easing.OutCubic}
}

// Erase reverses DrawOn — progressively hides the outline.
func Erase(target Revealer, duration time.Duration) Animation {
	target.SetReveal(1)
	return &revealAnim{target: target, from: 1, to: 0, dur: duration, ease: easing.InCubic}
}

type revealAnim struct {
	target   Revealer
	from, to float64
	dur      time.Duration
	ease     EasingFunc
}

func (a *revealAnim) Apply(t float64) {
	t = clamp01(t)
	a.target.SetReveal(a.from + (a.to-a.from)*a.ease(t))
}
func (a *revealAnim) Duration() time.Duration    { return a.dur }
func (a *revealAnim) Targets() []mobject.Mobject { return []mobject.Mobject{a.target} }

// PopIn scales the target from 0 → 1 with an EaseOutBack overshoot,
// for snappy "boop, here it is" appearances.
func PopIn(target Scaler, duration time.Duration) Animation {
	return &scaleAnim{
		target: target, from: 0, to: 1,
		dur: duration, ease: easing.OutBack,
	}
}

// Scaler is implemented by mobjects whose visual scale can be set.
type Scaler interface {
	mobject.Mobject
	SetVisualScale(s float64)
}

type scaleAnim struct {
	target   Scaler
	from, to float64
	dur      time.Duration
	ease     EasingFunc
}

func (a *scaleAnim) Apply(t float64) {
	t = clamp01(t)
	s := a.from + (a.to-a.from)*a.ease(t)
	a.target.SetVisualScale(s)
}
func (a *scaleAnim) Duration() time.Duration    { return a.dur }
func (a *scaleAnim) Targets() []mobject.Mobject { return []mobject.Mobject{a.target} }

// Flash temporarily changes the stroke color and fades back. Useful
// for "this node just received a packet" emphasis.
func Flash(target mobject.Mobject, flashColor interface {
	RGBA() (uint32, uint32, uint32, uint32)
}, duration time.Duration) Animation {
	originalStroke := target.Style().StrokeColor
	originalRoughness := target.Style().Roughness
	_ = originalRoughness
	return &flashAnim{
		target:         target,
		flashColor:     flashColor,
		originalStroke: originalStroke,
		dur:            duration,
	}
}

type flashAnim struct {
	target     mobject.Mobject
	flashColor interface {
		RGBA() (uint32, uint32, uint32, uint32)
	}
	originalStroke interface {
		RGBA() (uint32, uint32, uint32, uint32)
	}
	dur time.Duration
}

func (a *flashAnim) Apply(t float64) {
	t = clamp01(t)
	st := a.target.Style()
	// peak at t=0.5, return to original at t=1
	w := 1 - 2*absFloat(t-0.5)
	if w <= 0 {
		st.StrokeColor = a.originalStroke
		return
	}
	st.StrokeColor = blendColors(a.originalStroke, a.flashColor, w)
}
func (a *flashAnim) Duration() time.Duration    { return a.dur }
func (a *flashAnim) Targets() []mobject.Mobject { return []mobject.Mobject{a.target} }

// Wait is a no-op animation that holds the scene for a duration.
func Wait(d time.Duration) Animation {
	return &waitAnim{dur: d}
}

type waitAnim struct{ dur time.Duration }

func (w *waitAnim) Apply(float64)              {}
func (w *waitAnim) Duration() time.Duration    { return w.dur }
func (w *waitAnim) Targets() []mobject.Mobject { return nil }

// SetStyle is a small "step" animation that applies a style override
// at t≥0 and reverts to the original at t≥1. Useful as a building
// block; prefer MorphStyle for tweened style transitions.
func SetStyle(target mobject.Mobject, s style.Style, hold time.Duration) Animation {
	return &setStyleAnim{target: target, override: s, original: *target.Style(), dur: hold}
}

type setStyleAnim struct {
	target   mobject.Mobject
	override style.Style
	original style.Style
	dur      time.Duration
}

func (a *setStyleAnim) Apply(t float64) {
	t = clamp01(t)
	if t < 1 {
		a.target.SetStyle(a.override)
	} else {
		a.target.SetStyle(a.original)
	}
}
func (a *setStyleAnim) Duration() time.Duration    { return a.dur }
func (a *setStyleAnim) Targets() []mobject.Mobject { return []mobject.Mobject{a.target} }

// --- helpers ---

func clamp01(t float64) float64 {
	if t < 0 {
		return 0
	}
	if t > 1 {
		return 1
	}
	return t
}

func absFloat(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

// blendColors linearly interpolates between a and b at weight w∈[0,1].
// Both colors are converted through their RGBA() representation.
func blendColors(a, b interface {
	RGBA() (uint32, uint32, uint32, uint32)
}, w float64) interface {
	RGBA() (uint32, uint32, uint32, uint32)
} {
	if a == nil {
		return b
	}
	if b == nil {
		return a
	}
	ar, ag, ab, aa := a.RGBA()
	br, bg, bb, ba := b.RGBA()
	return colorRGBA{
		r: uint8((float64(ar>>8)*(1-w) + float64(br>>8)*w)),
		g: uint8((float64(ag>>8)*(1-w) + float64(bg>>8)*w)),
		b: uint8((float64(ab>>8)*(1-w) + float64(bb>>8)*w)),
		a: uint8((float64(aa>>8)*(1-w) + float64(ba>>8)*w)),
	}
}

// colorRGBA satisfies image/color.Color without importing it directly.
type colorRGBA struct{ r, g, b, a uint8 }

func (c colorRGBA) RGBA() (uint32, uint32, uint32, uint32) {
	return uint32(c.r) * 0x101, uint32(c.g) * 0x101, uint32(c.b) * 0x101, uint32(c.a) * 0x101
}
