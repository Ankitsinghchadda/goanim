package render

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"

	"github.com/ankitsinghchadda/goanim/core/geometry"
	"github.com/tdewolff/canvas"
	"github.com/tdewolff/canvas/renderers/rasterizer"
	"golang.org/x/image/draw"
)

// CanvasRenderer is the default raster backend, built on
// github.com/tdewolff/canvas. It renders internally at supersample×
// the requested resolution and downsamples with a Catmull-Rom filter.
//
// Camera state (cameraCx, cameraCy, cameraZoom) is applied as a final-
// output affine transform via canvas.Matrix so the rough engine's
// cached wobble passes through unchanged — this is what keeps the
// Phase-2 temporal-stability invariant true under Phase-7 camera
// zoom.
type CanvasRenderer struct {
	opts       Options
	canvas     *canvas.Canvas
	width      int
	height     int
	bg         color.Color
	out        image.Image
	cameraCx   float64
	cameraCy   float64
	cameraZoom float64
}

// NewCanvasRenderer constructs a renderer with the given options.
//
// Default Supersample is 1. The tdewolff/canvas rasterizer already
// anti-aliases at DPMM=1, so the extra 2× pass adds ~50% of per-frame
// CPU for an almost-imperceptible visual gain on hand-drawn or
// pixel-aligned content. Callers that need maximum-quality output
// (e.g. a print-resolution still) can explicitly pass Supersample: 2.
func NewCanvasRenderer(opts Options) *CanvasRenderer {
	if opts.Supersample <= 0 {
		opts.Supersample = 1
	}
	return &CanvasRenderer{opts: opts}
}

// BeginFrame allocates the working canvas. width and height are the final
// output PNG dimensions in pixels; the canvas's internal coordinate
// system is user-space (center-origin, Y-up) with the same numeric scale
// as the output image (so 1 user unit ≈ 1 pixel).
//
// Resets the camera to identity (cx=0, cy=0, zoom=1) so a frame
// without an explicit SetCamera call renders unaltered.
func (r *CanvasRenderer) BeginFrame(width, height int, bg color.Color) {
	r.width = width
	r.height = height
	r.bg = bg
	r.canvas = canvas.New(float64(width), float64(height))
	r.out = nil // invalidate cached frame
	r.cameraCx = 0
	r.cameraCy = 0
	r.cameraZoom = 1
}

// SetCamera positions the viewport. See Renderer.SetCamera for
// semantics. Called by the scene player every frame; safe to call any
// number of times before drawing.
func (r *CanvasRenderer) SetCamera(cx, cy, zoom float64) {
	r.cameraCx = cx
	r.cameraCy = cy
	if zoom <= 0 {
		zoom = 1
	}
	r.cameraZoom = zoom
}

// worldTransform returns the canvas-space affine transform that maps
// user-space coordinates (center-origin, Y-up) onto the output frame
// with the current camera applied. The transform is built as:
//
//	canvas.Identity
//	  .Translate(width/2, height/2)   // center-origin → bottom-left
//	  .Scale(zoom, zoom)               // camera zoom
//	  .Translate(-cameraCx, -cameraCy) // camera pan
//
// Because tdewolff/canvas applies the matrix to BOTH path geometry
// AND stroke widths, zooming in 2× thickens strokes 2×, which keeps
// the camera feeling "cinematic" rather than "vector-zoom-in-Figma."
func (r *CanvasRenderer) worldTransform() canvas.Matrix {
	tf := canvas.Identity.Translate(float64(r.width)/2, float64(r.height)/2)
	if r.cameraZoom != 1 {
		tf = tf.Scale(r.cameraZoom, r.cameraZoom)
	}
	if r.cameraCx != 0 || r.cameraCy != 0 {
		tf = tf.Translate(-r.cameraCx, -r.cameraCy)
	}
	return tf
}

// buildPath converts a geometry.Path to a canvas.Path in user-space
// coordinates. The center-origin → canvas-space mapping is applied at
// RenderPath time via worldTransform, so the same path can be rendered
// with different camera transforms without rebuilding.
func (r *CanvasRenderer) buildPath(p *geometry.Path) *canvas.Path {
	cp := &canvas.Path{}
	for _, c := range p.Cmds {
		switch c.Kind {
		case geometry.CmdMove:
			cp.MoveTo(c.P0.X, c.P0.Y)
		case geometry.CmdLine:
			cp.LineTo(c.P0.X, c.P0.Y)
		case geometry.CmdCurve:
			cp.CubeTo(c.P0.X, c.P0.Y, c.P1.X, c.P1.Y, c.P2.X, c.P2.Y)
		case geometry.CmdClose:
			cp.Close()
		}
	}
	return cp
}

