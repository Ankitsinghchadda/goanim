package direction

import (
	"image/color"
	"time"

	"github.com/ankitsinghchadda/goanim/core/animation"
	"github.com/ankitsinghchadda/goanim/core/animation/easing"
	"github.com/ankitsinghchadda/goanim/core/geometry"
	"github.com/ankitsinghchadda/goanim/core/mobject"
	"github.com/ankitsinghchadda/goanim/core/render"
	"github.com/ankitsinghchadda/goanim/core/style"
)

// Caption shows text at a fixed position near the bottom of the
// frame, independent of the scene's content. The caption fades in,
// holds, and fades out across the total duration (1/6 reveal :
// 4/6 hold : 1/6 fade).
//
// Captions are camera-independent: they don't move when the camera
// pans or zooms. Implementation hooks into the renderer's
// IgnoreCamera flag.
//
//	scene.Play(direction.Caption("This is the request flow", 3*time.Second))
//
// To position a caption at a custom frame location, use
// CaptionAt(text, frameX, frameY, duration). Default y is at
// 38% of frame height below center (i.e., bottom third).
func Caption(text string, duration time.Duration) animation.Animation {
	return CaptionAt(text, 0, -360, duration)
}

// CaptionAt places the caption at explicit frame-space coordinates
// (center-origin: (0,0) = frame center, +y up).
func CaptionAt(text string, frameX, frameY float64, duration time.Duration) animation.Animation {
	c := newCaption(text, frameX, frameY)
	c.SetReveal(0)
	d := duration / 6
	return animation.Sequence(
		&annotationReveal{m: c, from: 0, to: 1, dur: d, ease: easing.OutCubic},
		hold(c, duration-2*d),
		&annotationReveal{m: c, from: 1, to: 0, dur: d, ease: easing.InCubic},
	)
}

// LabelNear places a small text label adjacent to a target mobject,
// tracking the target's position. Unlike Callout, no arrow is drawn;
// the label simply sits in the named position relative to target.
//
// Reveal / hold / fade split is 1/4 : 1/2 : 1/4 like Callout.
//
//	scene.Play(direction.LabelNear(server,
//	    "stateless API",
//	    direction.LabelBelow,
//	    2*time.Second,
//	))
func LabelNear(target mobject.Mobject, text string, pos LabelPosition, totalDuration time.Duration) animation.Animation {
	l := newNearLabel(target, text, pos)
	l.SetReveal(0)
	q := totalDuration / 4
	return animation.Sequence(
		&annotationReveal{m: l, from: 0, to: 1, dur: q, ease: easing.OutCubic},
		hold(l, 2*q),
		&annotationReveal{m: l, from: 1, to: 0, dur: q, ease: easing.InCubic},
	)
}

// LabelPosition mirrors CalloutPosition for LabelNear.
type LabelPosition uint8

const (
	LabelAbove LabelPosition = iota
	LabelBelow
	LabelLeft
	LabelRight
)

// ----- caption mobject -----------------------------------------------------

type captionMobject struct {
	*mobject.Group
	text   string
	frameX float64
	frameY float64
	reveal float64
}

func newCaption(text string, fx, fy float64) *captionMobject {
	return &captionMobject{
		Group:  mobject.NewGroup(0xCAB7104),
		text:   text,
		frameX: fx,
		frameY: fy,
		reveal: 1,
	}
}

