package mobject

import (
	"crypto/sha256"
	"encoding/hex"
	"math"
	"testing"

	"github.com/ankitsinghchadda/goanim/core/geometry"
	"github.com/ankitsinghchadda/goanim/core/style"
)

// hashPath returns a stable hash of a path's command sequence, useful
// for comparing two paths for byte-identical geometry.
func hashPath(p *geometry.Path) string {
	h := sha256.New()
	for _, c := range p.Cmds {
		// Write the kind, then each point's x,y as 8-byte float bits.
		h.Write([]byte{byte(c.Kind)})
		writePoint := func(pt geometry.Point) {
			var buf [16]byte
			f64ToBytes(pt.X, buf[0:8])
			f64ToBytes(pt.Y, buf[8:16])
			h.Write(buf[:])
		}
		writePoint(c.P0)
		writePoint(c.P1)
		writePoint(c.P2)
	}
	return hex.EncodeToString(h.Sum(nil))
}

func f64ToBytes(f float64, dst []byte) {
	bits := math.Float64bits(f)
	for i := 0; i < 8; i++ {
		dst[i] = byte(bits >> (i * 8))
	}
}

// stabilityRoughRectKey hashes the cached rough geometry of a Rectangle
// to verify it's unchanged across position updates.
func stabilityRoughRectKey(r *Rectangle, ctx style.Context) string {
	eff := ctx.Resolve(*r.Style())
	tok := style.TokensFor(eff)
	cached := r.cache.roughRect(r.seed, r.w, r.h, eff, tok)
	return hashPath(cached)
}

// TestRoughGeometryStableAcrossTranslation — the cached rough path
// must NOT change when the mobject is repositioned. This is half of
// temporal stability.
func TestRoughGeometryStableAcrossTranslation(t *testing.T) {
	ctx := style.NewContext()
	ctx.SceneDefault = style.PresetExcalidraw

	rect := NewRectangle(42, 300, 200)
	rect.MoveTo(0, 0)

	initial := stabilityRoughRectKey(rect, ctx)

	// Move across 60 "frames" — equivalent to a 1-second translation.
	for i := 0; i < 60; i++ {
		x := float64(i) * 10
		rect.MoveTo(x, 0)
		got := stabilityRoughRectKey(rect, ctx)
		if got != initial {
			t.Fatalf("rough geometry changed at frame %d: %s vs %s", i, got, initial)
		}
	}
}

// TestRoughGeometryInvalidatesOnResize — resizing should rebuild the
// cached rough geometry (the same wobble doesn't apply to a different
// shape).
func TestRoughGeometryInvalidatesOnResize(t *testing.T) {
	ctx := style.NewContext()
	ctx.SceneDefault = style.PresetExcalidraw

	rect := NewRectangle(42, 300, 200)
	before := stabilityRoughRectKey(rect, ctx)
	rect.SetSize(400, 200)
	after := stabilityRoughRectKey(rect, ctx)
	if before == after {
		t.Fatalf("rough geometry didn't update after resize")
	}
}

// TestRoughGeometryStableAcrossStyleChange_OnlyColor — changing a
// style attribute that DOESN'T affect geometry (just color) should
// leave the cached path unchanged.
func TestRoughGeometryStableAcrossColorOnlyChange(t *testing.T) {
	ctx := style.NewContext()
	ctx.SceneDefault = style.PresetExcalidraw

	rect := NewRectangle(42, 300, 200)
	before := stabilityRoughRectKey(rect, ctx)

	// Change only color. Geometry should stay.
	st := *rect.Style()
	st.StrokeColor = nil // change a color-related field; geometry-relevant
	// fields (sloppiness, edges, stroke width) unchanged.
	rect.SetStyle(st)

	after := stabilityRoughRectKey(rect, ctx)
	if before != after {
		t.Fatalf("rough geometry changed on color-only style change: %s vs %s", before, after)
	}
}