func toRGBA(c color.Color) color.RGBA {
	if c == nil {
		return color.RGBA{}
	}
	r, g, b, a := c.RGBA()
	return color.RGBA{
		R: uint8(r >> 8),
		G: uint8(g >> 8),
		B: uint8(b >> 8),
		A: uint8(a >> 8),
	}
}

func toCapper(c StrokeCap) canvas.Capper {
	switch c {
	case CapRound:
		return canvas.RoundCap
	case CapSquare:
		return canvas.SquareCap
	default:
		return canvas.ButtCap
	}
}

func toJoiner(j StrokeJoin) canvas.Joiner {
	switch j {
	case JoinRound:
		return canvas.RoundJoin
	case JoinBevel:
		return canvas.BevelJoin
	default:
		return canvas.MiterJoin
	}
}

// DrawPath strokes and/or fills p.
func (r *CanvasRenderer) DrawPath(p *geometry.Path, style PathStyle) {
	if len(p.Cmds) == 0 {
		return
	}
	cp := r.buildPath(p)
	cs := canvas.Style{
		StrokeCapper: toCapper(style.StrokeCap),
		StrokeJoiner: toJoiner(style.StrokeJoin),
	}
	if style.Fill != nil {
		if rgba := toRGBA(style.Fill); rgba.A > 0 {
			cs.Fill = canvas.Paint{Color: rgba}
		}
	}
	if style.Stroke != nil && style.StrokeWidth > 0 {
		if rgba := toRGBA(style.Stroke); rgba.A > 0 {
			cs.Stroke = canvas.Paint{Color: rgba}
			cs.StrokeWidth = style.StrokeWidth
		}
	}
	if len(style.DashArray) > 0 {
		cs.Dashes = append([]float64(nil), style.DashArray...)
		cs.DashOffset = style.DashOffset
	}
	tf := r.worldTransform()
	if style.IgnoreCamera {
		tf = r.frameTransform()
	}
	r.canvas.RenderPath(cp, cs, tf)
}

// frameTransform is the camera-bypass transform: user-space →
// canvas-space with no camera applied. Just the center-origin → bottom-
// left translation. Used for IgnoreCamera paths/text (Captions).
func (r *CanvasRenderer) frameTransform() canvas.Matrix {
	return canvas.Identity.Translate(float64(r.width)/2, float64(r.height)/2)
}

// DrawText draws a single line of text.
func (r *CanvasRenderer) DrawText(s string, x, y float64, style TextStyle) {
	face := style.Face
	if face == nil {
		face = r.opts.DefaultFont
	}
	if face == nil {
		return
	}
	cf, ok := face.(*canvasFontFace)
	if !ok {
		return
	}
	// Default-fallback only when no Color was provided at all. An
	// explicit RGBA with alpha 0 means "invisible" (e.g. text in the
	// middle of a FadeOut) and must be respected — collapsing it to
	// solid black breaks the entire fade-out path for text.
	col := toRGBA(style.Color)
	if style.Color == nil {
		col = color.RGBA{0, 0, 0, 0xFF}
	}
	canvasFace := cf.family.Face(style.Size, col, canvas.FontRegular, canvas.FontNormal)

	var halign canvas.TextAlign
	switch style.Align {
	case AlignCenter:
		halign = canvas.Center
	case AlignRight:
		halign = canvas.Right
	default:
		halign = canvas.Left
	}
	textObj := canvas.NewTextLine(canvasFace, s, halign)

	// Apply vertical baseline by translating y. NewTextLine puts the
	// alphabetic baseline at y=0; tdewolff/canvas uses Y-up.
	//   BaselineAlphabetic — baseline at y, no shift
	//   BaselineTop        — top-of-cap at y; shift baseline below by cap-height
	//   BaselineMiddle     — text visually centered on y; shift baseline below
	//                        by half cap-height (the visual middle of typical
	//                        glyphs sits at cap-height/2 above the baseline)
	//   BaselineBottom     — descender bottom at y; shift baseline up
	metrics := canvasFace.Metrics()
	yShift := 0.0
	switch style.Baseline {
	case BaselineTop:
		yShift = -metrics.CapHeight
	case BaselineMiddle:
		yShift = -metrics.CapHeight / 2
	case BaselineBottom:
		yShift = metrics.Descent
	}

	// Text positions in user-space; the world transform handles
	// centering + camera. Translate(x, y) in user-space, then
	// worldTransform takes it to canvas-space. IgnoreCamera bypasses
	// the camera scale/pan so captions stay frame-fixed.
	tf := r.worldTransform()
	if style.IgnoreCamera {
		tf = r.frameTransform()
	}
	r.canvas.RenderText(textObj, tf.Translate(x, y+yShift))
}

