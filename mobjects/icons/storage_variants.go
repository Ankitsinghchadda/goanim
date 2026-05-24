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

// NewRelationalDB renders the canonical cylinder (same as Database)
// with a small grid-of-rows detail on the cylinder face — signaling
// "this is the relational / tabular variant." Distinguishes it from
// the generic Database by carrying a tabular cue explicitly.
func NewRelationalDB(seed int64, label string) *icon.IconBase {
	const w, h = 220, 200
	ic := icon.New(seed, w, h, label).WithLabelPosition(icon.LabelInside)
	ic.Add(newCylinder(seed, w, h, 28))
	ic.AddDetail(newRelationalGrid(seed+2001, w, h))
	return ic
}

// NewNoSQLDB renders a shorter, squatter cylinder with a small
// document-with-folded-corner detail — "non-tabular storage."
// Distinct from RelationalDB (which has a grid) and Database (which
// has neither detail). Shorter proportions reinforce that this isn't
// the canonical SQL DB.
func NewNoSQLDB(seed int64, label string) *icon.IconBase {
	const w, h = 220, 170
	ic := icon.New(seed, w, h, label).WithLabelPosition(icon.LabelInside)
	ic.Add(newCylinder(seed, w, h, 24))
	ic.AddDetail(newDocumentCorner(seed+2101, w, h))
	return ic
}

// NewKeyValueStore is intentionally NOT a cylinder. It's a compact
// rectangle with a "{ }" detail centered inside — the K:V semantic
// cue. Breaking the cylinder pattern keeps the catalog from drowning
// in six cylinder-shaped icons.
func NewKeyValueStore(seed int64, label string) *icon.IconBase {
	const w, h = 200, 130
	ic := icon.New(seed, w, h, label).WithLabelPosition(icon.LabelBelow)
	ic.Add(mobject.NewRectangle(seed, w, h))
	ic.AddDetail(newBracesDetail(seed+2201, w, h))
	return ic
}

// NewObjectStorage is a stack of small box shapes representing
// individual blob objects in object storage. Distinct from existing
// Bucket (which is three flat ellipses) by being explicit about
// individual objects rather than implying a single container.
func NewObjectStorage(seed int64, label string) *icon.IconBase {
	const w, h = 200, 160
	ic := icon.New(seed, w, h, label).WithLabelPosition(icon.LabelBelow)
	ic.Add(newObjectBoxes(seed+2301, w, h))
	return ic
}

// NewBlockStorage is a rectangle divided into a uniform grid of small
// cells — addressable blocks. Visually distinct from Queue (which
// also has cells but is a single row, not a grid).
func NewBlockStorage(seed int64, label string) *icon.IconBase {
	const w, h = 220, 160
	ic := icon.New(seed, w, h, label).WithLabelPosition(icon.LabelBelow)
	ic.Add(mobject.NewRectangle(seed, w, h))
	ic.Add(newBlockGrid(seed+2401, w, h, 4, 3))
	withLightDensity(ic) // grid IS the metaphor; sparser hatch keeps cells visible
	return ic
}

// --- relationalGrid ---------------------------------------------------------

// relationalGrid draws a small 3-row, 2-col grid centered on the
// cylinder face — the "table" metaphor.
type relationalGrid struct {
	*mobject.Group
	seed   int64
	w, h   float64
	cx, cy float64
	reveal float64
}

func newRelationalGrid(seed int64, w, h float64) *relationalGrid {
	return &relationalGrid{Group: mobject.NewGroup(seed), seed: seed, w: w, h: h, reveal: 1}
}

func (g *relationalGrid) Bounds() geometry.Rect {
	gw := g.w * 0.4
	gh := g.h * 0.35
	return geometry.RectFromCenter(geometry.Pt(g.cx, g.cy-g.h*0.05), gw+8, gh+8)
}
func (g *relationalGrid) Children() []mobject.Mobject  { return nil }
func (g *relationalGrid) Seed() int64                  { return g.seed }
func (g *relationalGrid) Style() *style.Style          { return g.Group.Style() }
func (g *relationalGrid) SetStyle(s style.Style)       { g.Group.SetStyle(s) }
func (g *relationalGrid) Position() (float64, float64) { return g.cx, g.cy }
func (g *relationalGrid) SetPosition(x, y float64)     { g.cx, g.cy = x, y }
func (g *relationalGrid) SetReveal(t float64)          { g.reveal = t }
func (g *relationalGrid) SetVisualScale(float64)       {}