func (c *captionMobject) Bounds() geometry.Rect {
	// Caption bounds are frame-space — useful for layout if a caller
	// wants to compose multiple captions. Conservative width: 12px
	// per char.
	w := float64(len(c.text))*12 + 80
	return geometry.RectFromCenter(geometry.Pt(c.frameX, c.frameY), w, 60)
}
func (c *captionMobject) Children() []mobject.Mobject { return nil }
func (c *captionMobject) SetReveal(t float64)         { c.reveal = clamp01(t) }
func (c *captionMobject) Render(rd render.Renderer, ctx style.Context) {
	if c.reveal <= 0 {
		return
	}
	eff := ctx.Resolve(*c.Group.Style())
	face := ctx.FontFace(eff.FontFamily)

	// Pill behind text so the caption reads against any diagram below
	// it. Frame-fixed (IgnoreCamera) so it doesn't move with the
	// camera. Sized to text width + padding.
	w := float64(len(c.text))*14 + 64
	const h = 56
	bg := ctx.BgColor
	if bg == nil {
		bg = color.RGBA{0xFF, 0xFF, 0xFF, 0xFF}
	}
	rd.DrawPath(
		geometry.RectanglePath(c.frameX-w/2, c.frameY-h/2, w, h, 16),
		render.PathStyle{
			Fill:         style.ApplyOpacity(bg, c.reveal*0.93),
			IgnoreCamera: true,
		},
	)
	// Subtle outline.
	rd.DrawPath(
		geometry.RectanglePath(c.frameX-w/2, c.frameY-h/2, w, h, 16),
		render.PathStyle{
			Stroke:       style.ApplyOpacity(eff.StrokeColor, c.reveal*0.35),
			StrokeWidth:  1.2,
			IgnoreCamera: true,
		},
	)
	rd.DrawText(c.text, c.frameX, c.frameY, render.TextStyle{
		Face:         face,
		Size:         style.TokensFor(eff).FontSizePx * 0.75,
		Color:        style.ApplyOpacity(eff.StrokeColor, c.reveal),
		Align:        render.AlignCenter,
		Baseline:     render.BaselineMiddle,
		IgnoreCamera: true,
	})
}

// ----- target-tracking label mobject ---------------------------------------

type nearLabel struct {
	*mobject.Group
	target mobject.Mobject
	text   string
	pos    LabelPosition
	reveal float64
}

func newNearLabel(target mobject.Mobject, text string, pos LabelPosition) *nearLabel {
	return &nearLabel{
		Group:  mobject.NewGroup(0x1ABE1 ^ targetSeed(target)),
		target: target,
		text:   text,
		pos:    pos,
		reveal: 1,
	}
}

func (l *nearLabel) anchor() (cx, cy float64) {
	b := l.target.Bounds()
	const gap = 16
	mx := (b.Min.X + b.Max.X) / 2
	my := (b.Min.Y + b.Max.Y) / 2
	switch l.pos {
	case LabelAbove:
		return mx, b.Max.Y + gap
	case LabelBelow:
		return mx, b.Min.Y - gap
	case LabelLeft:
		return b.Min.X - gap, my
	case LabelRight:
		return b.Max.X + gap, my
	}
	return mx, my
}

func (l *nearLabel) Bounds() geometry.Rect {
	cx, cy := l.anchor()
	w := float64(len(l.text))*11 + 24
	return geometry.RectFromCenter(geometry.Pt(cx, cy), w, 36)
}
func (l *nearLabel) Children() []mobject.Mobject { return nil }
func (l *nearLabel) SetReveal(t float64)         { l.reveal = clamp01(t) }
func (l *nearLabel) Render(rd render.Renderer, ctx style.Context) {
	if l.reveal <= 0 {
		return
	}
	eff := ctx.Resolve(*l.Group.Style())
	face := ctx.FontFace(eff.FontFamily)
	cx, cy := l.anchor()
	switch l.pos {
	case LabelAbove, LabelBelow:
		// nothing — center horizontally
	case LabelLeft:
		cx -= float64(len(l.text)) * 5
	case LabelRight:
		cx += float64(len(l.text)) * 5
	}
	rd.DrawText(l.text, cx, cy, render.TextStyle{
		Face:     face,
		Size:     style.TokensFor(eff).FontSizePx * 0.72,
		Color:    style.ApplyOpacity(eff.StrokeColor, l.reveal),
		Align:    render.AlignCenter,
		Baseline: render.BaselineMiddle,
	})
}
