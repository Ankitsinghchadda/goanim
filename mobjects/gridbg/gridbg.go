// Package gridbg provides a dark-theme grid background mobject — a
// subtle grid of thin gray lines on a black canvas. Mimics the
// dark-mode Excalidraw aesthetic used in the architecture diagrams
// the Gmail video draws on.
//
// The mobject is FRAME-FIXED (IgnoreCamera): the grid stays put when
// the camera pans or zooms, so it always reads as the canvas
// background rather than part of the diagram.
//
// Add it as the FIRST mobject in your scene so it renders behind
// everything else.
package gridbg

import (
	"image/color"

	"github.com/ankitsinghchadda/goanim/core/geometry"
	"github.com/ankitsinghchadda/goanim/core/mobject"
	"github.com/ankitsinghchadda/goanim/core/render"
	"github.com/ankitsinghchadda/goanim/core/style"
)

// Default grid colors — tuned for a black background.
var (
	defaultMinorColor = color.RGBA{0x1E, 0x21, 0x26, 0xFF} // very dim gray
	defaultMajorColor = color.RGBA{0x2A, 0x2E, 0x35, 0xFF} // slightly brighter for every 4th line
)

// GridBackground draws a regular grid pattern across the entire frame.
// Spacing is in scene-space units; the renderer is canvas-fixed at
// 1920×1080 by default, so spacing=80 gives a 24×14 grid.
type GridBackground struct {
	*mobject.Group
	seed        int64
	width       float64
	height      float64
	spacing     float64
	majorEvery  int // every Nth line is major; 0 = no majors
	minorColor  color.Color
	majorColor  color.Color
	strokeWidth float64
}

// New returns a GridBackground sized to the given (width, height)
// (typically the scene W and H). Default spacing 80px; major lines
// every 4 minors.
func New(seed int64, width, height float64) *GridBackground {
	return &GridBackground{
		Group:       mobject.NewGroup(seed),
		seed:        seed,
		width:       width,
		height:      height,
		spacing:     80,
		majorEvery:  4,
		minorColor:  defaultMinorColor,
		majorColor:  defaultMajorColor,
		strokeWidth: 1,
	}
}

// WithSpacing tunes the grid spacing in scene-space units.
func (g *GridBackground) WithSpacing(spacing float64) *GridBackground {
	if spacing > 0 {
		g.spacing = spacing
	}
	return g
}

// WithColors sets minor and major grid line colors.
func (g *GridBackground) WithColors(minor, major color.Color) *GridBackground {
	if minor != nil {
		g.minorColor = minor
	}
	if major != nil {
		g.majorColor = major
	}
	return g
}

// Bounds returns the frame-space bounds (width × height centered at
// origin).
func (g *GridBackground) Bounds() geometry.Rect {
	return geometry.RectFromCenter(geometry.Pt(0, 0), g.width, g.height)
}

// Render draws every vertical and horizontal grid line. Frame-fixed
// (IgnoreCamera).
func (g *GridBackground) Render(rd render.Renderer, _ style.Context) {
	halfW := g.width / 2
	halfH := g.height / 2

	// Vertical lines.
	idx := 0
	for x := -halfW; x <= halfW+0.001; x += g.spacing {
		col := g.minorColor
		w := g.strokeWidth
		if g.majorEvery > 0 && idx%g.majorEvery == 0 {
			col = g.majorColor
			w = g.strokeWidth + 0.4
		}
		rd.DrawPath(
			geometry.LinePath(geometry.Pt(x, -halfH), geometry.Pt(x, halfH)),
			render.PathStyle{
				Stroke:       col,
				StrokeWidth:  w,
				IgnoreCamera: true,
			},
		)
		idx++
	}
	// Horizontal lines.
	idx = 0
	for y := -halfH; y <= halfH+0.001; y += g.spacing {
		col := g.minorColor
		w := g.strokeWidth
		if g.majorEvery > 0 && idx%g.majorEvery == 0 {
			col = g.majorColor
			w = g.strokeWidth + 0.4
		}
		rd.DrawPath(
			geometry.LinePath(geometry.Pt(-halfW, y), geometry.Pt(halfW, y)),
			render.PathStyle{
				Stroke:       col,
				StrokeWidth:  w,
				IgnoreCamera: true,
			},
		)
		idx++
	}
}
