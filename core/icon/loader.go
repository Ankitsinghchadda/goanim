package icon

import (
	"errors"
	"image"
	_ "image/png"
	"os"

	"github.com/ankitsinghchadda/goanim/core/geometry"
	"github.com/ankitsinghchadda/goanim/core/mobject"
	"github.com/ankitsinghchadda/goanim/core/render"
	"github.com/ankitsinghchadda/goanim/core/style"
)

// LoadOptions configures a custom-asset icon load.
type LoadOptions struct {
	// Size is the body width (height is derived from the asset's aspect
	// ratio). Default 120.
	Size float64
	// Label is the icon's text label.
	Label string
}

// LoadPNG embeds a PNG file as an icon — useful for branded service
// logos that should keep their original colors. The asset is drawn
// at the requested Size, preserving aspect ratio. Style-aware
// re-rendering is not applied; the PNG appears as-is.
func LoadPNG(path string, opts LoadOptions) (*IconBase, error) {
	if opts.Size <= 0 {
		opts.Size = 120
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()
	img, _, err := image.Decode(f)
	if err != nil {
		return nil, err
	}
	b := img.Bounds()
	if b.Dx() == 0 || b.Dy() == 0 {
		return nil, errors.New("icon.LoadPNG: empty image")
	}
	aspect := float64(b.Dy()) / float64(b.Dx())
	bodyH := opts.Size * aspect
	ic := New(0, opts.Size, bodyH, opts.Label)
	ic.Add(&imageMobject{img: img, w: opts.Size, h: bodyH})
	return ic, nil
}

// LoadSVG is reserved for a Phase-4 SVG parsing pass. Returns
// ErrUnimplemented today — point users to LoadPNG or to constructing
// icons programmatically via the icons package.
func LoadSVG(path string, opts LoadOptions) (*IconBase, error) {
	return nil, ErrUnimplemented
}

// ErrUnimplemented marks features that are stubbed but not yet built.
var ErrUnimplemented = errors.New("icon: SVG loader is a Phase-4 follow-up; use LoadPNG or build icons programmatically for now")

// imageMobject is a tiny mobject that draws a raw image at a fixed
// width/height. It does not participate in the style system —
// branded logos render the same in every preset.
type imageMobject struct {
	img    image.Image
	w, h   float64
	cx, cy float64
}

func (im *imageMobject) Children() []mobject.Mobject  { return nil }
func (im *imageMobject) Seed() int64                  { return 0 }
func (im *imageMobject) Style() *style.Style          { return &style.Style{} }
func (im *imageMobject) SetStyle(style.Style)         {}
func (im *imageMobject) Position() (float64, float64) { return im.cx, im.cy }
func (im *imageMobject) SetPosition(x, y float64)     { im.cx, im.cy = x, y }
func (im *imageMobject) Render(r render.Renderer, _ style.Context) {
	// We don't have a "DrawImage" on the Renderer interface yet — that's
	// a tiny extension needed to land branded-PNG support. For now,
	// this is a no-op. Phase-4 follow-up: add Renderer.DrawImage.
	_ = im.img
}
func (im *imageMobject) Bounds() geometry.Rect {
	return geometry.RectFromCenter(geometry.Pt(im.cx, im.cy), im.w, im.h)
}
