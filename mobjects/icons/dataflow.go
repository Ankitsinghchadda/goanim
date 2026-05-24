package icons

import (
	"image/color"

	"github.com/ankitsinghchadda/goanim/core/geometry"
	"github.com/ankitsinghchadda/goanim/core/icon"
	"github.com/ankitsinghchadda/goanim/core/mobject"
	"github.com/ankitsinghchadda/goanim/core/render"
	"github.com/ankitsinghchadda/goanim/core/rough"
	"github.com/ankitsinghchadda/goanim/core/style"
)

// NewQueue is a horizontal FIFO: an outer rectangle subdivided into 5
// slots by 4 internal vertical lines. The leftmost slot is filled to
// indicate "next out."
//
// Hachure density is forced to Light because the slot dividers ARE
// the metaphor — a denser hatch overpowers them at Cartoonist
// sloppiness and the icon reads as a hatched rectangle.
func NewQueue(seed int64, label string) *icon.IconBase {
	const w, h = 280, 60
	ic := icon.New(seed, w, h, label)
	ic.Add(mobject.NewRectangle(seed, w, h))
	ic.Add(newQueueSlots(seed+201, w, h, 5, true /* horizontal */, true /* fillLeading */))
	withLightDensity(ic)
	return ic
}

// NewStack is the vertical version of Queue with the top slot marked
// as "top" — LIFO. Like Queue, density is forced Light so the
// horizontal slot dividers read through the hatch (the narrow body
// otherwise becomes a solid hatched column).
func NewStack(seed int64, label string) *icon.IconBase {
	const w, h = 80, 240
	ic := icon.New(seed, w, h, label)
	ic.Add(mobject.NewRectangle(seed, w, h))
	ic.Add(newQueueSlots(seed+201, w, h, 5, false, true))
	withLightDensity(ic)
	return ic
}

// NewMessageBroker resembles Queue but with input/output dots showing
// pub/sub fan-out.
//
// No leading-slot fill: a broker isn't a FIFO, the "next out" cue
// would be misleading. The slot dividers convey the messaging
// structure on their own.
//
// SidePorts emit their own per-dot knock-outs (see sidePorts.Render),
// so they remain a frame part. The dots themselves are positioned
// outside the body but their halos extend INTO the body, which is
// why per-dot knock-out is needed for the inner edges.
func NewMessageBroker(seed int64, label string) *icon.IconBase {
	const w, h = 240, 80
	ic := icon.New(seed, w, h, label)
	ic.Add(mobject.NewRectangle(seed, w, h))
	ic.Add(newQueueSlots(seed+201, w, h, 4, true, false /* no leading fill */))
	ic.Add(newSidePorts(seed+501, w, h))
	withLightDensity(ic)
	return ic
}

// withLightDensity forces DetailDensityLight on the icon's group
// style. Used by icons whose internal divider/structure conveys the
// concept and must survive Cartoonist cross-hatching.
func withLightDensity(ic *icon.IconBase) {
	s := *ic.Style()
	if s.DetailDensity == style.DetailDensityUnset {
		s.DetailDensity = style.DetailDensityLight
		ic.SetStyle(s)
	}
}

// queueSlots draws the internal dividers and (optionally) a filled
// "next out" indicator on the leading slot.
//
// fillLeading toggles the "next-out" shade. Queue/Stack want it as a
// FIFO/LIFO cue; Broker (pub/sub) does not have a leading slot and
// would be misled by the fill.
type queueSlots struct {
	*mobject.Group
	seed        int64
	w, h        float64
	slots       int
	horiz       bool
	fillLeading bool
	cx, cy      float64
	reveal      float64
}

func newQueueSlots(seed int64, w, h float64, slots int, horiz, fillLeading bool) *queueSlots {
	return &queueSlots{
		Group:       mobject.NewGroup(seed),
		seed:        seed,
		w:           w,
		h:           h,
		slots:       slots,
		horiz:       horiz,
		fillLeading: fillLeading,
		reveal:      1,
	}
}

