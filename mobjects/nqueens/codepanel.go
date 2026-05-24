package nqueens

import (
	"image/color"

	"github.com/ankitsinghchadda/goanim/core/geometry"
	"github.com/ankitsinghchadda/goanim/core/mobject"
	"github.com/ankitsinghchadda/goanim/core/render"
	"github.com/ankitsinghchadda/goanim/core/style"
)

// CodePanel renders a list of high-level "approach" steps directly
// on the scene's background — NO dark backdrop, NO panel outline.
// Header is rendered centered and large; body lines are stacked
// underneath in the scene's stroke color. The currently-highlighted
// line gets a solid amber backdrop strip and a slight forward weight.
//
// The panel is FRAME-FIXED (IgnoreCamera) — it stays put during
// camera zooms / pans on the chessboard.
//
// SetReveal cascades into a coordinated fade. SetHighlightLine picks
// which body line carries the amber strip; the highlight switch is
// applied on the next rendered frame.
type CodePanel struct {
	*mobject.Group
	seed      int64
	cx, cy    float64 // frame-space center
	width     float64
	lineH     float64
	headerPx  float64
	bodyPx    float64
	lines     []string
	highlight int
	reveal    float64
}

// NewCodePanel constructs a panel tuned for plain-English step lists.
// Defaults: width=900, lineH=92, headerPx=78, bodyPx=66 — sized so a
// 6-line list reads comfortably on a 1920×1080 frame next to a
// chessboard.
//
// The first line is treated as the header when it's all uppercase
// letters; it's rendered centered, slightly bigger, in the scene's
// amber accent color, with a hand-drawn underline.
func NewCodePanel(seed int64, lines []string) *CodePanel {
	return &CodePanel{
		Group:     mobject.NewGroup(seed),
		seed:      seed,
		width:     900,
		lineH:     92,
		headerPx:  78,
		bodyPx:    66,
		lines:     append([]string(nil), lines...),
		highlight: -1,
		reveal:    1,
	}
}

// MoveTo sets the panel's frame-space center.
func (c *CodePanel) MoveTo(fx, fy float64) *CodePanel { c.cx, c.cy = fx, fy; return c }

// Position returns the panel's frame-space center.
func (c *CodePanel) Position() (float64, float64) { return c.cx, c.cy }

// SetPosition is the imperative form of MoveTo.
func (c *CodePanel) SetPosition(fx, fy float64) { c.cx, c.cy = fx, fy }

// WithSize tunes the panel footprint. lineH is the per-step vertical
// stride; bodyPx is the body font size; headerPx is the header font
// size (the all-caps first line). Pass 0 to keep a default.
func (c *CodePanel) WithSize(width, lineH, bodyPx, headerPx float64) *CodePanel {
	if width > 0 {
		c.width = width
	}
	if lineH > 0 {
		c.lineH = lineH
	}
	if bodyPx > 0 {
		c.bodyPx = bodyPx
	}
	if headerPx > 0 {
		c.headerPx = headerPx
	}
	return c
}

// SetHighlightLine sets which line (0-indexed in the original lines
// slice — index 0 is the header, so highlights are meaningful for
// index >= 1) gets the amber backdrop on the next render. -1 clears.
func (c *CodePanel) SetHighlightLine(i int) {
	if i < -1 || i >= len(c.lines) {
		i = -1
	}
	c.highlight = i
}

// HighlightLine returns the currently-highlighted index (or -1).
func (c *CodePanel) HighlightLine() int { return c.highlight }

// SetReveal sets the panel's fade-in fraction (0..1).
func (c *CodePanel) SetReveal(t float64) {
	if t < 0 {
		t = 0
	}
	if t > 1 {
		t = 1
	}
	c.reveal = t
}

// Reveal returns the current reveal fraction.
func (c *CodePanel) Reveal() float64 { return c.reveal }

// isHeaderLine reports whether lines[0] should be rendered as a
// header. Heuristic: contains at least one letter, all letters are
// uppercase. (Numbers, dashes, em-dashes pass through.)
func (c *CodePanel) isHeaderLine() bool {
	if len(c.lines) == 0 {
		return false
	}
	hasLetter := false
	for _, ch := range c.lines[0] {
		if ch >= 'a' && ch <= 'z' {
			return false
		}
		if ch >= 'A' && ch <= 'Z' {
			hasLetter = true
		}
	}
	return hasLetter
}

// Bounds returns the panel's frame-space bounding rectangle.
func (c *CodePanel) Bounds() geometry.Rect {
	h := c.lineH * float64(len(c.lines))
	return geometry.RectFromCenter(geometry.Pt(c.cx, c.cy), c.width, h+60)
}

