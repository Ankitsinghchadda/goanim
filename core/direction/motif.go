package direction

import (
	"math"
	"time"

	"github.com/ankitsinghchadda/goanim/core/animation"
	"github.com/ankitsinghchadda/goanim/core/animation/easing"
	"github.com/ankitsinghchadda/goanim/core/geometry"
	"github.com/ankitsinghchadda/goanim/core/mobject"
	"github.com/ankitsinghchadda/goanim/core/render"
	"github.com/ankitsinghchadda/goanim/core/style"
)

// Motif is a Phase-9 decorative element whose job is to ESTABLISH
// CONTEXT for a moment, not carry information. The 100-player ring
// in BGMI's opener is the canonical example: dramatic entrance,
// brief hold, then either dissolves or recedes so the diagram that
// follows can be the focus.
//
// The contract is "introduce, hold, get out of the way." A motif
// that persists at full opacity for an entire scene becomes noise.
// Every motif comes with Enter and Exit animations encoded so the
// author doesn't have to remember to dismiss it.
type Motif interface {
	mobject.Mobject

	// EnterAnimation brings the motif on screen. Authors usually run
	// this for ~1-2 seconds early in a moment.
	EnterAnimation(duration time.Duration) animation.Animation

	// ExitAnimation moves the motif out of the way. ExitDissolve fades
	// it out entirely; ExitRecede dims + shrinks to background so the
	// motif still provides context but doesn't dominate; ExitCompose
	// settles into a target position/scale to BECOME part of the
	// scene's diagram.
	ExitAnimation(style ExitStyle, duration time.Duration) animation.Animation
}

// ExitStyle picks how a motif leaves a moment.
type ExitStyle uint8

const (
	// ExitDissolve fades the motif to opacity 0. Use when the next
	// scene doesn't need the motif as background anymore.
	ExitDissolve ExitStyle = iota

	// ExitRecede dims the motif to ~20% opacity AND shrinks it to ~60%
	// of its enter scale, so it's still visible as ambient context
	// while the foreground diagram takes over. The Phase-9 "get out
	// of the way without leaving entirely" pattern.
	ExitRecede

	// ExitCompose holds the motif at full presence — for the rare case
	// where the motif IS the moment (e.g. a final summary that's
	// just the player ring with a caption). Use sparingly.
	ExitCompose
)

// ----------------------------------------------------------------------------
// PlayerRing — N dots scattered in a noisy ring around a center point.
// Mirrors the BGMI opener visual. Each dot fades in via a staggered
// outward burst from the center, hangs at its ring position, then
// either dissolves or recedes (dim + shrink).
// ----------------------------------------------------------------------------

// PlayerRing constructs a Motif of `count` small dots arranged in a
// noisy ring around the origin. The ring's natural radius is
// `radius`; the noise amount is ~30px. Use PlayerRing(100) for the
// BGMI pattern.
func PlayerRing(count int, radius float64) Motif {
	return newPlayerRing(count, radius)
}

type playerRing struct {
	*mobject.Group
	count  int
	radius float64
	dots   []*mobject.Ellipse
	cx, cy float64
	// scale and opacity are updated by Enter/Exit animations. The
	// Render method consults them when drawing each dot.
	scale   float64
	opacity float64
}

func newPlayerRing(count int, radius float64) *playerRing {
	pr := &playerRing{
		Group:   mobject.NewGroup(0x7100),
		count:   count,
		radius:  radius,
		scale:   1,
		opacity: 1,
	}
	pr.dots = make([]*mobject.Ellipse, 0, count)
	for i := 0; i < count; i++ {
		dot := mobject.NewEllipse(int64(0x7100+i), 8, 8)
		x, y := pr.dotPos(i)
		dot.MoveTo(x, y)
		pr.dots = append(pr.dots, dot)
	}
	return pr
}

func (p *playerRing) dotPos(i int) (float64, float64) {
	theta := 2 * math.Pi * float64(i) / float64(p.count)
	r := p.radius + float64(i%5)*30
	return p.cx + r*math.Cos(theta)*p.scale, p.cy + r*math.Sin(theta)*p.scale
}

func (p *playerRing) Bounds() geometry.Rect {
	r := p.radius * p.scale * 1.2
	return geometry.RectFromCenter(geometry.Pt(p.cx, p.cy), 2*r, 2*r)
}

func (p *playerRing) Children() []mobject.Mobject {
	out := make([]mobject.Mobject, len(p.dots))
	for i, d := range p.dots {
		out[i] = d
	}
	return out
}

