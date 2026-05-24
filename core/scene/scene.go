package scene

import (
	"image/color"

	"github.com/ankitsinghchadda/goanim/core/mobject"
	"github.com/ankitsinghchadda/goanim/core/render"
	"github.com/ankitsinghchadda/goanim/core/style"
)

// Camera names the subset of viewport / dimming state the scene needs
// to read on every frame. The concrete Camera type lives in
// core/direction so it can also expose animation factories without an
// import cycle. The scene treats this as an opaque "current state of
// the lens."
type Camera interface {
	// Position returns the scene-space point that maps to the center
	// of the output frame. (0, 0) means "scene origin at frame center"
	// — the pre-Phase-7 default.
	Position() (cx, cy float64)
	// ZoomLevel returns the zoom factor (1.0 = identity).
	ZoomLevel() float64
	// FocusTarget returns the currently-focused mobject (nil when no
	// focus is active). Mobjects that aren't the focus target (and
	// aren't descendants of it) render at DimAmount opacity.
	FocusTarget() mobject.Mobject
	// DimAmount returns the opacity multiplier applied to non-focused
	// mobjects. 1.0 means no dim.
	DimAmount() float64
}

// Scene is goanim's top-level canvas: a list of mobjects, a renderer,
// a per-scene style default, an animation timeline, and a Camera
// (viewport / focus state applied by the player before every frame).
type Scene struct {
	Width, Height int
	FPS           int
	BgColor       color.Color
	DefaultStyle  style.Style
	Renderer      render.Renderer
	Fonts         map[style.FontFamily]render.FontFace
	Mobjects      []mobject.Mobject

	// Camera, when set, is consulted every frame: the player calls
	// Renderer.SetCamera with its Position/ZoomLevel and dims any
	// mobjects that aren't the FocusTarget (or a descendant). nil
	// means "no camera" — render with identity viewport, no dim. The
	// scene constructs a default Camera in NewScene; callers can
	// override.
	Camera Camera
}

// NewScene constructs a Scene with sane defaults: 60 FPS, white bg,
// the Excalidraw preset as the scene default style.
func NewScene(width, height int) *Scene {
	return &Scene{
		Width:        width,
		Height:       height,
		FPS:          60,
		BgColor:      color.RGBA{0xFF, 0xFF, 0xFF, 0xFF},
		DefaultStyle: style.PresetExcalidraw,
		Fonts:        map[style.FontFamily]render.FontFace{},
	}
}

// WithDefaultStyle replaces the scene's default style.
func (s *Scene) WithDefaultStyle(st style.Style) *Scene { s.DefaultStyle = st; return s }

// WithRenderer assigns the renderer.
func (s *Scene) WithRenderer(r render.Renderer) *Scene { s.Renderer = r; return s }

// WithFont registers a font face for a logical font family.
func (s *Scene) WithFont(fam style.FontFamily, face render.FontFace) *Scene {
	s.Fonts[fam] = face
	return s
}

// WithCamera attaches a Camera to the scene. The player consults the
// camera every frame to set the renderer viewport and dim non-focused
// mobjects. Most callers build a direction.Camera and pass it here.
func (s *Scene) WithCamera(c Camera) *Scene { s.Camera = c; return s }

// Add appends one or more mobjects.
func (s *Scene) Add(m ...mobject.Mobject) { s.Mobjects = append(s.Mobjects, m...) }

// Remove drops a mobject from the live set.
func (s *Scene) Remove(m mobject.Mobject) {
	for i, x := range s.Mobjects {
		if x == m {
			s.Mobjects = append(s.Mobjects[:i], s.Mobjects[i+1:]...)
			return
		}
	}
}

// Context returns the style.Context carrying scene defaults + fonts.
func (s *Scene) Context() style.Context {
	return style.Context{
		SceneDefault: s.DefaultStyle,
		Fonts:        s.Fonts,
		BgColor:      s.BgColor,
	}
}

// RenderFrame draws every live mobject onto the renderer. Caller is
// responsible for BeginFrame and EncodePNG/WriteFrame around it.
//
// Static renders honor the scene's Camera (if any) — set its
// Cx/Cy/Zoom directly to render a pre-zoomed still image.
func (s *Scene) RenderFrame() {
	if s.Camera != nil {
		cx, cy := s.Camera.Position()
		s.Renderer.SetCamera(cx, cy, s.Camera.ZoomLevel())
	}
	ctx := s.Context()
	for _, m := range s.Mobjects {
		m.Render(s.Renderer, ctx)
	}
}
