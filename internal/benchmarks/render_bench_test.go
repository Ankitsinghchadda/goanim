package benchmarks

import (
	"image"
	"image/color"
	"testing"

	"github.com/ankitsinghchadda/goanim/core/direction"
	"github.com/ankitsinghchadda/goanim/core/layout"
	"github.com/ankitsinghchadda/goanim/core/render"
	"github.com/ankitsinghchadda/goanim/core/scene"
	"github.com/ankitsinghchadda/goanim/core/style"
	"github.com/ankitsinghchadda/goanim/mobjects/icons"
)

// build10IconScene assembles a 10-icon scene representative of the
// medium-bench workload. Used by the single-frame and pause-frame
// benchmarks to isolate per-frame rendering cost from animation
// orchestration.
func build10IconScene(b *testing.B, preset style.Style) (*scene.Scene, render.Renderer) {
	b.Helper()
	hand, err := render.Excalifont()
	if err != nil {
		b.Fatalf("excalifont: %v", err)
	}
	sans, err := render.Inter()
	if err != nil {
		b.Fatalf("inter: %v", err)
	}
	r := render.NewCanvasRenderer(render.Options{Supersample: 1, DefaultFont: hand})
	s := scene.NewScene(1920, 1080).
		WithRenderer(r).
		WithDefaultStyle(preset).
		WithFont(style.FontHandDrawn, hand).
		WithFont(style.FontSans, sans).
		WithCamera(direction.NewCamera())
	s.BgColor = color.RGBA{0xFF, 0xFF, 0xFF, 0xFF}

	c1 := icons.NewClient(1, "Client")
	gw := icons.NewAPIGateway(2, "Gateway")
	lb := icons.NewLoadBalancer(3, "LB")
	top := layout.NewHBox(c1, gw, lb).WithSpacing(80).MoveTo(0, 200)

	api1 := icons.NewServer(11, "api-1")
	api2 := icons.NewServer(12, "api-2")
	cache := icons.NewCache(13, "cache")
	db := icons.NewDatabase(14, "db")
	q := icons.NewQueue(15, "q")
	w := icons.NewWorker(16, "worker")
	cdn := icons.NewCDN(17, "cdn")
	bot := layout.NewHBox(api1, api2, cache, db, q, w, cdn).WithSpacing(30).MoveTo(0, -80)

	s.Add(top, bot)
	_ = top.Bounds()
	_ = bot.Bounds()
	return s, r
}

// renderOnce is the inner loop used by the per-frame benchmarks. It
// matches what scene.Play does for a single tick.
func renderOnce(s *scene.Scene, r render.Renderer, sink image.Image) {
	r.BeginFrame(1920, 1080, color.RGBA{0xFF, 0xFF, 0xFF, 0xFF})
	s.RenderFrame()
	_ = r.Image()
}

// BenchmarkSingleFrameRender_Crisp — one frame, crisp style, 10
// icons. Architectural floor for crisp throughput: how fast can the
// library render a single typical frame?
func BenchmarkSingleFrameRender_Crisp(b *testing.B) {
	s, r := build10IconScene(b, style.PresetCrisp)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		renderOnce(s, r, nil)
	}
}

// BenchmarkSingleFrameRender_Sketchy — same scene, sketchy style.
// Comparison against the crisp version isolates the cost of
// hand-drawn aesthetics (rough geometry + hatch fill).
func BenchmarkSingleFrameRender_Sketchy(b *testing.B) {
	s, r := build10IconScene(b, style.PresetSketchy)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		renderOnce(s, r, nil)
	}
}

// BenchmarkPauseFrame measures the "scene state hasn't changed"
// rendering path. Today this is identical to a full render — the
// Phase-7 prompt called out a possible frame-byte-caching optimization
// to make pause frames near-free. If that optimization lands, this
// benchmark drops to near-zero alloc and microsecond-class wall time.
func BenchmarkPauseFrame(b *testing.B) {
	s, r := build10IconScene(b, style.PresetSketchy)
	// Pre-render once so any first-frame setup is excluded.
	renderOnce(s, r, nil)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		renderOnce(s, r, nil)
	}
}