func (g *relationalGrid) Render(rd render.Renderer, ctx style.Context) {
	if g.reveal <= 0 {
		return
	}
	eff := ctx.Resolve(*g.Group.Style())
	tok := style.TokensFor(eff)
	stroke := style.PathStyleStroke(eff, tok)
	stroke.StrokeWidth *= 0.85
	if g.reveal < 1 {
		stroke.Stroke = style.ApplyOpacity(eff.StrokeColor, tok.OpacityScale*g.reveal)
	}
	// Centered slightly above the cylinder midline so it sits on the
	// front face, not in the bottom-cap area.
	gw := g.w * 0.4
	gh := g.h * 0.35
	cx := g.cx
	cy := g.cy - g.h*0.05
	left, right := cx-gw/2, cx+gw/2
	top, bot := cy+gh/2, cy-gh/2
	// Outer rectangle.
	rd.DrawPath(makeLine(geometry.Pt(left, top), geometry.Pt(right, top), tok, eff, g.seed), stroke)
	rd.DrawPath(makeLine(geometry.Pt(right, top), geometry.Pt(right, bot), tok, eff, g.seed+1), stroke)
	rd.DrawPath(makeLine(geometry.Pt(right, bot), geometry.Pt(left, bot), tok, eff, g.seed+2), stroke)
	rd.DrawPath(makeLine(geometry.Pt(left, bot), geometry.Pt(left, top), tok, eff, g.seed+3), stroke)
	// 2 internal horizontal lines → 3 rows.
	for i := 1; i < 3; i++ {
		y := top - float64(i)*gh/3
		rd.DrawPath(makeLine(geometry.Pt(left, y), geometry.Pt(right, y), tok, eff, g.seed+int64(10+i)), stroke)
	}
	// 1 internal vertical → 2 cols.
	rd.DrawPath(makeLine(geometry.Pt(cx, top), geometry.Pt(cx, bot), tok, eff, g.seed+20), stroke)
}

// --- documentCorner --------------------------------------------------------

// documentCorner draws a small rectangle with the top-right corner
// folded down — the universal "document" cue.
type documentCorner struct {
	*mobject.Group
	seed   int64
	w, h   float64
	cx, cy float64
	reveal float64
}

func newDocumentCorner(seed int64, w, h float64) *documentCorner {
	return &documentCorner{Group: mobject.NewGroup(seed), seed: seed, w: w, h: h, reveal: 1}
}

func (d *documentCorner) Bounds() geometry.Rect {
	dw, dh := d.w*0.32, d.h*0.45
	return geometry.RectFromCenter(geometry.Pt(d.cx, d.cy-d.h*0.05), dw+8, dh+8)
}
func (d *documentCorner) Children() []mobject.Mobject  { return nil }
func (d *documentCorner) Seed() int64                  { return d.seed }
func (d *documentCorner) Style() *style.Style          { return d.Group.Style() }
func (d *documentCorner) SetStyle(s style.Style)       { d.Group.SetStyle(s) }
func (d *documentCorner) Position() (float64, float64) { return d.cx, d.cy }
func (d *documentCorner) SetPosition(x, y float64)     { d.cx, d.cy = x, y }
func (d *documentCorner) SetReveal(t float64)          { d.reveal = t }
func (d *documentCorner) SetVisualScale(float64)       {}

func (d *documentCorner) Render(rd render.Renderer, ctx style.Context) {
	if d.reveal <= 0 {
		return
	}
	eff := ctx.Resolve(*d.Group.Style())
	tok := style.TokensFor(eff)
	stroke := style.PathStyleStroke(eff, tok)
	stroke.StrokeWidth *= 0.9
	if d.reveal < 1 {
		stroke.Stroke = style.ApplyOpacity(eff.StrokeColor, tok.OpacityScale*d.reveal)
	}
	dw, dh := d.w*0.32, d.h*0.45
	cx := d.cx
	cy := d.cy - d.h*0.05
	left, right := cx-dw/2, cx+dw/2
	top, bot := cy+dh/2, cy-dh/2
	const fold = 10.0
	// Document outline with folded top-right corner: left → top-left →
	// (top-right minus fold) → diagonal down to top-right indent →
	// right → bottom-right → bottom-left.
	rd.DrawPath(makeLine(geometry.Pt(left, top), geometry.Pt(right-fold, top), tok, eff, d.seed), stroke)
	rd.DrawPath(makeLine(geometry.Pt(right-fold, top), geometry.Pt(right, top-fold), tok, eff, d.seed+1), stroke)
	rd.DrawPath(makeLine(geometry.Pt(right, top-fold), geometry.Pt(right, bot), tok, eff, d.seed+2), stroke)
	rd.DrawPath(makeLine(geometry.Pt(right, bot), geometry.Pt(left, bot), tok, eff, d.seed+3), stroke)
	rd.DrawPath(makeLine(geometry.Pt(left, bot), geometry.Pt(left, top), tok, eff, d.seed+4), stroke)
	// The fold line itself.
	rd.DrawPath(makeLine(geometry.Pt(right-fold, top), geometry.Pt(right-fold, top-fold), tok, eff, d.seed+5), stroke)
	rd.DrawPath(makeLine(geometry.Pt(right-fold, top-fold), geometry.Pt(right, top-fold), tok, eff, d.seed+6), stroke)
}

