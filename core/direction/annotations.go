package direction

import (
	"image/color"
	"time"

	"github.com/ankitsinghchadda/goanim/core/animation"
	"github.com/ankitsinghchadda/goanim/core/animation/easing"
	"github.com/ankitsinghchadda/goanim/core/geometry"
	"github.com/ankitsinghchadda/goanim/core/mobject"
	"github.com/ankitsinghchadda/goanim/core/render"
	"github.com/ankitsinghchadda/goanim/core/rough"
	"github.com/ankitsinghchadda/goanim/core/style"
)

// UnderlineOn draws a hand-drawn underline beneath the target's
// bounding box and reveals it over duration. Unlike LaserPointer,
// the underline IS a diagram element — it respects the scene's
// sloppiness. A sketchy scene gets a wobbly underline, a crisp scene
// gets a straight one.
//
//	scene.Play(direction.UnderlineOn(label, 500*time.Millisecond))
//	scene.Play(direction.Pause(2*time.Second))
//	scene.Play(direction.UnderlineOff(label, 300*time.Millisecond))
//
// The underline target tracks the mobject's bounds at render time, so
// if the target moves the underline tracks with it. Sized to span the
// target's width with a small inset on each side; positioned just
// below the target's bottom edge.
func UnderlineOn(target mobject.Mobject, duration time.Duration) animation.Animation {
	u := newUnderline(target)
	u.SetReveal(0)
	return &annotationReveal{m: u, from: 0, to: 1, dur: duration, ease: easing.OutCubic}
}

// UnderlineOff fades a previously-drawn underline. Pair with the
// SAME mobject reference that UnderlineOn returned via the Targets()
// list — easier to use the variant that takes the annotation back:
// see UnderlineOnReturning if you need explicit teardown control.
func UnderlineOff(target mobject.Mobject, duration time.Duration) animation.Animation {
	u := newUnderline(target)
	u.SetReveal(1)
	return &annotationReveal{m: u, from: 1, to: 0, dur: duration, ease: easing.InCubic}
}

// CircleAround draws a hand-drawn ellipse encompassing the target with
// padding, revealed over duration. Same sloppiness contract as
// UnderlineOn — respects scene style.
//
//	scene.Play(direction.CircleAround(database, 700*time.Millisecond))
func CircleAround(target mobject.Mobject, duration time.Duration) animation.Animation {
	c := newCircleAround(target)
	c.SetReveal(0)
	return &annotationReveal{m: c, from: 0, to: 1, dur: duration, ease: easing.OutCubic}
}

// CircleAroundOff fades a previously-drawn circle annotation.
func CircleAroundOff(target mobject.Mobject, duration time.Duration) animation.Animation {
	c := newCircleAround(target)
	c.SetReveal(1)
	return &annotationReveal{m: c, from: 1, to: 0, dur: duration, ease: easing.InCubic}
}

// CalloutPosition controls where a Callout sits relative to its
// target. The arrow always points from the pill toward the target.
type CalloutPosition uint8

const (
	CalloutAbove CalloutPosition = iota
	CalloutBelow
	CalloutLeft
	CalloutRight
)

// CalloutTiming subdivides a callout's lifecycle into reveal, hold,
// and fade phases. The default (used by Callout) splits as 1/4 :
// 1/2 : 1/4. CalloutWithTiming overrides per-call.
type CalloutTiming struct {
	Reveal time.Duration
	Hold   time.Duration
	Fade   time.Duration
}

// Callout pops a short text annotation in a sketchy pill with a thin
// arrow pointing at target. Appears, holds, fades. Default timing is
// the prompt's 1/4 reveal : 1/2 hold : 1/4 fade split.
//
//	scene.Play(direction.Callout(database,
//	    "single point of failure!",
//	    direction.CalloutBelow,
//	    2500*time.Millisecond,
//	))
func Callout(target mobject.Mobject, text string, pos CalloutPosition, totalDuration time.Duration) animation.Animation {
	q := totalDuration / 4
	return CalloutWithTiming(target, text, pos, CalloutTiming{
		Reveal: q,
		Hold:   2 * q,
		Fade:   q,
	})
}