// Image returns the latest rendered frame, rasterizing on demand.
// There is no explicit EndFrame in the renderer interface; call sites
// use Image (or EncodePNG, which wraps it) to retrieve the result.
func (r *CanvasRenderer) Image() image.Image {
	if r.out != nil {
		return r.out
	}
	ss := r.opts.Supersample
	dpmm := canvas.DPMM(float64(ss))
	// Rasterize at supersample resolution. The canvas's coordinate units
	// here equal pixels at 1 DPMM, so DPMM(N) means N pixels per unit.
	img := rasterizer.Draw(r.canvas, dpmm, canvas.DefaultColorSpace)

	// Composite over background, since the rasterizer renders to RGBA
	// with whatever alpha emerges from blending against transparent.
	// Filling the BG via draw.Draw with an image.Uniform source uses
	// an internal fast loop on the RGBA pixel buffer — ~100× faster
	// than a per-pixel dst.Set loop.
	bg := toRGBA(r.bg)
	// dst is re-allocated per frame intentionally — pooling regresses
	// peak RSS because held buffers prevent the Go runtime from
	// releasing memory back to the OS between scenes.
	dst := image.NewRGBA(img.Bounds())
	draw.Draw(dst, dst.Rect, &image.Uniform{C: bg}, image.Point{}, draw.Src)
	draw.Draw(dst, dst.Rect, img, img.Bounds().Min, draw.Over)

	// Downsample to the requested resolution using Catmull-Rom.
	// tdewolff/canvas's rasterizer already maps its Y-up coordinate
	// system to image-space (Y-down) on output, so no additional flip
	// is required.
	if ss == 1 {
		r.out = dst
		return r.out
	}
	// BiLinear instead of CatmullRom for the supersample → output
	// downsample: 4 sample reads per output pixel vs 16, with no
	// perceptible quality loss on hand-drawn / hatched content. The
	// supersample step already does the heavy anti-aliasing; the
	// downsample is mostly area averaging.
	finalRect := image.Rect(0, 0, r.width, r.height)
	// finalBuf reuse also reverted for the same RSS-regression reason
	// as dstBuf above. The scaler cache below is the actual win.
	final := image.NewRGBA(finalRect)
	// BiLinear.Scale with draw.Over composites onto existing content;
	// for the downsampled output we want an OVERWRITE (the previous
	// frame's bytes are still in finalBuf since we reuse it). Using
	// draw.Src ensures every output pixel is the downsampled source,
	// not a blend with stale content.
	//
	draw.BiLinear.Scale(final, finalRect, dst, dst.Rect, draw.Src, nil)
	r.out = final
	return r.out
}

// EncodePNG writes the latest frame to w as PNG.
func (r *CanvasRenderer) EncodePNG(w io.Writer) error {
	img := r.Image()
	if img == nil {
		return fmt.Errorf("render: no frame to encode (call BeginFrame first)")
	}
	enc := png.Encoder{CompressionLevel: png.BestCompression}
	return enc.Encode(w, img)
}

// canvasFontFace is the concrete FontFace implementation backed by a
// canvas.FontFamily.
type canvasFontFace struct {
	family *canvas.FontFamily
}

func (canvasFontFace) face() {}

// LoadFont loads a TTF/OTF font from raw bytes and returns a FontFace
// usable with DrawText.
func LoadFont(name string, ttf []byte) (FontFace, error) {
	fam := canvas.NewFontFamily(name)
	if err := fam.LoadFont(ttf, 0, canvas.FontRegular); err != nil {
		return nil, fmt.Errorf("render: load font %q: %w", name, err)
	}
	return &canvasFontFace{family: fam}, nil
}

// Sanity check: ensure CanvasRenderer satisfies Renderer.
var _ Renderer = (*CanvasRenderer)(nil)