// --- bracesDetail ----------------------------------------------------------

// bracesDetail draws a large "{ }" pair centered in the icon body —
// the key-value JSON-object cue.
type bracesDetail struct {
	*mobject.Group
	seed   int64
	w, h   float64
	cx, cy float64
	reveal float64
}

func newBracesDetail(seed int64, w, h float64) *bracesDetail {
	return &bracesDetail{Group: mobject.NewGroup(seed), seed: seed, w: w, h: h, reveal: 1}
}

func (b *bracesDetail) Bounds() geometry.Rect {
	return geometry.RectFromCenter(geometry.Pt(b.cx, b.cy), b.w*0.55, b.h*0.65)
}
func (b *bracesDetail) Children() []mobject.Mobject  { return nil }
func (b *bracesDetail) Seed() int64                  { return b.seed }
func (b *bracesDetail) Style() *style.Style          { return b.Group.Style() }
func (b *bracesDetail) SetStyle(s style.Style)       { b.Group.SetStyle(s) }
func (b *bracesDetail) Position() (float64, float64) { return b.cx, b.cy }
func (b *bracesDetail) SetPosition(x, y float64)     { b.cx, b.cy = x, y }
func (b *bracesDetail) SetReveal(t float64)          { b.reveal = t }
func (b *bracesDetail) SetVisualScale(float64)       {}

func (b *bracesDetail) Render(rd render.Renderer, ctx style.Context) {
	if b.reveal <= 0 {
		return
	}
	// Use a Text mobject for "{ }" so it picks up the icon's font /
	// style and renders consistently with other label text.
	t := mobject.NewText(b.seed, "{ }").MoveTo(b.cx, b.cy)
	t.SetStyle(style.Style{FontSize: style.FontXLarge})
	t.SetReveal(b.reveal)
	t.Render(rd, ctx)
}

// --- objectBoxes -----------------------------------------------------------

// objectBoxes draws a 2x2 arrangement of small filled boxes (with one
// offset to suggest "more on top") — the blob-objects cue.
type objectBoxes struct {
	*mobject.Group
	seed   int64
	w, h   float64
	cx, cy float64
	reveal float64
}

func newObjectBoxes(seed int64, w, h float64) *objectBoxes {
	return &objectBoxes{Group: mobject.NewGroup(seed), seed: seed, w: w, h: h, reveal: 1}
}

func (o *objectBoxes) Bounds() geometry.Rect {
	return geometry.RectFromCenter(geometry.Pt(o.cx, o.cy), o.w, o.h)
}
func (o *objectBoxes) Children() []mobject.Mobject  { return nil }
func (o *objectBoxes) Seed() int64                  { return o.seed }
func (o *objectBoxes) Style() *style.Style          { return o.Group.Style() }
func (o *objectBoxes) SetStyle(s style.Style)       { o.Group.SetStyle(s) }
func (o *objectBoxes) Position() (float64, float64) { return o.cx, o.cy }
func (o *objectBoxes) SetPosition(x, y float64)     { o.cx, o.cy = x, y }
func (o *objectBoxes) SetReveal(t float64)          { o.reveal = t }
func (o *objectBoxes) SetVisualScale(float64)       {}

func (o *objectBoxes) Render(rd render.Renderer, ctx style.Context) {
	if o.reveal <= 0 {
		return
	}
	eff := ctx.Resolve(*o.Group.Style())
	tok := style.TokensFor(eff)
	stroke := style.PathStyleStroke(eff, tok)
	if o.reveal < 1 {
		stroke.Stroke = style.ApplyOpacity(eff.StrokeColor, tok.OpacityScale*o.reveal)
	}
	// Four boxes in a 2×2 grid, slightly offset and overlapping like
	// stacked blobs. Each box gets a filled background + outline.
	const bw, bh = 70.0, 55.0
	const gapX, gapY = 14.0, 12.0
	positions := [][2]float64{
		{o.cx - bw/2 - gapX/2, o.cy + bh/2 + gapY/2}, // top-left
		{o.cx + bw/2 + gapX/2, o.cy + bh/2 + gapY/2}, // top-right
		{o.cx - bw/2 - gapX/2, o.cy - bh/2 - gapY/2}, // bottom-left
		{o.cx + bw/2 + gapX/2, o.cy - bh/2 - gapY/2}, // bottom-right
	}
	for i, p := range positions {
		x, y := p[0], p[1]
		// Filled background per box (in the icon's fill color, so it
		// reads against the page bg even for hatch fills).
		if eff.FillColor != nil {
			fillCol := style.ApplyOpacity(eff.FillColor, tok.OpacityScale*o.reveal)
			// In hatch modes we also paint a separate solid color first so
			// each box reads as its own filled tile.
			rd.DrawPath(geometry.RectanglePath(x-bw/2, y-bh/2, bw, bh, tok.CornerRadius/2),
				render.PathStyle{Fill: fillCol})
		}
		// Outline.
		var path *geometry.Path
		if tok.Roughness == 0 {
			path = geometry.RectanglePath(x-bw/2, y-bh/2, bw, bh, tok.CornerRadius/2)
		} else {
			opts := style.RoughOptions(eff, tok, o.seed+int64(i*7))
			opts.Roughness = tok.Roughness * 0.7
			opts.DisableMultiStroke = true
			path = rough.RoughRectangle(x-bw/2, y-bh/2, bw, bh, opts)
		}
		rd.DrawPath(path, stroke)
	}
}

