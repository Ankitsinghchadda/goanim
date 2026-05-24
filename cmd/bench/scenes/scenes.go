// Package scenes is the Phase-8 benchmark scene catalog: three
// representative goanim workloads (small / medium / large) that the
// bench runner exercises in both crisp and sketchy styles. Each
// scene returns its assembled *scene.Scene plus a function that
// drives the animation timeline. Splitting construction from
// execution lets the runner exclude scene-build cost from the timed
// region if it wants to.
//
// The scenes are intentionally NOT the same as cmd/example/* — those
// are user-facing demos with arbitrary tuning. Bench scenes are
// stable, reproducible, and designed to exercise the primitives
// that matter for performance: large numbers of mobjects, hatched
// fills (the most expensive render path), pauses, camera moves,
// and the direction-layer primitives.
package scenes

import (
	"image/color"
	"time"

	"github.com/ankitsinghchadda/goanim/core/animation"
	"github.com/ankitsinghchadda/goanim/core/direction"
	"github.com/ankitsinghchadda/goanim/core/layout"
	"github.com/ankitsinghchadda/goanim/core/mobject"
	"github.com/ankitsinghchadda/goanim/core/render"
	"github.com/ankitsinghchadda/goanim/core/scene"
	"github.com/ankitsinghchadda/goanim/core/style"
	"github.com/ankitsinghchadda/goanim/mobjects/icons"
	"github.com/ankitsinghchadda/goanim/mobjects/systemdesign"
)

// Style names the visual preset a benchmark scene runs in. Mirrors
// the cmd/example/* "crisp" vs "sketchy" distinction. Kept here as
// a string to avoid importing style across the public bench
// reporting boundary.
type Style string

const (
	StyleCrisp   Style = "crisp"
	StyleSketchy Style = "sketchy"
)

// SceneSpec describes a benchmark workload — what to render and at
// what visual fidelity. The Build closure returns a ready-to-play
// scene and the animation that drives the full timeline.
type SceneSpec struct {
	Name     string
	Style    Style
	Build    func(hand, sans render.FontFace) (*scene.Scene, animation.Animation, error)
	Duration time.Duration // approximate wall-clock from Build's animation
}

// All returns the full benchmark catalog (small/medium/large ×
// crisp/sketchy). 6 scenes total.
func All() []SceneSpec {
	return []SceneSpec{
		{Name: "small", Style: StyleCrisp, Build: buildSmall(StyleCrisp), Duration: 5 * time.Second},
		{Name: "small", Style: StyleSketchy, Build: buildSmall(StyleSketchy), Duration: 5 * time.Second},
		{Name: "medium", Style: StyleCrisp, Build: buildMedium(StyleCrisp), Duration: 30 * time.Second},
		{Name: "medium", Style: StyleSketchy, Build: buildMedium(StyleSketchy), Duration: 30 * time.Second},
		{Name: "large", Style: StyleCrisp, Build: buildLarge(StyleCrisp), Duration: 90 * time.Second},
		{Name: "large", Style: StyleSketchy, Build: buildLarge(StyleSketchy), Duration: 90 * time.Second},
	}
}

func presetFor(s Style) (style.Style, color.Color) {
	switch s {
	case StyleSketchy:
		return style.PresetSketchy, color.RGBA{0xFF, 0xF8, 0xE1, 0xFF}
	default:
		return style.PresetCrisp, color.RGBA{0xFF, 0xFF, 0xFF, 0xFF}
	}
}

// asRevealer is a convenience wrapper for animation.DrawOn targets.
func asRevealer(m mobject.Mobject) animation.Revealer {
	if r, ok := m.(animation.Revealer); ok {
		return r
	}
	panic("scenes: mobject is not a Revealer")
}

