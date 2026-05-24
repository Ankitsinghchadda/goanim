// Package mascot is a tiny cartoon "host" character for explainer
// videos. The mascot is a round head with two dot eyes and a mouth
// path that morphs between named expressions: happy, neutral, thinking,
// wow, oof. Optional speech bubble accompanies a one-liner.
//
// The whole face is a single mobject — animations on it work like any
// other (FadeIn, Shift, Flash). SetExpression is cheap (rebuilds the
// mouth path only).
package mascot

import (
	"image/color"
	"math"

	"github.com/ankitsinghchadda/goanim/core/geometry"
	"github.com/ankitsinghchadda/goanim/core/mobject"
	"github.com/ankitsinghchadda/goanim/core/render"
	"github.com/ankitsinghchadda/goanim/core/style"
)

// Expression names a face preset.
type Expression uint8

const (
	ExprNeutral  Expression = iota // straight line mouth
	ExprHappy                      // smile arc
	ExprThinking                   // small line + raised eyebrow
	ExprWow                        // big "o" mouth, wide eyes
	ExprOof                        // wavy frown, narrow eyes (worried/burnt out)
)

// Mascot is a round-headed cartoon face. Defaults to ExprHappy.
type Mascot struct {
	seed   int64
	cx, cy float64
	radius float64
	expr   Expression
	style  style.Style
	reveal float64
	body   color.Color // head fill color; defaults to a warm yellow
	stroke color.Color // outline color; defaults to ink
}

// NewMascot constructs a face centered at (0,0) with the given radius.
// radius=80 reads as a corner-sized avatar at 1920×1080.
func NewMascot(seed int64, radius float64) *Mascot {
	return &Mascot{
		seed:   seed,
		radius: radius,
		expr:   ExprHappy,
		reveal: 1,
	}
}

func (m *Mascot) MoveTo(x, y float64) *Mascot {
	m.cx, m.cy = x, y
	return m
}
func (m *Mascot) SetPosition(x, y float64)     { m.cx, m.cy = x, y }
func (m *Mascot) Position() (float64, float64) { return m.cx, m.cy }
func (m *Mascot) SetExpression(e Expression)   { m.expr = e }
func (m *Mascot) Expression() Expression       { return m.expr }
func (m *Mascot) WithExpression(e Expression) *Mascot {
	m.expr = e
	return m
}
func (m *Mascot) WithBodyColor(c color.Color) *Mascot   { m.body = c; return m }
func (m *Mascot) WithStrokeColor(c color.Color) *Mascot { m.stroke = c; return m }
func (m *Mascot) Bounds() geometry.Rect {
	return geometry.RectFromCenter(geometry.Pt(m.cx, m.cy), m.radius*2.2, m.radius*2.4)
}
func (m *Mascot) Children() []mobject.Mobject { return nil }
func (m *Mascot) Seed() int64                 { return m.seed }
func (m *Mascot) Style() *style.Style         { return &m.style }
func (m *Mascot) SetStyle(s style.Style)      { m.style = s }
func (m *Mascot) SetReveal(t float64) {
	if t < 0 {
		t = 0
	}
	if t > 1 {
		t = 1
	}
	m.reveal = t
}
func (m *Mascot) Reveal() float64        { return m.reveal }
func (m *Mascot) SetVisualScale(float64) {}