// --- blockGrid -------------------------------------------------------------

// blockGrid draws an N×M grid of cell dividers across the icon body —
// the "addressable blocks" cue. Each cell is drawn as a divider line;
// per-line knock-out keeps them legible against the icon's hatch.
type blockGrid struct {
	*mobject.Group
	seed   int64
	w, h   float64
	cols   int
	rows   int
	cx, cy float64
	reveal float64
}

func newBlockGrid(seed int64, w, h float64, cols, rows int) *blockGrid {
	return &blockGrid{
		Group:  mobject.NewGroup(seed),
		seed:   seed,
		w:      w,
		h:      h,
		cols:   cols,
		rows:   rows,
		reveal: 1,
	}
}

func (g *blockGrid) Bounds() geometry.Rect {
	return geometry.RectFromCenter(geometry.Pt(g.cx, g.cy), g.w, g.h)
}
func (g *blockGrid) Children() []mobject.Mobject  { return nil }
func (g *blockGrid) Seed() int64                  { return g.seed }
func (g *blockGrid) Style() *style.Style          { return g.Group.Style() }
func (g *blockGrid) SetStyle(s style.Style)       { g.Group.SetStyle(s) }
func (g *blockGrid) Position() (float64, float64) { return g.cx, g.cy }
func (g *blockGrid) SetPosition(x, y float64)     { g.cx, g.cy = x, y }
func (g *blockGrid) SetReveal(t float64)          { g.reveal = t }
func (g *blockGrid) SetVisualScale(float64)       {}

func (g *blockGrid) Render(rd render.Renderer, ctx style.Context) {
	if g.reveal <= 0 {
		return
	}
	eff := ctx.Resolve(*g.Group.Style())
	tok := style.TokensFor(eff)
	stroke := style.PathStyleStroke(eff, tok)
	stroke.StrokeWidth *= 0.9
	if g.reveal < 1 {
		stroke.Stroke = style.ApplyOpacity(eff.StrokeColor, tok.OpacityScale*g.reveal)
	}
	// Per-line knock-out color.
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
		if tok.Roughness >= 2 {
			koHalo = 5
		} else if tok.Roughness >= 1 {
			koHalo = 3
		}
	}

	bb := g.Bounds()
	colW := g.w / float64(g.cols)
	rowH := g.h / float64(g.rows)
	// Vertical dividers.
	for i := 1; i < g.cols; i++ {
		x := bb.Min.X + float64(i)*colW
		y1 := bb.Min.Y + 6
		y2 := bb.Max.Y - 6
		if koHalo > 0 {
			ko := geometry.RectanglePath(x-koHalo, y1, 2*koHalo, y2-y1, 0)
			rd.DrawPath(ko, render.PathStyle{Fill: style.ApplyOpacity(koColor, tok.OpacityScale*g.reveal)})
		}
		rd.DrawPath(makeLine(geometry.Pt(x, y1), geometry.Pt(x, y2), tok, eff, g.seed+int64(i)), stroke)
	}
	// Horizontal dividers.
	for j := 1; j < g.rows; j++ {
		y := bb.Min.Y + float64(j)*rowH
		x1 := bb.Min.X + 6
		x2 := bb.Max.X - 6
		if koHalo > 0 {
			ko := geometry.RectanglePath(x1, y-koHalo, x2-x1, 2*koHalo, 0)
			rd.DrawPath(ko, render.PathStyle{Fill: style.ApplyOpacity(koColor, tok.OpacityScale*g.reveal)})
		}
		rd.DrawPath(makeLine(geometry.Pt(x1, y), geometry.Pt(x2, y), tok, eff, g.seed+int64(100+j)), stroke)
	}
}