func (p *playerRing) Position() (float64, float64) { return p.cx, p.cy }
func (p *playerRing) SetPosition(x, y float64) {
	p.cx, p.cy = x, y
	p.relayoutDots()
}

func (p *playerRing) relayoutDots() {
	for i, d := range p.dots {
		x, y := p.dotPos(i)
		d.MoveTo(x, y)
	}
}

func (p *playerRing) Render(rd render.Renderer, ctx style.Context) {
	if p.opacity <= 0 {
		return
	}
	// Apply opacity to each dot's style before rendering.
	op := p.opacity
	for _, d := range p.dots {
		st := *d.Style()
		st.Opacity = &op
		d.SetStyle(st)
		d.Render(rd, ctx)
	}
}

// EnterAnimation: dots burst outward from the center, staggered.
// Visually: from radius 0 to full radius over the duration, with a
// per-dot delay so they appear to scatter outward in sequence rather
// than all at once.
func (p *playerRing) EnterAnimation(d time.Duration) animation.Animation {
	return &ringEnterAnim{p: p, dur: d}
}

// ExitAnimation produces the appropriate teardown for the style.
func (p *playerRing) ExitAnimation(s ExitStyle, d time.Duration) animation.Animation {
	switch s {
	case ExitRecede:
		return &ringRecedeAnim{p: p, dur: d}
	case ExitCompose:
		// Hold in place; no-op animation of the duration.
		return &ringHoldAnim{p: p, dur: d}
	default: // ExitDissolve
		return &ringDissolveAnim{p: p, dur: d}
	}
}

// --- Enter ------------------------------------------------------------------

type ringEnterAnim struct {
	p   *playerRing
	dur time.Duration
}

func (a *ringEnterAnim) Apply(t float64) {
	if t < 0 {
		t = 0
	}
	if t > 1 {
		t = 1
	}
	// scale ramps from 0 to 1 across the FIRST 80% of t — then holds.
	// Opacity ramps with a slight overshoot via OutCubic for energy.
	tScale := t / 0.8
	if tScale > 1 {
		tScale = 1
	}
	a.p.scale = easing.OutCubic(tScale)
	a.p.opacity = easing.OutCubic(t)
	a.p.relayoutDots()
}
func (a *ringEnterAnim) Duration() time.Duration    { return a.dur }
func (a *ringEnterAnim) Targets() []mobject.Mobject { return []mobject.Mobject{a.p} }

// --- Recede ----------------------------------------------------------------

type ringRecedeAnim struct {
	p   *playerRing
	dur time.Duration
}

func (a *ringRecedeAnim) Apply(t float64) {
	if t < 0 {
		t = 0
	}
	if t > 1 {
		t = 1
	}
	e := easing.InOutCubic(t)
	// scale 1.0 → 0.6, opacity 1.0 → 0.2 — recede pattern.
	a.p.scale = 1.0 + (0.6-1.0)*e
	a.p.opacity = 1.0 + (0.2-1.0)*e
	a.p.relayoutDots()
}
func (a *ringRecedeAnim) Duration() time.Duration    { return a.dur }
func (a *ringRecedeAnim) Targets() []mobject.Mobject { return []mobject.Mobject{a.p} }

// --- Dissolve --------------------------------------------------------------

type ringDissolveAnim struct {
	p   *playerRing
	dur time.Duration
}

func (a *ringDissolveAnim) Apply(t float64) {
	if t < 0 {
		t = 0
	}
	if t > 1 {
		t = 1
	}
	a.p.opacity = 1.0 - easing.InCubic(t)
	a.p.relayoutDots()
}
func (a *ringDissolveAnim) Duration() time.Duration    { return a.dur }
func (a *ringDissolveAnim) Targets() []mobject.Mobject { return []mobject.Mobject{a.p} }

// --- Hold (ExitCompose) ----------------------------------------------------

type ringHoldAnim struct {
	p   *playerRing
	dur time.Duration
}

func (a *ringHoldAnim) Apply(float64)              {} // no-op
func (a *ringHoldAnim) Duration() time.Duration    { return a.dur }
func (a *ringHoldAnim) Targets() []mobject.Mobject { return []mobject.Mobject{a.p} }

// IsConstantAnim marks ringHoldAnim as a constant animation so the
// scene player skips per-tick re-rendering during the hold.
func (a *ringHoldAnim) IsConstantAnim() {}