// CalloutWithTiming gives the creator explicit control over each
// phase of the callout's lifecycle.
func CalloutWithTiming(target mobject.Mobject, text string, pos CalloutPosition, timing CalloutTiming) animation.Animation {
	c := newCallout(target, text, pos)
	c.SetReveal(0)
	return animation.Sequence(
		&annotationReveal{m: c, from: 0, to: 1, dur: timing.Reveal, ease: easing.OutCubic},
		hold(c, timing.Hold),
		&annotationReveal{m: c, from: 1, to: 0, dur: timing.Fade, ease: easing.InCubic},
	)
}

// hold keeps a target at its current reveal during a Pause-duration
// segment. Used by composite annotations whose middle phase is a hold.
func hold(m mobject.Mobject, d time.Duration) animation.Animation {
	return &annotationHold{m: m, dur: d}
}

type annotationHold struct {
	m   mobject.Mobject
	dur time.Duration
}

func (h *annotationHold) Apply(float64)              {} // no-op; the previous reveal value persists
func (h *annotationHold) Duration() time.Duration    { return h.dur }
func (h *annotationHold) Targets() []mobject.Mobject { return []mobject.Mobject{h.m} }

// annotationReveal animates the reveal fraction of an annotation
// mobject. The mobject must satisfy animation.Revealer (SetReveal).
type annotationReveal struct {
	m        mobject.Mobject
	from, to float64
	dur      time.Duration
	ease     animation.EasingFunc
}

func (a *annotationReveal) Apply(t float64) {
	t = clamp01(t)
	val := a.from + (a.to-a.from)*a.ease(t)
	if r, ok := a.m.(interface{ SetReveal(float64) }); ok {
		r.SetReveal(val)
	}
}
func (a *annotationReveal) Duration() time.Duration    { return a.dur }
func (a *annotationReveal) Targets() []mobject.Mobject { return []mobject.Mobject{a.m} }

// ----- underline mobject ---------------------------------------------------

type underline struct {
	*mobject.Group
	target  mobject.Mobject
	reveal  float64
	seed    int64
	padding float64
	offset  float64 // distance below target's bottom edge
}

func newUnderline(target mobject.Mobject) *underline {
	return &underline{
		Group:   mobject.NewGroup(0xDEADBEEF ^ targetSeed(target)),
		target:  target,
		reveal:  1,
		seed:    0xDEADBEEF ^ targetSeed(target),
		padding: 8,
		offset:  10,
	}
}

func (u *underline) Bounds() geometry.Rect {
	b := u.target.Bounds()
	w := b.Max.X - b.Min.X + 2*u.padding
	return geometry.RectFromCenter(
		geometry.Pt((b.Min.X+b.Max.X)/2, b.Min.Y-u.offset),
		w, 6,
	)
}
func (u *underline) Children() []mobject.Mobject { return nil }
func (u *underline) SetReveal(t float64)         { u.reveal = clamp01(t) }
func (u *underline) Render(rd render.Renderer, ctx style.Context) {
	if u.reveal <= 0 {
		return
	}
	eff := ctx.Resolve(*u.Group.Style())
	tok := style.TokensFor(eff)
	b := u.target.Bounds()
	span := (b.Max.X - b.Min.X) * u.reveal
	cx := (b.Min.X + b.Max.X) / 2
	from := geometry.Pt(cx-span/2, b.Min.Y-u.offset)
	to := geometry.Pt(cx+span/2, b.Min.Y-u.offset)
	stroke := style.PathStyleStroke(eff, tok)
	stroke.StrokeWidth *= 1.4 // a touch bolder than typical strokes — underlines should read confidently
	if u.reveal < 1 {
		stroke.Stroke = style.ApplyOpacity(eff.StrokeColor, tok.OpacityScale*u.reveal)
	}
	if tok.Roughness == 0 {
		rd.DrawPath(geometry.LinePath(from, to), stroke)
		return
	}
	opts := style.RoughOptions(eff, tok, u.seed)
	opts.Roughness = tok.Roughness * 0.7
	opts.DisableMultiStroke = true
	rd.DrawPath(rough.RoughLine(from, to, opts), stroke)
}

