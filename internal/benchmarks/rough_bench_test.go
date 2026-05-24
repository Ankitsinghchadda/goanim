// Package benchmarks holds component-level micro-benchmarks for
// the Phase-8 performance work. The scene-level benchmarks live in
// cmd/bench; these are the focused ones for isolating individual
// hot paths (rough geometry, layout, camera transform, etc.).
//
// Run with:
//
//	go test ./internal/benchmarks/... -bench=. -benchmem -count=3
package benchmarks

import (
	"image/color"
	"testing"

	"github.com/ankitsinghchadda/goanim/core/geometry"
	"github.com/ankitsinghchadda/goanim/core/rough"
	"github.com/ankitsinghchadda/goanim/core/style"
)

// roughOpts builds the canonical Artist-roughness option set —
// representative of the production Excalidraw preset (1.0 roughness
// + Artist-class jitter + StrokeNormal width).
func roughOpts(seed int64) rough.Options {
	eff := style.PresetExcalidraw
	tok := style.TokensFor(eff)
	return style.RoughOptions(eff, tok, seed)
}

// BenchmarkRoughRectangle_Small — 200×120, no fill. Representative
// of an icon body.
func BenchmarkRoughRectangle_Small(b *testing.B) {
	opts := roughOpts(42)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = rough.RoughRectangle(-100, -60, 200, 120, opts)
	}
}

// BenchmarkRoughRectangle_Large — 800×600.
func BenchmarkRoughRectangle_Large(b *testing.B) {
	opts := roughOpts(42)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = rough.RoughRectangle(-400, -300, 800, 600, opts)
	}
}

// BenchmarkRoughEllipse_Small — 80×60.
func BenchmarkRoughEllipse_Small(b *testing.B) {
	opts := roughOpts(42)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = rough.RoughEllipse(0, 0, 40, 30, opts)
	}
}

// BenchmarkRoughEllipse_Large — 300×200.
func BenchmarkRoughEllipse_Large(b *testing.B) {
	opts := roughOpts(42)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = rough.RoughEllipse(0, 0, 150, 100, opts)
	}
}

// BenchmarkRoughLine — single-stroke line.
func BenchmarkRoughLine(b *testing.B) {
	opts := roughOpts(42)
	p1, p2 := geometry.Pt(-200, 0), geometry.Pt(200, 0)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = rough.RoughLine(p1, p2, opts)
	}
}

// BenchmarkHatchFill_Rectangle — generate a hatch fill for a 200×120
// polygon. The hatch fill is one of the most expensive operations
// for sketchy mode and a strong candidate for caching.
func BenchmarkHatchFill_Rectangle(b *testing.B) {
	opts := roughOpts(42)
	poly := rough.RectToPolygon(-100, -60, 200, 120)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = rough.Hatch(poly, opts)
	}
}

// BenchmarkCrossHatchFill_Rectangle — Cartoonist-mode fill cost.
// Roughly 2× a single Hatch — exercises both diagonals.
func BenchmarkCrossHatchFill_Rectangle(b *testing.B) {
	opts := roughOpts(42)
	poly := rough.RectToPolygon(-100, -60, 200, 120)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = rough.CrossHatch(poly, opts)
	}
}

// BenchmarkSolidFill_Rectangle — the cheapest fill path. Baseline
// against which Hatch/CrossHatch overhead is measured.
func BenchmarkSolidFill_Rectangle(b *testing.B) {
	poly := rough.RectToPolygon(-100, -60, 200, 120)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = rough.SolidFill(poly)
	}
}

// Ensure imports stay used even if the impl evolves.
var _ color.Color = color.RGBA{}
