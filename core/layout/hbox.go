package layout

import (
	"github.com/ankitsinghchadda/goanim/core/geometry"
	"github.com/ankitsinghchadda/goanim/core/mobject"
	"github.com/ankitsinghchadda/goanim/core/render"
	"github.com/ankitsinghchadda/goanim/core/style"
)

// HBox arranges children in a horizontal row.
//
// Children appear left → right at uniform vertical alignment (default
// VMiddle = aligned centers). Spacing is added between consecutive
// children. Padding adds empty space inside the container's bounds.
//
// HBox is a Mobject — it can be added to a scene, animated as a unit,
// and styled (though styling a container has no visible effect; the
// container is invisible).
type HBox struct {
	*mobject.Group
	cx, cy                 float64
	children               []mobject.Mobject
	spacing                float64
	valign                 VerticalAlign
	padT, padR, padB, padL float64
	dirty                  bool

	// Auto-fit. fitMode controls how children scale when the natural
	// total width exceeds (or under-fills) the target. safeW is the
	// target width the layout fits into; default 0 means "use the
	// canonical canvas width minus safe margins."
	fitMode FitMode
	safeW   float64
	// safeMin / safeMax cap how aggressively Fit can scale children.
	// safeMax (e.g. 1.6) prevents FitContent from blowing up a tiny
	// 2-component layout into giants; safeMin (e.g. 0.5) prevents
	// FitToWidth from making 30 components into unreadable specks.
	safeMin float64
	safeMax float64
}

// NewHBox builds a horizontal row of children. The row is centered at
// the origin until MoveTo or SetPosition repositions it.
func NewHBox(children ...mobject.Mobject) *HBox {
	h := &HBox{
		Group:    mobject.NewGroup(0),
		children: append([]mobject.Mobject{}, children...),
		valign:   VMiddle,
		dirty:    true,
	}
	for _, c := range children {
		h.Group.Add(c)
	}
	return h
}

// WithSpacing sets the gap (in user-space units) between consecutive
// children.
func (h *HBox) WithSpacing(px float64) *HBox { h.spacing = px; h.dirty = true; return h }

// WithAlign sets the vertical alignment of children along the row's
// cross axis.
func (h *HBox) WithAlign(v VerticalAlign) *HBox { h.valign = v; h.dirty = true; return h }

// WithPadding adds empty space inside the container's bounds, in the
// order top, right, bottom, left (CSS convention).
func (h *HBox) WithPadding(top, right, bottom, left float64) *HBox {
	h.padT, h.padR, h.padB, h.padL = top, right, bottom, left
	h.dirty = true
	return h
}

// WithPaddingAll sets all four padding values to v.
func (h *HBox) WithPaddingAll(v float64) *HBox { return h.WithPadding(v, v, v, v) }

// Fit sets the Phase-9 auto-fit mode. Default is FitFixed.
//
//	layout.NewHBox(a, b, c, d, e, f).Fit(layout.FitToWidth)
func (h *HBox) Fit(m FitMode) *HBox { h.fitMode = m; h.dirty = true; return h }

// WithSafeWidth overrides the default safe-area width target that
// Fit modes scale to. Most callers use the default 1728px (1920 minus
// 5% on each side) but a wider/narrower scene can override.
func (h *HBox) WithSafeWidth(w float64) *HBox { h.safeW = w; h.dirty = true; return h }

// WithFitCaps overrides the default 0.5 / 1.6 min/max scale clamps.
// Useful when you want a tighter or looser auto-fit behavior.
func (h *HBox) WithFitCaps(minScale, maxScale float64) *HBox {
	h.safeMin = minScale
	h.safeMax = maxScale
	h.dirty = true
	return h
}

// MoveTo sets the container's center and triggers a relayout.
func (h *HBox) MoveTo(x, y float64) *HBox { h.cx, h.cy = x, y; h.dirty = true; return h }

// SetPosition is the imperative form of MoveTo.
func (h *HBox) SetPosition(x, y float64) { h.cx, h.cy = x, y; h.dirty = true }

// Position returns the container's center.
func (h *HBox) Position() (float64, float64) { return h.cx, h.cy }

// Children returns the laid-out children.
func (h *HBox) Children() []mobject.Mobject { return h.children }