// Render draws the panel directly onto the scene background — no
// rounded rectangle, no outline. Just the header text, optional
// hand-drawn underline, and stacked body steps. Every draw uses
// IgnoreCamera so the panel stays frame-fixed.
func (c *CodePanel) Render(rd render.Renderer, ctx style.Context) {
	if c.reveal <= 0 {
		return
	}
	hasHeader := c.isHeaderLine()
	n := len(c.lines)

	// Resolved scene style (we use it for the body-line color so the
	// panel inherits the scene's stroke palette).
	eff := ctx.Resolve(*c.Group.Style())
	bodyTextColor := eff.StrokeColor
	if bodyTextColor == nil {
		bodyTextColor = color.RGBA{0x18, 0x10, 0x10, 0xFF}
	}
	// Hand-drawn font matches the rest of the video.
	face := ctx.FontFace(style.FontHandDrawn)
	if face == nil {
		face = ctx.FontFace(eff.FontFamily)
	}

	// Vertical layout: compute y for each line, top-down.
	totalH := c.lineH * float64(n)
	yTop := c.cy + totalH/2
	lineY := func(i int) float64 { return yTop - (float64(i)+0.5)*c.lineH }

	// Highlight backdrop behind the active body line. Skip the header.
	if c.highlight >= 0 && c.highlight < n && !(hasHeader && c.highlight == 0) {
		hy := lineY(c.highlight)
		hh := c.lineH * 0.78
		hw := c.width - 24
		// Amber pre-blended over the scene's cream bg so we can fill
		// with full alpha (avoids the tdewolff partial-alpha quirk).
		hl := applyAlpha(color.RGBA{0xF6, 0xC0, 0x4A, 0xFF}, c.reveal)
		rd.DrawPath(
			geometry.RectanglePath(c.cx-hw/2, hy-hh/2, hw, hh, 14),
			render.PathStyle{Fill: hl, IgnoreCamera: true},
		)
		// Hand-drawn dark border under the highlight to make the
		// active step pop on the cream background.
		rd.DrawPath(
			geometry.RectanglePath(c.cx-hw/2, hy-hh/2, hw, hh, 14),
			render.PathStyle{
				Stroke:       applyAlpha(bodyTextColor, c.reveal*0.55),
				StrokeWidth:  2,
				IgnoreCamera: true,
			},
		)
	}

	// Header (line 0) — centered, larger, in the scene's stroke color
	// with a hand-drawn underline.
	startIdx := 0
	if hasHeader {
		hy := lineY(0)
		rd.DrawText(c.lines[0], c.cx, hy, render.TextStyle{
			Face:         face,
			Size:         c.headerPx,
			Color:        applyAlpha(bodyTextColor, c.reveal),
			Align:        render.AlignCenter,
			Baseline:     render.BaselineMiddle,
			IgnoreCamera: true,
		})
		// Underline strip below header — narrower than the panel, gives
		// a "title" feel without re-introducing a backdrop.
		uy := hy - c.headerPx*0.55
		uw := c.width * 0.55
		rd.DrawPath(
			geometry.RectanglePath(c.cx-uw/2, uy-2, uw, 4, 2),
			render.PathStyle{
				Fill:         applyAlpha(bodyTextColor, c.reveal),
				IgnoreCamera: true,
			},
		)
		startIdx = 1
	}

	// Body lines — left-aligned, scene stroke color, hand-drawn font.
	for i := startIdx; i < n; i++ {
		ly := lineY(i)
		col := bodyTextColor
		// On the highlighted line, write in deep brown for max contrast
		// against the amber backdrop (white-on-amber is harder to read
		// at this size).
		if i == c.highlight {
			col = color.RGBA{0x18, 0x10, 0x10, 0xFF}
		}
		rd.DrawText(c.lines[i], c.cx-c.width/2+36, ly, render.TextStyle{
			Face:         face,
			Size:         c.bodyPx,
			Color:        applyAlpha(col, c.reveal),
			Align:        render.AlignLeft,
			Baseline:     render.BaselineMiddle,
			IgnoreCamera: true,
		})
	}
}

// applyAlpha returns c with its alpha multiplied by the given fraction.
// Mirrors style.ApplyOpacity locally so this package doesn't need a
// circular dep on style.
func applyAlpha(c color.Color, alpha float64) color.Color {
	if alpha <= 0 {
		return color.RGBA{0, 0, 0, 0}
	}
	if alpha >= 1 {
		return c
	}
	r, g, b, a := c.RGBA()
	return color.RGBA{
		uint8(r >> 8),
		uint8(g >> 8),
		uint8(b >> 8),
		uint8(float64(a>>8) * alpha),
	}
}
