package icons

import (
	"github.com/ankitsinghchadda/goanim/core/geometry"
	"github.com/ankitsinghchadda/goanim/core/icon"
	"github.com/ankitsinghchadda/goanim/core/mobject"
	"github.com/ankitsinghchadda/goanim/core/render"
	"github.com/ankitsinghchadda/goanim/core/rough"
	"github.com/ankitsinghchadda/goanim/core/style"
)

// NewClient is a labeled rectangular icon representing an end-user
// device. Distinguished from User (a person) by being a "thing."
func NewClient(seed int64, label string) *icon.IconBase {
	const w, h = 220, 140
	ic := icon.New(seed, w, h, label).WithLabelPosition(icon.LabelInside)
	rect := mobject.NewRectangle(seed, w, h)
	ic.Add(rect)
	return ic
}

// NewServer is a labeled rectangle with three small "rack lines" in
// the top-right corner.
func NewServer(seed int64, label string) *icon.IconBase {
	const w, h = 240, 160
	ic := icon.New(seed, w, h, label).WithLabelPosition(icon.LabelInside)
	ic.Add(mobject.NewRectangle(seed, w, h))
	ic.AddDetail(newRackLines(seed+101, w, h))
	return ic
}

// NewService is the generic rectangle with a small dot indicator —
// for the "I don't know what to call this thing" fallback.
func NewService(seed int64, label string) *icon.IconBase {
	const w, h = 220, 140
	ic := icon.New(seed, w, h, label).WithLabelPosition(icon.LabelInside)
	ic.Add(mobject.NewRectangle(seed, w, h))
	// Tiny solid dot in the corner — override the inherited hatch fill
	// so the dot stays a crisp filled circle.
	dot := mobject.NewEllipse(seed+201, 6, 6).MoveTo(w/2-22, h/2-22)
	dot.SetStyle(style.Style{FillStyle: style.FillSolid})
	ic.AddDetail(dot)
	return ic
}

// rackLines is a small mobject drawing three short horizontal lines
// in the top-right area of an icon body — the "rack" cue. We define
// it as its own composite so the icon's part list stays tidy.
type rackLines struct {
	*mobject.Group
	seed   int64
	bodyW  float64
	bodyH  float64
	cx, cy float64
	scale  float64
	reveal float64
}

func newRackLines(seed int64, bodyW, bodyH float64) *rackLines {
	return &rackLines{
		Group:  mobject.NewGroup(seed),
		seed:   seed,
		bodyW:  bodyW,
		bodyH:  bodyH,
		scale:  1,
		reveal: 1,
	}
}

// Bounds returns a tight rectangle around the three rack lines in the
// top-right corner. Used by IconBase's knock-out pass to halo this
// detail against a hatched body.
func (r *rackLines) Bounds() geometry.Rect {
	body := geometry.RectFromCenter(geometry.Pt(r.cx, r.cy), r.bodyW, r.bodyH)
	rightX := body.Max.X - 22
	topY := body.Max.Y - 24
	// 3 lines: lengths 44px, vertical span ~24px starting at topY going down.
	return geometry.Rect{
		Min: geometry.Pt(rightX-44, topY-2*12),
		Max: geometry.Pt(rightX, topY),
	}
}
func (r *rackLines) Children() []mobject.Mobject  { return nil }
func (r *rackLines) Seed() int64                  { return r.seed }
func (r *rackLines) Style() *style.Style          { return r.Group.Style() }
func (r *rackLines) SetStyle(s style.Style)       { r.Group.SetStyle(s) }
func (r *rackLines) Position() (float64, float64) { return r.cx, r.cy }
func (r *rackLines) SetPosition(x, y float64)     { r.cx, r.cy = x, y }
func (r *rackLines) SetReveal(t float64)          { r.reveal = t }
func (r *rackLines) SetVisualScale(s float64)     { r.scale = s }

func (r *rackLines) Render(rd render.Renderer, ctx style.Context) {
	if r.reveal <= 0 {
		return
	}
	eff := ctx.Resolve(*r.Group.Style())
	tok := style.TokensFor(eff)
	bb := r.Bounds()
	rightX := bb.Max.X - 22
	topY := bb.Max.Y - 24
	for i := 0; i < 3; i++ {
		y := topY - float64(i)*12
		p1 := geometry.Pt(rightX-44, y)
		p2 := geometry.Pt(rightX, y)
		ps := style.PathStyleStroke(eff, tok)
		ps.StrokeWidth = ps.StrokeWidth * 0.7
		var path *geometry.Path
		if tok.Roughness == 0 {
			path = geometry.LinePath(p1, p2)
		} else {
			opts := style.RoughOptions(eff, tok, r.seed+int64(1001+i*17))
			opts.Roughness = tok.Roughness * 0.6
			opts.DisableMultiStroke = true
			path = rough.RoughLine(p1, p2, opts)
		}
		if r.reveal < 1 {
			ps.Stroke = style.ApplyOpacity(eff.StrokeColor, tok.OpacityScale*r.reveal)
		}
		rd.DrawPath(path, ps)
	}
}
