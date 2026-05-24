// Package icon defines goanim's generic icon abstraction: a mobject
// type that draws a recognizable system-design concept (a queue, a
// cache, etc.) by composing style-aware primitives.
//
// An icon is implemented as a Group of primitive shapes drawn
// programmatically — NOT as a static SVG. That way the icon's
// appearance picks up the active style: rough wobble in sketchy mode,
// clean lines in crisp mode, etc.
//
// Icons participate in arrow routing via the Attachable interface (they
// inherit it from Group). The standard convention is that an icon's
// "shape body" is the part arrows attach to, and the label sits below
// (or, when the body is large enough, inside).
package icon

import (
	"image/color"

	"github.com/ankitsinghchadda/goanim/core/geometry"
	"github.com/ankitsinghchadda/goanim/core/mobject"
	"github.com/ankitsinghchadda/goanim/core/render"
	"github.com/ankitsinghchadda/goanim/core/style"
)

// LabelPosition picks where the label sits relative to the icon body.
type LabelPosition uint8

const (
	LabelBelow  LabelPosition = iota // default: centered under the body
	LabelAbove                       // centered over the body
	LabelInside                      // centered inside the body (for icons big enough)
	LabelRight
	LabelLeft
)

// Icon is the contract every generic icon satisfies. Icons embed
// *IconBase which provides the standard implementation.
type Icon interface {
	mobject.Mobject
	mobject.Attachable
	Label() string
	SetLabel(s string)
	BodyBounds() geometry.Rect // arrows attach to this rectangle
}

// IconBase is the shared scaffolding for every generic icon.
//
// Subclasses construct their visual by appending primitive child
// mobjects (Rectangles, Ellipses, Lines, custom paths) via Add. The
// base handles label placement, position tracking, reveal cascading,
// and attachment-point computation.
//
// An icon's "body" is a logical rectangle (BodyW × BodyH around the
// icon's center). The body is conceptual — there isn't necessarily a
// rectangle child of those exact dimensions; rather, BodyW/BodyH is
// the geometric region arrows attach to and labels are placed below.
type IconBase struct {
	*mobject.Group
	cx, cy        float64
	bodyW, bodyH  float64
	label         string
	labelMobj     *mobject.Text
	labelPosition LabelPosition
	labelGap      float64
	reveal        float64
	scale         float64
	parts         []mobject.Mobject // visual children, owned by IconBase
	isDetail      []bool            // parallel to parts: true → gets knock-out
}

// New constructs a base icon with the given seed, body size, and label.
// Subclass constructors call this then Add their visual primitives.
//
// The label is held aside and rendered LAST — after all visual parts —
// so it sits on top of any fill or decorative element. (Label-inside
// icons would otherwise be invisible.)
func New(seed int64, bodyW, bodyH float64, label string) *IconBase {
	ib := &IconBase{
		Group:         mobject.NewGroup(seed),
		bodyW:         bodyW,
		bodyH:         bodyH,
		label:         label,
		labelGap:      14,
		labelPosition: LabelBelow,
		reveal:        1,
		scale:         1,
	}
	if label != "" {
		ib.labelMobj = mobject.NewText(seed+7777, label)
		// Label inherits FontFamily/FontSize from scene — we don't pin
		// them. This is the same fix as the Phase-2 packet bug.
		// We don't add the label to the Group's children directly; the
		// Render method draws it last after every visual part.
	}
	ib.positionLabel()
	return ib
}

// Add appends a "frame" visual child — anything that contributes to
// the icon's main shape (outer rectangle, cylinder body, etc.). Frame
// parts render in pass 1 with their own fill / stroke style.
func (ib *IconBase) Add(m mobject.Mobject) *IconBase {
	ib.Group.Add(m)
	ib.parts = append(ib.parts, m)
	ib.isDetail = append(ib.isDetail, false)
	return ib
}