func (m *Mascot) Render(r render.Renderer, ctx style.Context) {
	if m.reveal <= 0 {
		return
	}
	eff := ctx.Resolve(m.style)
	tok := style.TokensFor(eff)
	stroke := m.stroke
	if stroke == nil {
		stroke = eff.StrokeColor
	}
	body := m.body
	if body == nil {
		// Warm honey yellow — friendly, not aggressive.
		body = color.RGBA{0xFF, 0xD5, 0x4F, 0xFF}
	}
	op := tok.OpacityScale * m.reveal
	strokePS := render.PathStyle{
		Stroke:      style.ApplyOpacity(stroke, op),
		StrokeWidth: tok.StrokeWidthPx * 1.4,
		StrokeCap:   render.CapRound,
		StrokeJoin:  render.JoinRound,
	}
	fillPS := render.PathStyle{
		Fill: style.ApplyOpacity(body, op),
	}

	// --- Head -----------------------------------------------------
	head := geometry.EllipsePath(m.cx, m.cy, m.radius, m.radius*1.02)
	r.DrawPath(head, fillPS)
	r.DrawPath(head, strokePS)

	// --- Eyes -----------------------------------------------------
	// Eye geometry depends on the expression. Most expressions: two
	// small filled dots. ExprWow: bigger circles with a small white
	// reflection dot. ExprOof: narrow horizontal slits.
	eyeY := m.cy + m.radius*0.18
	eyeOffsetX := m.radius * 0.35
	eyeFill := render.PathStyle{Fill: style.ApplyOpacity(stroke, op)}
	switch m.expr {
	case ExprOof:
		// Narrow slits — eyes squeezed shut, "ugh".
		w := m.radius * 0.22
		h := m.radius * 0.04
		l := geometry.EllipsePath(m.cx-eyeOffsetX, eyeY, w, h)
		rr := geometry.EllipsePath(m.cx+eyeOffsetX, eyeY, w, h)
		r.DrawPath(l, eyeFill)
		r.DrawPath(rr, eyeFill)
	case ExprWow:
		w := m.radius * 0.16
		l := geometry.EllipsePath(m.cx-eyeOffsetX, eyeY, w, w)
		rr := geometry.EllipsePath(m.cx+eyeOffsetX, eyeY, w, w)
		r.DrawPath(l, eyeFill)
		r.DrawPath(rr, eyeFill)
		// White highlight reflections.
		hi := render.PathStyle{Fill: style.ApplyOpacity(color.White, op)}
		r.DrawPath(geometry.EllipsePath(m.cx-eyeOffsetX-w*0.4, eyeY+w*0.4, w*0.3, w*0.3), hi)
		r.DrawPath(geometry.EllipsePath(m.cx+eyeOffsetX-w*0.4, eyeY+w*0.4, w*0.3, w*0.3), hi)
	default:
		w := m.radius * 0.09
		l := geometry.EllipsePath(m.cx-eyeOffsetX, eyeY, w, w*1.1)
		rr := geometry.EllipsePath(m.cx+eyeOffsetX, eyeY, w, w*1.1)
		r.DrawPath(l, eyeFill)
		r.DrawPath(rr, eyeFill)
	}

	// --- Eyebrow for thinking -------------------------------------
	if m.expr == ExprThinking {
		// One raised eyebrow above the right eye.
		browY := eyeY + m.radius*0.18
		browL := geometry.Pt(m.cx+eyeOffsetX-m.radius*0.08, browY)
		browR := geometry.Pt(m.cx+eyeOffsetX+m.radius*0.18, browY+m.radius*0.12)
		brow := geometry.LinePath(browL, browR)
		ps := strokePS
		ps.StrokeWidth = tok.StrokeWidthPx
		r.DrawPath(brow, ps)
	}

	// --- Mouth ---------------------------------------------------
	mouth := m.mouthPath()
	if mouth != nil {
		ps := strokePS
		// "wow" mouth is round/open — draw with fill if it's a closed shape.
		if m.expr == ExprWow {
			closedPS := render.PathStyle{
				Fill: style.ApplyOpacity(color.RGBA{0x2B, 0x2B, 0x2B, 0xFF}, op),
			}
			r.DrawPath(mouth, closedPS)
		}
		r.DrawPath(mouth, ps)
	}
}

// mouthPath builds the mouth geometry for the current expression.
// Returns nil if the expression has no separate mouth shape.
func (m *Mascot) mouthPath() *geometry.Path {
	mouthY := m.cy - m.radius*0.30
	switch m.expr {
	case ExprHappy:
		// Smile arc — cubic Bezier curving downward then back up.
		p := geometry.NewPath()
		left := geometry.Pt(m.cx-m.radius*0.35, mouthY)
		right := geometry.Pt(m.cx+m.radius*0.35, mouthY)
		c1 := geometry.Pt(m.cx-m.radius*0.18, mouthY-m.radius*0.30)
		c2 := geometry.Pt(m.cx+m.radius*0.18, mouthY-m.radius*0.30)
		p.MoveTo(left.X, left.Y)
		p.CurveTo(c1.X, c1.Y, c2.X, c2.Y, right.X, right.Y)
		return p
	case ExprThinking:
		// Short horizontal line — neutral pondering.
		return geometry.LinePath(
			geometry.Pt(m.cx-m.radius*0.15, mouthY),
			geometry.Pt(m.cx+m.radius*0.15, mouthY),
		)
	case ExprWow:
		// Open "O" mouth.
		return geometry.EllipsePath(m.cx, mouthY, m.radius*0.18, m.radius*0.22)
	case ExprOof:
		// Wavy frown — a small zigzag implying gritted teeth / discomfort.
		p := geometry.NewPath()
		x0 := m.cx - m.radius*0.35
		x1 := m.cx + m.radius*0.35
		y := mouthY + m.radius*0.05 // frown sits lower
		const steps = 5
		p.MoveTo(x0, y)
		for i := 1; i <= steps; i++ {
			t := float64(i) / float64(steps)
			x := x0 + t*(x1-x0)
			yo := y
			if i%2 == 0 {
				yo += m.radius * 0.05
			} else {
				yo -= m.radius * 0.05
			}
			p.LineTo(x, yo)
		}
		return p
	default:
		return geometry.LinePath(
			geometry.Pt(m.cx-m.radius*0.22, mouthY),
			geometry.Pt(m.cx+m.radius*0.22, mouthY),
		)
	}
}

// SpeechBubble draws a labeled callout bubble pointing at the mascot's
// head from a side. Use it for one-liners that lock voice-over phrasing
// to the visuals.
type SpeechBubble struct {
	*mobject.Group
	mascot *Mascot
	text   *mobject.Text
	side   Side
	w, h   float64
	style  style.Style
	reveal float64
}

// Side names where the bubble sits relative to the mascot.
type Side uint8

const (
	SideRight Side = iota
	SideLeft
	SideTop
)