// buildSmall — 3 components, 5 seconds, ~300 frames @ 60fps.
// Animations: stagger FadeIn, LaserPointer trace, FadeOut.
func buildSmall(st Style) func(hand, sans render.FontFace) (*scene.Scene, animation.Animation, error) {
	return func(hand, sans render.FontFace) (*scene.Scene, animation.Animation, error) {
		preset, bg := presetFor(st)
		r := render.NewCanvasRenderer(render.Options{Supersample: 1, DefaultFont: hand})
		s := scene.NewScene(1920, 1080).
			WithRenderer(r).
			WithDefaultStyle(preset).
			WithFont(style.FontHandDrawn, hand).
			WithFont(style.FontSans, sans)
		s.BgColor = bg

		client := icons.NewClient(101, "Client")
		server := icons.NewServer(102, "Server")
		db := icons.NewDatabase(103, "Database")
		row := layout.NewHBox(client, server, db).WithSpacing(180).MoveTo(0, 0)
		s.Add(row)
		_ = row.Bounds()

		for _, m := range []interface{ SetReveal(float64) }{client, server, db} {
			m.SetReveal(0)
		}

		// Total ≈ 5s: 1.5s stagger + 0.5s pause + 2.5s laser + 0.5s fade
		anim := animation.Sequence(
			animation.Stagger(180*time.Millisecond,
				animation.DrawOn(asRevealer(client), 500*time.Millisecond),
				animation.DrawOn(asRevealer(server), 500*time.Millisecond),
				animation.DrawOn(asRevealer(db), 500*time.Millisecond),
			),
			direction.Pause(500*time.Millisecond),
			direction.LaserPointer(direction.PathThrough(client, server, db), 2500*time.Millisecond),
			animation.Parallel(
				animation.FadeOut(client, 500*time.Millisecond),
				animation.FadeOut(server, 500*time.Millisecond),
				animation.FadeOut(db, 500*time.Millisecond),
			),
		)
		return s, anim, nil
	}
}

// buildMedium — 10 components, 30s, ~1800 frames.
// Top row: Client, Gateway, LB. Bottom row: 3 servers, Cache, Database, Queue.
// Animations: Stagger DrawOn × 10, Connect arrows, 3× laser, 2× pulse, 2× spotlight,
// 1× Camera ZoomTo + Reset, Caption.
func buildMedium(st Style) func(hand, sans render.FontFace) (*scene.Scene, animation.Animation, error) {
	return func(hand, sans render.FontFace) (*scene.Scene, animation.Animation, error) {
		preset, bg := presetFor(st)
		r := render.NewCanvasRenderer(render.Options{Supersample: 1, DefaultFont: hand})
		cam := direction.NewCamera()
		s := scene.NewScene(1920, 1080).
			WithRenderer(r).
			WithDefaultStyle(preset).
			WithFont(style.FontHandDrawn, hand).
			WithFont(style.FontSans, sans).
			WithCamera(cam)
		s.BgColor = bg

		client := icons.NewClient(201, "Client")
		gw := icons.NewAPIGateway(202, "GW")
		lb := icons.NewLoadBalancer(203, "LB")
		topRow := layout.NewHBox(client, gw, lb).WithSpacing(110).MoveTo(0, 220)

		api1 := icons.NewServer(211, "api-1")
		api2 := icons.NewServer(212, "api-2")
		api3 := icons.NewServer(213, "api-3")
		cache := icons.NewCache(214, "cache")
		db := icons.NewDatabase(215, "db")
		q := icons.NewQueue(216, "q")
		botRow := layout.NewHBox(api1, api2, api3, cache, db, q).WithSpacing(50).MoveTo(0, -80)
		s.Add(topRow)
		s.Add(botRow)
		_ = topRow.Bounds()
		_ = botRow.Bounds()

		nodes := []interface{ SetReveal(float64) }{client, gw, lb, api1, api2, api3, cache, db, q}
		for _, m := range nodes {
			m.SetReveal(0)
		}

		// Arrows.
		a1 := systemdesign.NewArrow(301, client, gw)
		a2 := systemdesign.NewArrow(302, gw, lb)
		a3 := systemdesign.NewArrow(303, lb, api1).WithRouting(systemdesign.RoutingOrthogonal)
		a4 := systemdesign.NewArrow(304, lb, api2).WithRouting(systemdesign.RoutingOrthogonal)
		a5 := systemdesign.NewArrow(305, lb, api3).WithRouting(systemdesign.RoutingOrthogonal)
		a6 := systemdesign.NewArrow(306, api2, cache).WithRouting(systemdesign.RoutingOrthogonal)
		a7 := systemdesign.NewArrow(307, cache, db).WithLabel("miss")
		a8 := systemdesign.NewArrow(308, api3, q).WithRouting(systemdesign.RoutingOrthogonal)
		arrows := []*systemdesign.Arrow{a1, a2, a3, a4, a5, a6, a7, a8}
		for _, a := range arrows {
			a.SetReveal(0)
			s.Add(a)
		}

		stagger := animation.Stagger(120*time.Millisecond,
			animation.DrawOn(asRevealer(client), 500*time.Millisecond),
			animation.DrawOn(asRevealer(gw), 500*time.Millisecond),
			animation.DrawOn(asRevealer(lb), 500*time.Millisecond),
			animation.DrawOn(asRevealer(api1), 500*time.Millisecond),
			animation.DrawOn(asRevealer(api2), 500*time.Millisecond),
			animation.DrawOn(asRevealer(api3), 500*time.Millisecond),
			animation.DrawOn(asRevealer(cache), 500*time.Millisecond),
			animation.DrawOn(asRevealer(db), 500*time.Millisecond),
			animation.DrawOn(asRevealer(q), 500*time.Millisecond),
		)
		arrowDraw := make([]animation.Animation, 0, len(arrows))
		for _, a := range arrows {
			arrowDraw = append(arrowDraw, animation.DrawOn(a, 600*time.Millisecond))
		}

		// Timeline ≈ 30s:
		//   ~2.0s stagger draw
		//   ~0.7s arrows
		//   ~1.0s caption + pause
		//   ~6s laser × 3
		//   ~1.5s pulse × 2
		//   ~6s spotlight + pause + remove
		//   ~3s camera zoom + reset
		//   ~1.5s caption
		//   ~8s pauses for slack
		anim := animation.Sequence(
			stagger,
			animation.Parallel(arrowDraw...),
			direction.Caption("medium-scene benchmark", 1*time.Second),
			direction.Pause(500*time.Millisecond),
			direction.LaserPointer(direction.PathThrough(client, gw, lb, api2), 2*time.Second),
			direction.Pause(300*time.Millisecond),
			direction.LaserPointer(direction.PathThrough(api2, cache, db), 2*time.Second),
			direction.Pause(300*time.Millisecond),
			direction.LaserPointer(direction.PathThrough(api3, q), 1500*time.Millisecond),
			direction.Pulse(cache, 3, 1500*time.Millisecond),
			direction.Pause(500*time.Millisecond),
			direction.Pulse(db, 2, 1*time.Second),
			direction.Spotlight(cam, lb, 500*time.Millisecond),
			direction.Pause(2*time.Second),
			direction.RemoveSpotlight(cam, 500*time.Millisecond),
			cam.ZoomTo(cache, 1.5, 1*time.Second),
			direction.Pause(1*time.Second),
			cam.Reset(1*time.Second),
			direction.Caption("end of medium", 1500*time.Millisecond),
			direction.Pause(3500*time.Millisecond),
		)
		return s, anim, nil
	}
}