// AddDetail appends a detail element (lightning bolt, gear, fan-out
// arrows, etc.). Detail elements get knock-out treatment — a clean
// background-colored halo is punched through the fill pattern around
// each detail's bounds before the detail itself is drawn. This keeps
// small visual cues legible against dense hatch / cross-hatch fills.
//
// Detail elements render LAST among visual parts (after frame parts
// and their knock-outs), but BEFORE the label.
//
// As a side effect, the icon's group style is bumped to
// DetailDensityLight (unless the icon's style already specifies a
// density). Icons that carry visual cues want sparser fills so the
// cues read through the pattern.
func (ib *IconBase) AddDetail(m mobject.Mobject) *IconBase {
	ib.Group.Add(m)
	ib.parts = append(ib.parts, m)
	ib.isDetail = append(ib.isDetail, true)
	if ib.Group.Style().DetailDensity == style.DetailDensityUnset {
		s := *ib.Group.Style()
		s.DetailDensity = style.DetailDensityLight
		ib.Group.SetStyle(s)
	}
	return ib
}

// Label returns the current label string.
func (ib *IconBase) Label() string { return ib.label }

// SetLabel updates the label text and repositions.
func (ib *IconBase) SetLabel(s string) {
	ib.label = s
	if ib.labelMobj == nil {
		return
	}
	ib.labelMobj.Content = s
	ib.positionLabel()
}

// WithLabelPosition picks where the label sits relative to the body.
func (ib *IconBase) WithLabelPosition(p LabelPosition) *IconBase {
	ib.labelPosition = p
	ib.positionLabel()
	return ib
}

// MoveTo sets the icon's center and repositions all parts.
func (ib *IconBase) MoveTo(x, y float64) *IconBase {
	ib.SetPosition(x, y)
	return ib
}

// SetPosition is the imperative form of MoveTo. Repositions visual
// parts by the delta and re-anchors the label.
func (ib *IconBase) SetPosition(x, y float64) {
	dx := x - ib.cx
	dy := y - ib.cy
	ib.cx, ib.cy = x, y
	for _, p := range ib.parts {
		movePartBy(p, dx, dy)
	}
	ib.positionLabel()
}

// Position returns the icon's center.
func (ib *IconBase) Position() (float64, float64) { return ib.cx, ib.cy }

// Bounds returns the union of body + label bounds.
func (ib *IconBase) Bounds() geometry.Rect {
	body := ib.BodyBounds()
	if ib.labelMobj == nil {
		return body
	}
	return body.Union(ib.labelMobj.Bounds())
}

// BodyBounds returns just the body region (no label) — what arrows
// attach to. Phase-9: scaled by VisualScale so auto-fit layouts can
// pack scaled icons correctly. Brief Pulse-style scale animations
// will cause attachment points to oscillate; this is intentional
// (an icon that's visually bigger should reach further) and the
// magnitude is small enough not to be visually jarring.
func (ib *IconBase) BodyBounds() geometry.Rect {
	return geometry.RectFromCenter(geometry.Pt(ib.cx, ib.cy), ib.bodyW*ib.scale, ib.bodyH*ib.scale)
}

// VisualBounds returns the icon's visual region — same as BodyBounds
// today, but kept as a separate method so external code (arrow
// attachment, label placement) doesn't accidentally include label
// space the way Bounds() does.
//
// Arrows must use VisualBounds (or AttachmentPoint) to determine the
// connection point — otherwise arrows to label-below icons land in
// the label gap below the visual.
func (ib *IconBase) VisualBounds() geometry.Rect { return ib.BodyBounds() }

// AttachmentPoint returns the attachment point on the icon body.
// Subclasses with curved bodies (Database) override this; the default
// uses the body (NOT the bounding box including label) edge midpoints.
func (ib *IconBase) AttachmentPoint(side mobject.Side) geometry.Point {
	return mobject.AttachToBoundsEdge(ib.BodyBounds(), side)
}

// SetReveal cascades the reveal fraction to parts AND the label.
func (ib *IconBase) SetReveal(t float64) {
	if t < 0 {
		t = 0
	}
	if t > 1 {
		t = 1
	}
	ib.reveal = t
	for _, p := range ib.parts {
		if r, ok := p.(interface{ SetReveal(float64) }); ok {
			r.SetReveal(t)
		}
	}
	if ib.labelMobj != nil {
		ib.labelMobj.SetReveal(t)
	}
}

// Reveal returns the current reveal fraction.
func (ib *IconBase) Reveal() float64 { return ib.reveal }

