package render

import (
	"image"
	"image/color"
	"io"

	"github.com/ankitsinghchadda/goanim/core/geometry"
)

// PathStyle describes how to stroke and/or fill a path.
//
// A zero-alpha Stroke disables stroking; a zero-alpha Fill disables
// filling. StrokeWidth is in user-space units (post-camera transform).
// StrokeCap and StrokeJoin control corner and endpoint geometry.
// DashArray, when non-empty, draws a dashed/dotted stroke (in canvas
// units; pairs of (dash, gap, dash, gap, ...)).
//
// IgnoreCamera marks this path as frame-fixed: the renderer bypasses
// the camera transform and renders the path's user-space coordinates
// directly to the output frame (center-origin still applies). Used by
// the direction layer's Caption — UI chrome that should NOT move when
// the camera pans or scale when it zooms.
type PathStyle struct {
	Stroke       color.Color
	StrokeWidth  float64
	StrokeCap    StrokeCap
	StrokeJoin   StrokeJoin
	Fill         color.Color
	DashArray    []float64
	DashOffset   float64
	IgnoreCamera bool
}

// StrokeCap names the cap style used at the ends of open subpaths.
type StrokeCap uint8

const (
	CapButt StrokeCap = iota
	CapRound
	CapSquare
)

// StrokeJoin names the join style at path corners.
type StrokeJoin uint8

const (
	JoinMiter StrokeJoin = iota
	JoinRound
	JoinBevel
)

// TextStyle describes how to render a run of text.
//
// IgnoreCamera marks this text as frame-fixed (see PathStyle docs).
// Captions use this so their position and size stay constant when
// the camera moves or zooms.
type TextStyle struct {
	Face         FontFace // if nil, the renderer's default face is used
	Size         float64  // user-space units
	Color        color.Color
	Align        TextAlign
	Baseline     TextBaseline
	IgnoreCamera bool
}

// TextAlign controls horizontal alignment relative to the anchor point.
type TextAlign uint8

const (
	AlignLeft TextAlign = iota
	AlignCenter
	AlignRight
)

// TextBaseline controls vertical alignment relative to the anchor point.
type TextBaseline uint8

const (
	BaselineAlphabetic TextBaseline = iota
	BaselineTop
	BaselineMiddle
	BaselineBottom
)

// FontFace is an opaque handle to a loaded font face. Renderers consume
// it; callers obtain one from LoadFont.
type FontFace interface {
	face() // marker
}

// Renderer is the abstract drawing surface. The pipeline is:
//
//	BeginFrame(w, h, bg)
//	  [optional] SetCamera(cx, cy, zoom)
//	  ... DrawPath(...) / DrawText(...) ...
//	EndFrame(...)  // emits the rendered image
//
// SetCamera is the Phase-7 hook for the direction layer: the scene
// calls it with the current viewport before drawing mobjects. The
// renderer applies the camera as a final-output affine transform so
// the camera doesn't disturb cached rough geometry (wobble stays
// stable under camera zoom). Default (no SetCamera call) is identity:
// viewport centered at scene origin, zoom 1.
type Renderer interface {
	// BeginFrame starts a new frame at the given pixel dimensions, clearing
	// to the given background color. Resets the camera to identity.
	BeginFrame(width, height int, bg color.Color)

	// SetCamera sets the current viewport. cx, cy are scene-space
	// coordinates that map to the center of the output frame; zoom
	// multiplies all rendered geometry (1.0 = identity, 2.0 = zoom in
	// 2x). Stroke widths scale with zoom — a 2px stroke at 2x zoom
	// renders as 4px in the output. This matches the cinematic
	// camera-zoom feel and keeps the rough engine's wobble undisturbed.
	SetCamera(cx, cy, zoom float64)

	// DrawPath strokes and/or fills the given path in user-space coordinates.
	DrawPath(p *geometry.Path, style PathStyle)

	// DrawText draws s anchored at (x, y) in user-space coordinates with
	// the given style.
	DrawText(s string, x, y float64, style TextStyle)

	// Image returns the most recently rendered frame.
	Image() image.Image

	// EncodePNG writes the most recently rendered frame as PNG bytes to w.
	EncodePNG(w io.Writer) error
}

// Options configures the raster backend.
type Options struct {
	// Supersample, when > 1, renders internally at Supersample× resolution
	// and downsamples with a high-quality filter. Default (zero) treated
	// as 2.
	Supersample int

	// DefaultFont is the FontFace used for DrawText calls that don't
	// specify a Face. May be nil if no text is drawn.
	DefaultFont FontFace
}