// buildLarge — 20 components, 90s, ~5400 frames.
// Multi-row architecture similar to the URL-shortener pattern.
// Lots of pauses (~20s total) to exercise the pause-frame path.
func buildLarge(st Style) func(hand, sans render.FontFace) (*scene.Scene, animation.Animation, error) {
	return func(hand, sans render.FontFace) (*scene.Scene, animation.Animation, error) {
		preset, bg := presetFor(st)
		r := render.NewCanvasRenderer(render.Options{Supersample: 1, DefaultFont: hand})
		cam := direction.NewCamera()
		s := scene.NewScene(1920, 1080).
			WithRenderer(r).
			WithDefaultStyle(preset).
			WithFont(style.FontHandDrawn, hand).
			WithFont(style.FontSans, sans).
			WithCamera(cam)
		s.BgColor = bg

		// Endpoints
		mobile := icons.NewMobileClient(101, "mobile")
		web := icons.NewClient(102, "web")
		iot := icons.NewIoTDevice(103, "iot")
		clientCol := layout.NewVBox(mobile, web, iot).WithSpacing(30).MoveTo(-820, 0)

		// Edge
		cdn := icons.NewCDN(201, "cdn")
		fw := icons.NewFirewall(202, "fw")
		gw := icons.NewAPIGateway(203, "gw")
		edgeCol := layout.NewVBox(cdn, fw, gw).WithSpacing(30).MoveTo(-450, 0)

		// Services
		lb := icons.NewLoadBalancer(301, "lb")
		api1 := icons.NewServer(302, "api-1")
		api2 := icons.NewServer(303, "api-2")
		api3 := icons.NewServer(304, "api-3")
		worker := icons.NewWorker(305, "worker")
		svcCol := layout.NewVBox(lb, api1, api2, api3, worker).WithSpacing(20).MoveTo(-90, 0)

		// Stream + storage
		stream := icons.NewEventStream(401, "kafka")
		cache := icons.NewCache(402, "cache")
		kv := icons.NewKeyValueStore(403, "kv")
		streamCol := layout.NewVBox(stream, cache, kv).WithSpacing(30).MoveTo(240, 0)

		db := icons.NewDatabase(501, "db")
		tsdb := icons.NewTimeSeriesDB(502, "tsdb")
		dw := icons.NewDataWarehouse(503, "warehouse")
		objs := icons.NewObjectStorage(504, "objects")
		dataCol := layout.NewVBox(db, tsdb, dw, objs).WithSpacing(25).MoveTo(580, 0)

		// Observability sidecar
		metrics := icons.NewMetrics(601, "metrics")
		logs := icons.NewLogs(602, "logs")
		obsCol := layout.NewVBox(metrics, logs).WithSpacing(30).MoveTo(880, 0)

		for _, c := range []*layout.VBox{clientCol, edgeCol, svcCol, streamCol, dataCol, obsCol} {
			s.Add(c)
			_ = c.Bounds()
		}

		nodes := []mobject.Mobject{
			mobile, web, iot, cdn, fw, gw, lb, api1, api2, api3, worker,
			stream, cache, kv, db, tsdb, dw, objs, metrics, logs,
		}
		for _, n := range nodes {
			if r, ok := n.(interface{ SetReveal(float64) }); ok {
				r.SetReveal(0)
			}
		}

		drawAll := make([]animation.Animation, 0, len(nodes))
		for _, n := range nodes {
			drawAll = append(drawAll, animation.DrawOn(asRevealer(n), 400*time.Millisecond))
		}

		// Long timeline — 90s. Heavy in pauses + laser tracing + camera
		// moves + spotlights, which are the workloads the URL-shortener
		// teaching video would exercise. Tuned to hit ~5400 frames.
		anim := animation.Sequence(
			animation.Stagger(80*time.Millisecond, drawAll...),                // ~1.7s
			direction.Pause(2*time.Second),                                    // 2s
			direction.Caption("large-scene benchmark", 1500*time.Millisecond), // 1.5s
			direction.LaserPointer(direction.PathThrough(mobile, gw, lb, api2), 2500*time.Millisecond),
			direction.Pause(1*time.Second),
			direction.LaserPointer(direction.PathThrough(api2, cache, kv), 2*time.Second),
			direction.Pause(1*time.Second),
			direction.LaserPointer(direction.PathThrough(api3, stream, tsdb), 2500*time.Millisecond),
			direction.Pause(1*time.Second),
			cam.ZoomTo(cache, 1.6, 1500*time.Millisecond),
			direction.Pulse(cache, 4, 2*time.Second),
			direction.Callout(cache, "hot URLs cached", direction.CalloutAbove, 2500*time.Millisecond),
			cam.Reset(1500*time.Millisecond),
			direction.Pause(2*time.Second),
			direction.Spotlight(cam, lb, 600*time.Millisecond),
			direction.Pause(2500*time.Millisecond),
			direction.RemoveSpotlight(cam, 600*time.Millisecond),
			direction.LaserPointer(direction.PathThrough(worker, dw, objs), 2*time.Second),
			direction.Pause(1500*time.Millisecond),
			direction.Pulse(metrics, 3, 1500*time.Millisecond),
			direction.Pulse(logs, 3, 1500*time.Millisecond),
			direction.Pause(2*time.Second),
			cam.ZoomTo(stream, 1.4, 1500*time.Millisecond),
			direction.Pause(2*time.Second),
			cam.Reset(1500*time.Millisecond),
			direction.Caption("end of large", 2*time.Second),
			direction.Pause(4*time.Second), // intentional long pause — exercises the pause path
			animation.Parallel(
				animation.FadeOut(mobile, 500*time.Millisecond),
				animation.FadeOut(web, 500*time.Millisecond),
				animation.FadeOut(iot, 500*time.Millisecond),
			),
			direction.Pause(8*time.Second), // long final pause — half the scene becomes pauses
		)
		return s, anim, nil
	}
}