// SetVisualScale cascades visual scale to scalable parts. Phase-10 —
// after the scale changes, repositions the label so it clears the
// new (scaled) body edge.
func (ib *IconBase) SetVisualScale(s float64) {
	ib.scale = s
	for _, p := range ib.parts {
		if r, ok := p.(interface{ SetVisualScale(float64) }); ok {
			r.SetVisualScale(s)
		}
	}
	ib.positionLabel()
}

// VisualScale returns the current visual scale.
func (ib *IconBase) VisualScale() float64 { return ib.scale }

// Render draws the icon in four passes:
//
//  1. Frame parts (outer rectangle, cylinder body, etc.) — these
//     contribute the icon's fill / hatch pattern.
//  2. Knock-outs for every detail and (when applicable) the label —
//     filled rectangles in the fill / scene background color, drawn
//     with NO stroke, that punch clean areas through the hatch.
//  3. Detail parts — the visual cues that need to read against the
//     fill (lightning bolts, gears, fan-out arrows).
//  4. Label — drawn last so it sits on top of everything.
//
// Passes 2 and 3 are skipped when there's nothing to knock out (the
// icon's effective fill is None or Solid, so no hatch is in the way).
func (ib *IconBase) Render(r render.Renderer, ctx style.Context) {
	if ib.reveal <= 0 {
		return
	}

	// Push the icon's own style into a child context so attributes set
	// on the icon (notably DetailDensity, but also FillStyle overrides)
	// flow down to children — children's own override → icon style →
	// scene default → library default. Groups don't normally introduce
	// a scope, but icons are a logical visual unit and per-icon style
	// tuning is meaningful.
	childCtx := ctx
	childCtx.SceneDefault = ib.Group.Style().Merge(ctx.SceneDefault)

	eff := childCtx.Resolve(style.Style{})

	// Pass 1: frame parts.
	for i, part := range ib.parts {
		if !ib.isDetail[i] {
			part.Render(r, childCtx)
		}
	}

	needsKO := fillNeedsKnockout(eff.FillStyle)
	koColor := knockoutColor(eff, childCtx)
	koMargin := knockoutMargin(eff)

	// Pass 2: knock-outs for details.
	if needsKO && koColor != nil {
		for i, part := range ib.parts {
			if !ib.isDetail[i] {
				continue
			}
			drawKnockoutFor(r, part, koColor, koMargin, ib.reveal)
		}
	}

	// Pass 3: detail parts.
	for i, part := range ib.parts {
		if ib.isDetail[i] {
			part.Render(r, childCtx)
		}
	}

	// Pass 4: label, with a knock-out underneath when it's positioned
	// inside the icon body (otherwise the hatch pattern slashes
	// through the text).
	if ib.labelMobj != nil {
		if ib.labelPosition == LabelInside && needsKO && koColor != nil {
			drawTextKnockout(r, ib.labelMobj, koColor, koMargin, ib.reveal)
		}
		ib.labelMobj.Render(r, childCtx)
	}
}

// fillNeedsKnockout reports whether the icon's fill is one of the
// hatch-style patterns where punching a clean area would actually
// help. For solid/none fills, knock-outs are pointless and skipped.
func fillNeedsKnockout(fs style.FillStyle) bool {
	switch fs {
	case style.FillHatch, style.FillCrossHatch, style.FillZigzag, style.FillDots:
		return true
	}
	return false
}

// knockoutColor picks the color used to clear a halo around a detail.
// Preference order:
//
//  1. The icon's resolved FillColor (the color UNDER the hatch — the
//     hatch is drawn on top of this, so re-painting with this color
//     erases the hatch and leaves the underlying fill).
//  2. The scene background, when FillColor is nil (e.g. blueprint).
//  3. Nil → caller should skip the knock-out entirely.
func knockoutColor(eff style.Style, ctx style.Context) color.Color {
	if eff.FillColor != nil {
		return eff.FillColor
	}
	return ctx.BgColor
}

// knockoutMargin returns the halo expansion (px) around each detail.
// Larger for Cartoonist (denser hatch needs a wider halo to read).
func knockoutMargin(eff style.Style) float64 {
	tok := style.TokensFor(eff)
	switch {
	case tok.Roughness >= 2:
		return 8
	case tok.Roughness >= 1:
		return 4
	}
	return 0
}

