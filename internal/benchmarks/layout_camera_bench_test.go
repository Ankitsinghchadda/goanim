package benchmarks

import (
	"image/color"
	"testing"

	"github.com/ankitsinghchadda/goanim/core/direction"
	"github.com/ankitsinghchadda/goanim/core/layout"
	"github.com/ankitsinghchadda/goanim/core/mobject"
	"github.com/ankitsinghchadda/goanim/core/render"
	"github.com/ankitsinghchadda/goanim/core/scene"
	"github.com/ankitsinghchadda/goanim/core/style"
	"github.com/ankitsinghchadda/goanim/mobjects/icons"
)

// BenchmarkLayoutCompute_HBox20 — time to compute positions for a
// 20-child HBox. Captures the cost of HBox.Bounds() (which triggers
// layout) — if a scene calls Bounds() per frame, this matters.
func BenchmarkLayoutCompute_HBox20(b *testing.B) {
	children := make([]mobject.Mobject, 0, 20)
	for i := 0; i < 20; i++ {
		children = append(children, icons.NewServer(int64(i), "S"))
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		box := layout.NewHBox(children...).WithSpacing(40).MoveTo(0, 0)
		_ = box.Bounds()
	}
}

// BenchmarkLayoutCompute_VBox20 — analogous, vertical.
func BenchmarkLayoutCompute_VBox20(b *testing.B) {
	children := make([]mobject.Mobject, 0, 20)
	for i := 0; i < 20; i++ {
		children = append(children, icons.NewServer(int64(i), "S"))
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		box := layout.NewVBox(children...).WithSpacing(40).MoveTo(0, 0)
		_ = box.Bounds()
	}
}

// BenchmarkCameraTransform — apply a camera transform when rendering
// a single 10-icon frame. The Phase-7 camera transform is applied
// per-DrawPath via canvas.Matrix multiplication; this benchmark
// reveals whether the matrix construction or the per-path application
// is the hot path.
func BenchmarkCameraTransform(b *testing.B) {
	hand, err := render.Excalifont()
	if err != nil {
		b.Fatalf("excalifont: %v", err)
	}
	sans, _ := render.Inter()
	r := render.NewCanvasRenderer(render.Options{Supersample: 1, DefaultFont: hand})
	cam := direction.NewCamera()
	s := scene.NewScene(1920, 1080).
		WithRenderer(r).
		WithDefaultStyle(style.PresetCrisp).
		WithFont(style.FontHandDrawn, hand).
		WithFont(style.FontSans, sans).
		WithCamera(cam)
	s.BgColor = color.RGBA{0xFF, 0xFF, 0xFF, 0xFF}
	c1 := icons.NewClient(1, "Client")
	srv := icons.NewServer(2, "Server")
	db := icons.NewDatabase(3, "Database")
	row := layout.NewHBox(c1, srv, db).WithSpacing(80).MoveTo(0, 0)
	s.Add(row)
	_ = row.Bounds()
	// Zoom 1.5× centered on server.
	cam.Cx, cam.Cy, cam.Zoom = 0, 0, 1.5
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.BeginFrame(1920, 1080, color.RGBA{0xFF, 0xFF, 0xFF, 0xFF})
		s.RenderFrame()
		_ = r.Image()
	}
}