// ----- circle-around mobject -----------------------------------------------

type circleAround struct {
	*mobject.Group
	target  mobject.Mobject
	reveal  float64
	seed    int64
	padding float64
}

func newCircleAround(target mobject.Mobject) *circleAround {
	return &circleAround{
		Group:   mobject.NewGroup(0xC1AC1E ^ targetSeed(target)),
		target:  target,
		reveal:  1,
		seed:    0xC1AC1E ^ targetSeed(target),
		padding: 20,
	}
}

func (c *circleAround) Bounds() geometry.Rect {
	b := c.target.Bounds()
	w := b.Max.X - b.Min.X + 2*c.padding
	h := b.Max.Y - b.Min.Y + 2*c.padding
	return geometry.RectFromCenter(geometry.Pt((b.Min.X+b.Max.X)/2, (b.Min.Y+b.Max.Y)/2), w, h)
}
func (c *circleAround) Children() []mobject.Mobject { return nil }
func (c *circleAround) SetReveal(t float64)         { c.reveal = clamp01(t) }
func (c *circleAround) Render(rd render.Renderer, ctx style.Context) {
	if c.reveal <= 0 {
		return
	}
	eff := ctx.Resolve(*c.Group.Style())
	tok := style.TokensFor(eff)
	b := c.target.Bounds()
	cx := (b.Min.X + b.Max.X) / 2
	cy := (b.Min.Y + b.Max.Y) / 2
	rx := (b.Max.X-b.Min.X)/2 + c.padding
	ry := (b.Max.Y-b.Min.Y)/2 + c.padding
	stroke := style.PathStyleStroke(eff, tok)
	stroke.StrokeWidth *= 1.25
	stroke.Stroke = style.ApplyOpacity(eff.StrokeColor, tok.OpacityScale*c.reveal)
	if tok.Roughness == 0 {
		rd.DrawPath(geometry.EllipsePath(cx, cy, rx, ry), stroke)
		return
	}
	opts := style.RoughOptions(eff, tok, c.seed)
	rd.DrawPath(rough.RoughEllipse(cx, cy, rx, ry, opts), stroke)
}

// ----- callout mobject -----------------------------------------------------

type callout struct {
	*mobject.Group
	target mobject.Mobject
	text   string
	pos    CalloutPosition
	reveal float64
	seed   int64
}

func newCallout(target mobject.Mobject, text string, pos CalloutPosition) *callout {
	return &callout{
		Group:  mobject.NewGroup(0xCA110 ^ targetSeed(target)),
		target: target,
		text:   text,
		pos:    pos,
		reveal: 1,
		seed:   0xCA110 ^ targetSeed(target),
	}
}

// pillRect returns the pill background's center, width, height — sized
// to the text plus padding, positioned relative to the target's bounds
// according to c.pos.
func (c *callout) pillRect() (cx, cy, w, h float64) {
	b := c.target.Bounds()
	// Estimate text width: 11px per character at FontMedium. Cheap
	// approximation — close enough for positioning since the pill is
	// padded on both sides.
	w = float64(len(c.text))*11 + 36
	h = 44
	const gap = 24
	mx := (b.Min.X + b.Max.X) / 2
	my := (b.Min.Y + b.Max.Y) / 2
	switch c.pos {
	case CalloutAbove:
		cx = mx
		cy = b.Max.Y + gap + h/2
	case CalloutBelow:
		cx = mx
		cy = b.Min.Y - gap - h/2
	case CalloutLeft:
		cx = b.Min.X - gap - w/2
		cy = my
	case CalloutRight:
		cx = b.Max.X + gap + w/2
		cy = my
	}
	return
}