func (q *queueSlots) Bounds() geometry.Rect {
	return geometry.RectFromCenter(geometry.Pt(q.cx, q.cy), q.w, q.h)
}
func (q *queueSlots) Children() []mobject.Mobject  { return nil }
func (q *queueSlots) Seed() int64                  { return q.seed }
func (q *queueSlots) Style() *style.Style          { return q.Group.Style() }
func (q *queueSlots) SetStyle(s style.Style)       { q.Group.SetStyle(s) }
func (q *queueSlots) Position() (float64, float64) { return q.cx, q.cy }
func (q *queueSlots) SetPosition(x, y float64)     { q.cx, q.cy = x, y }
func (q *queueSlots) SetReveal(t float64)          { q.reveal = t }
func (q *queueSlots) SetVisualScale(float64)       {}

func (q *queueSlots) Render(rd render.Renderer, ctx style.Context) {
	if q.reveal <= 0 {
		return
	}
	eff := ctx.Resolve(*q.Group.Style())
	tok := style.TokensFor(eff)
	stroke := style.PathStyleStroke(eff, tok)
	// Slot dividers are slightly thicker than 0.7× — they need to read
	// through dense Cartoonist hatching, which is otherwise the same
	// color as the divider stroke.
	stroke.StrokeWidth *= 0.95

	// koColor is the color we paint behind each divider line so the
	// hatch pattern doesn't slash through it. Same priority as
	// IconBase's knockoutColor: fill first, then scene bg.
	var koColor color.Color
	switch eff.FillStyle {
	case style.FillHatch, style.FillCrossHatch, style.FillZigzag, style.FillDots:
		if eff.FillColor != nil {
			koColor = eff.FillColor
		} else if ctx.BgColor != nil {
			koColor = ctx.BgColor
		}
	}
	koHalo := 0.0
	if koColor != nil {
		switch {
		case tok.Roughness >= 2:
			koHalo = 5
		case tok.Roughness >= 1:
			koHalo = 3
		}
	}

	bb := q.Bounds()
	if q.horiz {
		// Vertical divider lines for a horizontal queue.
		slotW := q.w / float64(q.slots)
		if q.fillLeading {
			// Filled "next-out" indicator on the leading (leftmost) slot.
			fillPoly := rough.RectToPolygon(bb.Min.X+1, bb.Min.Y+1, slotW-2, q.h-2)
			rd.DrawPath(rough.SolidFill(fillPoly), render.PathStyle{
				Fill: style.ApplyOpacity(eff.StrokeColor, tok.OpacityScale*0.18),
			})
		}
		for i := 1; i < q.slots; i++ {
			x := bb.Min.X + float64(i)*slotW
			p1 := geometry.Pt(x, bb.Min.Y+6)
			p2 := geometry.Pt(x, bb.Max.Y-6)
			if koHalo > 0 {
				ko := geometry.RectanglePath(x-koHalo, p1.Y, 2*koHalo, p2.Y-p1.Y, 0)
				rd.DrawPath(ko, render.PathStyle{
					Fill: style.ApplyOpacity(koColor, tok.OpacityScale*q.reveal),
				})
			}
			rd.DrawPath(makeLine(p1, p2, tok, eff, q.seed+int64(i)), stroke)
		}
	} else {
		// Horizontal divider lines for a vertical stack. Top slot is filled.
		slotH := q.h / float64(q.slots)
		if q.fillLeading {
			fillPoly := rough.RectToPolygon(bb.Min.X+1, bb.Max.Y-slotH+1, q.w-2, slotH-2)
			rd.DrawPath(rough.SolidFill(fillPoly), render.PathStyle{
				Fill: style.ApplyOpacity(eff.StrokeColor, tok.OpacityScale*0.18),
			})
		}
		for i := 1; i < q.slots; i++ {
			y := bb.Max.Y - float64(i)*slotH
			p1 := geometry.Pt(bb.Min.X+6, y)
			p2 := geometry.Pt(bb.Max.X-6, y)
			if koHalo > 0 {
				ko := geometry.RectanglePath(p1.X, y-koHalo, p2.X-p1.X, 2*koHalo, 0)
				rd.DrawPath(ko, render.PathStyle{
					Fill: style.ApplyOpacity(koColor, tok.OpacityScale*q.reveal),
				})
			}
			rd.DrawPath(makeLine(p1, p2, tok, eff, q.seed+int64(i)), stroke)
		}
	}
}

