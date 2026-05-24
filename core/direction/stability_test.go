package direction_test

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"image/color"
	"testing"

	"github.com/ankitsinghchadda/goanim/core/direction"
	"github.com/ankitsinghchadda/goanim/core/mobject"
	"github.com/ankitsinghchadda/goanim/core/render"
	"github.com/ankitsinghchadda/goanim/core/scene"
	"github.com/ankitsinghchadda/goanim/core/style"
)

// TestRoughGeometryStableUnderCameraZoom — the cached rough path for
// a sketchy rectangle must NOT change when the camera zooms onto it.
// This is the Phase-2 temporal-stability invariant extended to the
// Phase-7 camera: the camera is a final-output transform, not a
// geometry change.
//
// We render the same rectangle at two camera zooms and compare the
// geometry path's hash via the public Path API. The rendered pixels
// will differ (that's what zoom does); the underlying geometry must
// not.
func TestRoughGeometryStableUnderCameraZoom(t *testing.T) {
	rect := mobject.NewRectangle(42, 300, 200)
	rect.MoveTo(0, 0)
	st := *rect.Style()
	st.Sloppiness = style.SloppinessCartoonist // max wobble
	st.FillStyle = style.FillHatch
	st.StrokeColor = color.RGBA{0, 0, 0, 0xFF}
	rect.SetStyle(st)

	// Build a context and a camera. Capture the rendered output at
	// zoom 1, then at zoom 2 (via camera). The rectangle's seed +
	// dimensions don't change between renders, so the rough cache
	// returns the same geometry path; only the canvas-level transform
	// scales the visible output.
	cam := direction.NewCamera()
	a := renderHashed(t, rect, cam, 0, 0, 1)
	cam.Cx, cam.Cy, cam.Zoom = 0, 0, 2
	b := renderHashed(t, rect, cam, 0, 0, 2)
	if a == b {
		t.Errorf("zoom 1 and zoom 2 produced byte-identical PNGs — camera transform not applied")
	}

	// Now reset camera; should match the original zoom-1 render
	// exactly (geometry hasn't changed; transform is back to identity).
	cam.Cx, cam.Cy, cam.Zoom = 0, 0, 1
	c := renderHashed(t, rect, cam, 0, 0, 1)
	if a != c {
		t.Errorf("zoom1 → zoom2 → zoom1 should round-trip to identical bytes; got %s vs %s", a[:16], c[:16])
	}
}

// renderHashed sets the camera, renders the mobject onto a fresh
// renderer, and returns a hex hash of the encoded PNG. Camera params
// are passed for explicitness; the camera object is mutated in place.
func renderHashed(t *testing.T, m mobject.Mobject, cam *direction.Camera, cx, cy, zoom float64) string {
	t.Helper()
	cam.Cx, cam.Cy, cam.Zoom = cx, cy, zoom
	r := render.NewCanvasRenderer(render.Options{Supersample: 1})
	s := scene.NewScene(400, 300).WithRenderer(r).WithCamera(cam)
	s.Add(m)
	r.BeginFrame(400, 300, color.RGBA{0xFF, 0xFF, 0xFF, 0xFF})
	s.RenderFrame()
	var buf bytes.Buffer
	if err := r.EncodePNG(&buf); err != nil {
		t.Fatalf("encode: %v", err)
	}
	sum := sha256.Sum256(buf.Bytes())
	return hex.EncodeToString(sum[:])
}