func (c *callout) Bounds() geometry.Rect {
	cx, cy, w, h := c.pillRect()
	return geometry.RectFromCenter(geometry.Pt(cx, cy), w+20, h+20)
}
func (c *callout) Children() []mobject.Mobject { return nil }
func (c *callout) SetReveal(t float64)         { c.reveal = clamp01(t) }
func (c *callout) Render(rd render.Renderer, ctx style.Context) {
	if c.reveal <= 0 {
		return
	}
	eff := ctx.Resolve(*c.Group.Style())
	tok := style.TokensFor(eff)
	cx, cy, w, h := c.pillRect()
	rev := c.reveal

	// 1) Knock-out background — paint the pill region in the icon
	//    fill color (or scene bg) so the text reads clearly over busy
	//    diagrams behind.
	var bgCol color.Color
	if eff.FillColor != nil {
		bgCol = eff.FillColor
	} else {
		bgCol = ctx.BgColor
	}
	if bgCol == nil {
		bgCol = color.RGBA{0xFF, 0xFF, 0xFF, 0xFF}
	}
	// Inflate slightly so the pill stroke has clean room.
	pillBgPath := geometry.RectanglePath(cx-w/2-2, cy-h/2-2, w+4, h+4, tok.CornerRadius+6)
	rd.DrawPath(pillBgPath, render.PathStyle{
		Fill: style.ApplyOpacity(bgCol, rev),
	})

	// 2) Pill outline — respects sloppiness.
	stroke := style.PathStyleStroke(eff, tok)
	stroke.StrokeWidth *= 1.05
	stroke.Stroke = style.ApplyOpacity(eff.StrokeColor, tok.OpacityScale*rev)
	var pillStrokePath *geometry.Path
	if tok.Roughness == 0 {
		pillStrokePath = geometry.RectanglePath(cx-w/2, cy-h/2, w, h, tok.CornerRadius+6)
	} else {
		opts := style.RoughOptions(eff, tok, c.seed)
		opts.Roughness = tok.Roughness * 0.6
		opts.DisableMultiStroke = true
		pillStrokePath = rough.RoughRectangle(cx-w/2, cy-h/2, w, h, opts)
	}
	rd.DrawPath(pillStrokePath, stroke)

	// 3) Arrow from pill toward target's nearest edge.
	tb := c.target.Bounds()
	var from, to geometry.Point
	switch c.pos {
	case CalloutAbove:
		from = geometry.Pt(cx, cy-h/2)
		to = geometry.Pt(cx, tb.Max.Y+2)
	case CalloutBelow:
		from = geometry.Pt(cx, cy+h/2)
		to = geometry.Pt(cx, tb.Min.Y-2)
	case CalloutLeft:
		from = geometry.Pt(cx+w/2, cy)
		to = geometry.Pt(tb.Min.X-2, cy)
	case CalloutRight:
		from = geometry.Pt(cx-w/2, cy)
		to = geometry.Pt(tb.Max.X+2, cy)
	}
	if tok.Roughness == 0 {
		rd.DrawPath(geometry.LinePath(from, to), stroke)
	} else {
		opts := style.RoughOptions(eff, tok, c.seed+99)
		opts.DisableMultiStroke = true
		opts.Roughness = tok.Roughness * 0.6
		rd.DrawPath(rough.RoughLine(from, to, opts), stroke)
	}

	// 4) Text — centered in the pill.
	face := ctx.FontFace(eff.FontFamily)
	rd.DrawText(c.text, cx, cy, render.TextStyle{
		Face:     face,
		Size:     style.TokensFor(eff).FontSizePx * 0.7,
		Color:    style.ApplyOpacity(eff.StrokeColor, rev),
		Align:    render.AlignCenter,
		Baseline: render.BaselineMiddle,
	})
}

// targetSeed derives a stable hash from a target mobject's identity so
// the wobble of an underline / circle / callout is reproducible
// per-target. We use Seed() if the mobject implements it; otherwise 0.
func targetSeed(m mobject.Mobject) int64 {
	if s, ok := m.(interface{ Seed() int64 }); ok {
		return s.Seed()
	}
	return 0
}