// sidePorts draws three small filled dots on each side of a horizontal
// rectangle, conveying multiple inputs/outputs.
type sidePorts struct {
	*mobject.Group
	seed   int64
	w, h   float64
	cx, cy float64
	reveal float64
}

func newSidePorts(seed int64, w, h float64) *sidePorts {
	return &sidePorts{Group: mobject.NewGroup(seed), seed: seed, w: w, h: h, reveal: 1}
}

func (s *sidePorts) Bounds() geometry.Rect {
	return geometry.RectFromCenter(geometry.Pt(s.cx, s.cy), s.w+24, s.h)
}
func (s *sidePorts) Children() []mobject.Mobject  { return nil }
func (s *sidePorts) Seed() int64                  { return s.seed }
func (s *sidePorts) Style() *style.Style          { return s.Group.Style() }
func (s *sidePorts) SetStyle(st style.Style)      { s.Group.SetStyle(st) }
func (s *sidePorts) Position() (float64, float64) { return s.cx, s.cy }
func (s *sidePorts) SetPosition(x, y float64)     { s.cx, s.cy = x, y }
func (s *sidePorts) SetReveal(t float64)          { s.reveal = t }
func (s *sidePorts) SetVisualScale(float64)       {}

func (s *sidePorts) Render(rd render.Renderer, ctx style.Context) {
	if s.reveal <= 0 {
		return
	}
	eff := ctx.Resolve(*s.Group.Style())
	tok := style.TokensFor(eff)
	col := style.ApplyOpacity(eff.StrokeColor, tok.OpacityScale*s.reveal)
	// Dots: a bit larger and pushed further from the body edge so they
	// don't visually merge with the rectangle outline under sketchy
	// stroke wobble. Pub/sub endpoints want to read as distinct nodes.
	const r = 6.0
	const offset = 14.0
	leftX := s.cx - s.w/2 - offset
	rightX := s.cx + s.w/2 + offset

	// Per-dot knock-out. The dot center sits outside the icon body but
	// at Cartoonist roughness the dot's outline jitter can extend back
	// into the hatched region; halo each dot individually with a
	// circle in the fill / background color so it always reads as a
	// crisp solid against any backdrop.
	var koColor color.Color
	switch eff.FillStyle {
	case style.FillHatch, style.FillCrossHatch, style.FillZigzag, style.FillDots:
		if eff.FillColor != nil {
			koColor = eff.FillColor
		} else if ctx.BgColor != nil {
			koColor = ctx.BgColor
		}
	}
	haloR := r + 2
	if tok.Roughness >= 2 {
		haloR = r + 4
	}

	dots := [][2]float64{}
	for i := -1; i <= 1; i++ {
		y := s.cy + float64(i)*(s.h/3)
		dots = append(dots, [2]float64{leftX, y}, [2]float64{rightX, y})
	}
	if koColor != nil {
		haloCol := style.ApplyOpacity(koColor, tok.OpacityScale*s.reveal)
		for _, d := range dots {
			rd.DrawPath(geometry.EllipsePath(d[0], d[1], haloR, haloR), render.PathStyle{Fill: haloCol})
		}
	}
	for _, d := range dots {
		rd.DrawPath(geometry.EllipsePath(d[0], d[1], r, r), render.PathStyle{Fill: col})
	}
}

// makeLine returns a rough or clean line based on tok.Roughness.
func makeLine(p1, p2 geometry.Point, tok style.Tokens, eff style.Style, seed int64) *geometry.Path {
	if tok.Roughness == 0 {
		return geometry.LinePath(p1, p2)
	}
	opts := style.RoughOptions(eff, tok, seed)
	opts.DisableMultiStroke = true
	opts.Roughness = tok.Roughness * 0.7
	return rough.RoughLine(p1, p2, opts)
}