// drawKnockoutFor punches a filled rectangle in koColor around the
// part's bounds (expanded by margin). No stroke — purely a fill that
// erases hatch lines beneath it.
//
// Opacity is cubed against reveal so the knock-out fades in at the
// same rate as the detail it precedes — otherwise a fully-opaque
// cream rectangle would appear over a half-drawn icon during DrawOn.
func drawKnockoutFor(r render.Renderer, part mobject.Mobject, koColor color.Color, margin, reveal float64) {
	if reveal <= 0 {
		return
	}
	bb := part.Bounds()
	if bb.Empty() {
		return
	}
	w := bb.Max.X - bb.Min.X + 2*margin
	h := bb.Max.Y - bb.Min.Y + 2*margin
	if w <= 0 || h <= 0 {
		return
	}
	path := geometry.RectanglePath(bb.Min.X-margin, bb.Min.Y-margin, w, h, 0)
	op := reveal * reveal * reveal
	r.DrawPath(path, render.PathStyle{Fill: style.ApplyOpacity(koColor, op)})
}

// drawTextKnockout punches a filled rectangle around the text's
// approximate bounds. We size from the text mobject's own Bounds()
// (which uses its font tokens to estimate glyph extents) plus the
// supplied margin.
func drawTextKnockout(r render.Renderer, txt *mobject.Text, koColor color.Color, margin, reveal float64) {
	if reveal <= 0 {
		return
	}
	bb := txt.Bounds()
	if bb.Empty() {
		return
	}
	// Labels need a chunkier halo than details so the text never feels
	// jammed up against the hatch on the sides.
	m := margin + 4
	w := bb.Max.X - bb.Min.X + 2*m
	h := bb.Max.Y - bb.Min.Y + 2*m
	path := geometry.RectanglePath(bb.Min.X-m, bb.Min.Y-m, w, h, 0)
	r.DrawPath(path, render.PathStyle{Fill: style.ApplyOpacity(koColor, reveal*reveal*reveal)})
}

// positionLabel places the label according to the label position.
// Phase-10 polish — label edges are measured against the SCALED body
// dimensions (bodyW*scale, bodyH*scale) so that when an icon is
// auto-fit scaled up by FitContent, the label still clears the
// rendered body edge. The pre-fix version used the natural-size
// bodyW/bodyH, which produced labels that slashed through the
// rendered rectangle at scale > 1.
func (ib *IconBase) positionLabel() {
	if ib.labelMobj == nil {
		return
	}
	scale := ib.scale
	if scale <= 0 {
		scale = 1
	}
	halfW := ib.bodyW / 2 * scale
	halfH := ib.bodyH / 2 * scale
	switch ib.labelPosition {
	case LabelAbove:
		ib.labelMobj.MoveTo(ib.cx, ib.cy+halfH+ib.labelGap)
	case LabelInside:
		ib.labelMobj.MoveTo(ib.cx, ib.cy)
	case LabelRight:
		ib.labelMobj.MoveTo(ib.cx+halfW+ib.labelGap, ib.cy)
	case LabelLeft:
		ib.labelMobj.MoveTo(ib.cx-halfW-ib.labelGap, ib.cy)
	default: // LabelBelow
		ib.labelMobj.MoveTo(ib.cx, ib.cy-halfH-ib.labelGap)
	}
}

// movePartBy translates a part by (dx, dy) using whichever method it
// exposes. This is used by IconBase to shift every visual element when
// the icon is repositioned.
func movePartBy(m mobject.Mobject, dx, dy float64) {
	type positioner interface {
		Position() (float64, float64)
		SetPosition(x, y float64)
	}
	if p, ok := m.(positioner); ok {
		x, y := p.Position()
		p.SetPosition(x+dx, y+dy)
		return
	}
	// Some primitives expose only From/To (Line) — translate both.
	type liner interface {
		From() geometry.Point
		To() geometry.Point
		SetFrom(geometry.Point)
		SetTo(geometry.Point)
	}
	if l, ok := m.(liner); ok {
		f := l.From()
		t := l.To()
		l.SetFrom(geometry.Pt(f.X+dx, f.Y+dy))
		l.SetTo(geometry.Pt(t.X+dx, t.Y+dy))
		return
	}
}