// NewSpeechBubble builds a bubble attached to m. The bubble label uses
// the scene's hand-drawn font by default.
func NewSpeechBubble(seed int64, m *Mascot, side Side, label string) *SpeechBubble {
	txt := mobject.NewText(seed+1, label)
	sb := &SpeechBubble{
		Group:  mobject.NewGroup(seed),
		mascot: m,
		text:   txt,
		side:   side,
		w:      540,
		h:      170,
		reveal: 1,
	}
	sb.Group.Add(sb.text)
	sb.layout()
	return sb
}

// WithSize overrides the bubble's pixel dimensions.
func (b *SpeechBubble) WithSize(w, h float64) *SpeechBubble {
	b.w = w
	b.h = h
	b.layout()
	return b
}

// WithTextStyle overrides the inner text style.
func (b *SpeechBubble) WithTextStyle(s style.Style) *SpeechBubble {
	b.text.SetStyle(s)
	return b
}

func (b *SpeechBubble) layout() {
	cx, cy := b.mascot.Position()
	radius := b.mascot.radius
	bx, by := cx, cy
	switch b.side {
	case SideRight:
		bx = cx + radius + b.w/2 + 30
	case SideLeft:
		bx = cx - radius - b.w/2 - 30
	case SideTop:
		by = cy + radius + b.h/2 + 30
	}
	b.text.MoveTo(bx, by)
}

func (b *SpeechBubble) Bounds() geometry.Rect {
	tx, ty := b.text.Bounds().Center().X, b.text.Bounds().Center().Y
	return geometry.RectFromCenter(geometry.Pt(tx, ty), b.w, b.h)
}
func (b *SpeechBubble) Children() []mobject.Mobject { return nil }
func (b *SpeechBubble) Seed() int64                 { return b.Group.Seed() }
func (b *SpeechBubble) Style() *style.Style         { return &b.style }
func (b *SpeechBubble) SetStyle(s style.Style)      { b.style = s }
func (b *SpeechBubble) SetReveal(t float64) {
	if t < 0 {
		t = 0
	}
	if t > 1 {
		t = 1
	}
	b.reveal = t
	b.text.SetReveal(t)
}
func (b *SpeechBubble) Reveal() float64        { return b.reveal }
func (b *SpeechBubble) SetVisualScale(float64) {}

func (b *SpeechBubble) Render(r render.Renderer, ctx style.Context) {
	if b.reveal <= 0 {
		return
	}
	eff := ctx.Resolve(b.style)
	tok := style.TokensFor(eff)
	op := tok.OpacityScale * b.reveal

	// Bubble body — a rounded rectangle.
	b.layout()
	tx, ty := b.text.Position()
	bg := geometry.RectanglePath(tx-b.w/2, ty-b.h/2, b.w, b.h, 22)
	bgFill := color.Color(color.RGBA{0xFF, 0xFF, 0xFF, 0xFF})
	r.DrawPath(bg, render.PathStyle{Fill: style.ApplyOpacity(bgFill, op)})
	r.DrawPath(bg, render.PathStyle{
		Stroke:      style.ApplyOpacity(eff.StrokeColor, op),
		StrokeWidth: tok.StrokeWidthPx,
		StrokeCap:   render.CapRound,
	})

	// Tail — a small triangle pointing from the bubble toward the mascot's head.
	cx, cy := b.mascot.Position()
	tail := geometry.NewPath()
	switch b.side {
	case SideRight:
		tip := geometry.Pt(cx+b.mascot.radius, cy+b.mascot.radius*0.1)
		base1 := geometry.Pt(tx-b.w/2+10, cy+30)
		base2 := geometry.Pt(tx-b.w/2+10, cy-30)
		tail.MoveTo(tip.X, tip.Y)
		tail.LineTo(base1.X, base1.Y)
		tail.LineTo(base2.X, base2.Y)
		tail.Close()
	case SideLeft:
		tip := geometry.Pt(cx-b.mascot.radius, cy+b.mascot.radius*0.1)
		base1 := geometry.Pt(tx+b.w/2-10, cy+30)
		base2 := geometry.Pt(tx+b.w/2-10, cy-30)
		tail.MoveTo(tip.X, tip.Y)
		tail.LineTo(base1.X, base1.Y)
		tail.LineTo(base2.X, base2.Y)
		tail.Close()
	case SideTop:
		tip := geometry.Pt(cx, cy+b.mascot.radius)
		base1 := geometry.Pt(tx-30, ty-b.h/2+10)
		base2 := geometry.Pt(tx+30, ty-b.h/2+10)
		tail.MoveTo(tip.X, tip.Y)
		tail.LineTo(base1.X, base1.Y)
		tail.LineTo(base2.X, base2.Y)
		tail.Close()
	}
	r.DrawPath(tail, render.PathStyle{Fill: style.ApplyOpacity(bgFill, op)})
	r.DrawPath(tail, render.PathStyle{
		Stroke:      style.ApplyOpacity(eff.StrokeColor, op),
		StrokeWidth: tok.StrokeWidthPx,
	})

	// Inner text.
	b.text.Render(r, ctx)
	_ = math.Pi
}