// Layout forces a re-position of children. Normally called lazily by
// Bounds() and Render().
//
// Phase-9 — when fitMode is not FitFixed, the natural total width is
// measured first and a scale factor is computed. The factor is
// clamped by safeMin/safeMax, applied via SetVisualScale to each
// child, and used to scale the child's contribution to positioning.
// Children that don't implement SetVisualScale skip the scale step
// (they keep their natural size in the layout).
func (h *HBox) Layout() {
	h.dirty = false
	if len(h.children) == 0 {
		return
	}

	// Measure pass. When fitMode is FitFixed (the default) we only need
	// to call Bounds() once per child and skip the scale loop entirely
	// — this preserves Phase-1/2/8 layout cost.
	widths := make([]float64, len(h.children))
	heights := make([]float64, len(h.children))
	for i, c := range h.children {
		b := c.Bounds()
		widths[i] = b.Width()
		heights[i] = b.Height()
	}
	scale := 1.0
	if h.fitMode != FitFixed {
		// Auto-fit modes: compute scale and apply via SetVisualScale.
		// The widths we just measured are the natural widths because the
		// children's current VisualScale was 1.0 (we reset it below).
		naturalTotal := h.spacing * float64(len(h.children)-1)
		for _, w := range widths {
			naturalTotal += w
		}
		naturalH := 0.0
		for _, ht := range heights {
			if ht > naturalH {
				naturalH = ht
			}
		}
		scale = h.computeFitScale(naturalTotal, naturalH)
		if scale != 1.0 {
			for i, c := range h.children {
				if s, ok := c.(interface{ SetVisualScale(float64) }); ok {
					s.SetVisualScale(scale)
				}
				widths[i] *= scale
				heights[i] *= scale
			}
		}
	}
	totalW := h.spacing*float64(len(h.children)-1)*scale + sumf(widths)
	maxH := 0.0
	for _, ht := range heights {
		if ht > maxH {
			maxH = ht
		}
	}

	// Lay out: left edge of the row sits at cx - totalW/2.
	leftX := h.cx - totalW/2
	for i, c := range h.children {
		cx := leftX + widths[i]/2
		var cy float64
		switch h.valign {
		case VTop:
			cy = h.cy + maxH/2 - heights[i]/2
		case VBottom:
			cy = h.cy - maxH/2 + heights[i]/2
		default: // VMiddle
			cy = h.cy
		}
		posMove(c, cx, cy)
		leftX += widths[i] + h.spacing*scale
	}
}

// computeFitScale returns the uniform scale factor to apply to each
// child given the natural total width / max height and the configured
// fit mode. Phase-10 Fix 4 — FitContent now considers BOTH width and
// height so a 2-3 component layout grows to fill the canvas instead
// of floating in it. The cap is DefaultFillMaxUp (2.5×) so a tiny
// 2-component layout can scale up dramatically without blowing out.
//
// Per-mode behavior:
//
//   - FitToWidth: scale ≤ 1.0; clamped at [DefaultSafeMin, 1.0].
//     Strictly prevents horizontal overflow.
//   - FitContent: scale ≥ 1.0; clamped at [1.0, DefaultFillMaxUp].
//     Computes max uniform scale fitting BOTH safeW and safeH; never
//     shrinks (that's FitToWidth's job).
//   - FitToCanvas: scale fitted to both axes; clamped at
//     [DefaultSafeMin, DefaultFillMaxUp]. The full both-directions
//     scale — may shrink or grow.
//   - FitFixed: 1.0 (no scaling).
func (h *HBox) computeFitScale(naturalTotal, naturalMaxH float64) float64 {
	if h.fitMode == FitFixed {
		return 1.0
	}
	targetW := h.safeW
	if targetW <= 0 {
		targetW = DefaultSafeWidth
	}
	targetH := DefaultSafeHeight
	minS := h.safeMin
	if minS <= 0 {
		minS = DefaultSafeMin
	}
	maxS := h.safeMax
	if maxS <= 0 {
		// FitContent / FitToCanvas grow to fill; FitToWidth caps lower
		// to avoid blowing up icons just because the row is short.
		if h.fitMode == FitContent || h.fitMode == FitToCanvas {
			maxS = DefaultFillMaxUp
		} else {
			maxS = DefaultSafeMax
		}
	}
	if naturalTotal <= 0 {
		return 1.0
	}
	sx := targetW / naturalTotal
	sy := sx
	if naturalMaxH > 0 {
		sy = targetH / naturalMaxH
	}
	// Max uniform scale that fits both dimensions.
	sFit := sx
	if sy < sFit {
		sFit = sy
	}
	s := sFit
	switch h.fitMode {
	case FitToWidth:
		// Width-only fit: ignore height; never scale up.
		s = sx
		if s > 1.0 {
			s = 1.0
		}
	case FitContent:
		// Aggressive scale-up to fill canvas. Never shrink (FitToWidth's
		// job). Uses both-dim sFit so we don't overflow either axis.
		if sFit < 1.0 {
			s = 1.0
		}
	case FitToCanvas:
		// Both-axes fit, may shrink or grow.
		s = sFit
	}
	if s < minS {
		s = minS
	}
	if s > maxS {
		s = maxS
	}
	return s
}

func sumf(xs []float64) float64 {
	t := 0.0
	for _, x := range xs {
		t += x
	}
	return t
}

// Bounds returns the bounding box of the laid-out container including
// padding. Triggers layout if necessary.
func (h *HBox) Bounds() geometry.Rect {
	if h.dirty {
		h.Layout()
	}
	if len(h.children) == 0 {
		return geometry.RectFromCenter(geometry.Pt(h.cx, h.cy), 0, 0)
	}
	var b geometry.Rect
	first := true
	for _, c := range h.children {
		cb := c.Bounds()
		if cb.Empty() {
			continue
		}
		if first {
			b = cb
			first = false
		} else {
			b = b.Union(cb)
		}
	}
	// Apply padding.
	b.Min.X -= h.padL
	b.Min.Y -= h.padB
	b.Max.X += h.padR
	b.Max.Y += h.padT
	return b
}

// Render lays out (if dirty) and draws each child.
func (h *HBox) Render(r render.Renderer, ctx style.Context) {
	if h.dirty {
		h.Layout()
	}
	for _, c := range h.children {
		c.Render(r, ctx)
	}
}

// Seed returns the container's seed (always 0 — containers are visually
// invisible, so the seed doesn't matter).
func (h *HBox) Seed() int64 { return 0 }
