// Command animated_system_diagram builds a non-trivial system-design
// diagram composed entirely through layout helpers (no manual
// coordinates beyond the scene origin), connects components with
// smart-routed arrows of mixed types, and animates a multi-hop
// packet flowing through the diagram.
//
// The same choreography is rendered in two presets:
//
//	out_crisp.mp4    — Architect sloppiness, sharp edges, sans font
//	out_sketchy.mp4  — Cartoonist sloppiness, round edges, hand-drawn font
//
// Run:
//
//	go run ./examples/system_diagram_animated
package main

import (
	"fmt"
	"image"
	"image/color"
	"os"
	"time"

	"github.com/ankitsinghchadda/goanim/core/animation"
	"github.com/ankitsinghchadda/goanim/core/layout"
	"github.com/ankitsinghchadda/goanim/core/mobject"
	"github.com/ankitsinghchadda/goanim/core/render"
	"github.com/ankitsinghchadda/goanim/core/scene"
	"github.com/ankitsinghchadda/goanim/core/style"
	"github.com/ankitsinghchadda/goanim/mobjects/icons"
	"github.com/ankitsinghchadda/goanim/mobjects/systemdesign"
)

const (
	W   = 1920
	H   = 1080
	FPS = 60
)

func main() {
	hand, err := render.Excalifont()
	must(err, "excalifont")
	sans, err := render.Inter()
	must(err, "inter")

	for _, v := range variants() {
		fmt.Printf("rendering %s...\n", v.name)
		out := "out_" + v.name + ".mp4"
		if err := renderVariant(v, hand, sans, out); err != nil {
			fmt.Fprintln(os.Stderr, "failed:", err)
			os.Exit(1)
		}
		fmt.Println("wrote", out)
	}
}

type variant struct {
	name  string
	style style.Style
	bg    color.Color
}

func variants() []variant {
	all := []variant{
		{"crisp", style.PresetCrisp, color.RGBA{0xFF, 0xFF, 0xFF, 0xFF}},
		{"sketchy", style.PresetSketchy, color.RGBA{0xFF, 0xF8, 0xE1, 0xFF}},
	}
	if only := os.Getenv("VARIANT"); only != "" {
		for _, v := range all {
			if v.name == only {
				return []variant{v}
			}
		}
	}
	return all
}

func renderVariant(v variant, hand, sans render.FontFace, outPath string) error {
	r := render.NewCanvasRenderer(render.Options{Supersample: 2, DefaultFont: hand})
	s := scene.NewScene(W, H).
		WithRenderer(r).
		WithDefaultStyle(v.style).
		WithFont(style.FontHandDrawn, hand).
		WithFont(style.FontSans, sans)
	s.BgColor = v.bg
	s.FPS = FPS

	// --- Build the diagram using layout helpers -------------------------------
	//
	// Top row:    User → APIGateway → LoadBalancer
	// Middle row: API-1, API-2, API-3 (fed by LB)
	// Bottom row: Cache → Postgres   |   Queue → Worker-1, Worker-2
	//
	// All laid out via HBox/VBox; no manual coordinates needed.

	usr := icons.NewUser(1, "User")
	gw := icons.NewAPIGateway(2, "API Gateway")
	lb := icons.NewLoadBalancer(3, "LB")
	topRow := layout.NewHBox(usr, gw, lb).WithSpacing(70)

	api1 := icons.NewServer(11, "API-1")
	api2 := icons.NewServer(12, "API-2")
	api3 := icons.NewServer(13, "API-3")
	midRow := layout.NewHBox(api1, api2, api3).WithSpacing(40)

	cache := icons.NewCache(21, "Cache")
	pg := icons.NewDatabase(22, "Postgres")
	dataPair := layout.NewHBox(cache, pg).WithSpacing(70)

	queue := icons.NewQueue(31, "Job Queue")
	w1 := icons.NewWorker(32, "Worker-1")
	w2 := icons.NewWorker(33, "Worker-2")
	workers := layout.NewHBox(w1, w2).WithSpacing(40)
	queueWorkers := layout.NewVBox(queue, workers).WithSpacing(40)

	bottomRow := layout.NewHBox(dataPair, queueWorkers).WithSpacing(120)

	column := layout.NewVBox(topRow, midRow, bottomRow).WithSpacing(80).MoveTo(0, -20)
	s.Add(column)

	// --- Arrows ---------------------------------------------------------------
	//
	// Lay out positions by calling Bounds() on the column (triggers layout).
	_ = column.Bounds()

	// Top-row arrows: User → Gateway → LB (straight, auto-routed).
	a1 := systemdesign.NewArrow(101, usr, gw)
	a2 := systemdesign.NewArrow(102, gw, lb)

	// LB → each API (orthogonal — LB is above APIs).
	a3 := systemdesign.NewArrow(103, lb, api1).WithRouting(systemdesign.RoutingOrthogonal)
	a4 := systemdesign.NewArrow(104, lb, api2).WithRouting(systemdesign.RoutingOrthogonal)
	a5 := systemdesign.NewArrow(105, lb, api3).WithRouting(systemdesign.RoutingOrthogonal)

	// API-2 → Cache, with label.
	a6 := systemdesign.NewArrow(106, api2, cache).
		WithRouting(systemdesign.RoutingOrthogonal).
		WithLabel("read")

	// Cache → Postgres (cache miss).
	a7 := systemdesign.NewArrow(107, cache, pg).WithLabel("miss")

	// API-2 → Queue (writes).
	a8 := systemdesign.NewArrow(108, api2, queue).
		WithRouting(systemdesign.RoutingOrthogonal).
		WithLabel("write")

	// Queue → Workers (fan-out).
	a9 := systemdesign.NewArrow(109, queue, w1).WithRouting(systemdesign.RoutingOrthogonal)
	a10 := systemdesign.NewArrow(110, queue, w2).WithRouting(systemdesign.RoutingOrthogonal)

	// Workers → Postgres.
	a11 := systemdesign.NewArrow(111, w1, pg).
		WithRouting(systemdesign.RoutingCurved)
	a12 := systemdesign.NewArrow(112, w2, pg).
		WithRouting(systemdesign.RoutingCurved)

	// Hide all arrows initially.
	allArrows := []*systemdesign.Arrow{a1, a2, a3, a4, a5, a6, a7, a8, a9, a10, a11, a12}
	for _, a := range allArrows {
		a.SetReveal(0)
		s.Add(a)
	}

	// Hide all nodes initially.
	allNodes := []mobject.Mobject{usr, gw, lb, api1, api2, api3, cache, pg, queue, w1, w2}
	for _, n := range allNodes {
		if r, ok := n.(interface{ SetReveal(float64) }); ok {
			r.SetReveal(0)
		}
	}

	// Packet for the multi-hop flow.
	packet := systemdesign.NewPacket(999, "GET /user")
	// Place at user's position initially.
	{
		x, y := positionOf(usr)
		packet.MoveTo(x, y)
		packet.SetReveal(0)
	}
	s.Add(packet)

	// --- Encoder --------------------------------------------------------------
	enc, err := render.NewVideoEncoder(outPath, render.VideoOptions{
		Width: W, Height: H, FPS: FPS, CRF: 18, Preset: "slow",
	})
	if err != nil {
		return err
	}
	defer enc.Close()
	sink := frameSink{enc: enc}

	// --- Animation ------------------------------------------------------------
	// 0–1.5s: stagger-draw nodes top → bottom.
	nodeAnims := []animation.Animation{
		animation.DrawOn(asRevealer(usr), 600*time.Millisecond),
		animation.DrawOn(asRevealer(gw), 600*time.Millisecond),
		animation.DrawOn(asRevealer(lb), 600*time.Millisecond),
		animation.DrawOn(asRevealer(api1), 600*time.Millisecond),
		animation.DrawOn(asRevealer(api2), 600*time.Millisecond),
		animation.DrawOn(asRevealer(api3), 600*time.Millisecond),
		animation.DrawOn(asRevealer(cache), 600*time.Millisecond),
		animation.DrawOn(asRevealer(pg), 600*time.Millisecond),
		animation.DrawOn(asRevealer(queue), 600*time.Millisecond),
		animation.DrawOn(asRevealer(w1), 600*time.Millisecond),
		animation.DrawOn(asRevealer(w2), 600*time.Millisecond),
	}
	if _, err := s.Play(sink, animation.Stagger(80*time.Millisecond, nodeAnims...)); err != nil {
		return err
	}

	// 1.5–2.3s: arrows draw on in parallel.
	arrowAnims := make([]animation.Animation, 0, len(allArrows))
	for _, a := range allArrows {
		arrowAnims = append(arrowAnims, animation.DrawOn(a, 700*time.Millisecond))
	}
	if _, err := s.Play(sink, animation.Parallel(arrowAnims...)); err != nil {
		return err
	}

	// 2.3–4.5s: packet flows User → Gateway → LB → API-2 → Cache → Postgres,
	// then a write hop to the Queue.
	hops := []mobject.Mobject{usr, gw, lb, api2, cache, pg}
	if _, err := s.Play(sink, animation.PopIn(packet, 200*time.Millisecond)); err != nil {
		return err
	}
	for i := 1; i < len(hops); i++ {
		x, y := positionOf(hops[i])
		if _, err := s.Play(sink, animation.MoveTo(packet, x, y, 350*time.Millisecond)); err != nil {
			return err
		}
		// Flash the destination node briefly.
		if _, err := s.Play(sink,
			animation.Flash(hops[i], color.RGBA{0xF5, 0x9F, 0x00, 0xFF}, 200*time.Millisecond),
		); err != nil {
			return err
		}
	}

	// 4.5–5s: fade the packet out, hold final state briefly.
	if _, err := s.Play(sink, animation.FadeOut(packet, 250*time.Millisecond)); err != nil {
		return err
	}
	if _, err := s.PlayStill(sink, 900*time.Millisecond); err != nil {
		return err
	}
	return nil
}

// frameSink adapts a VideoEncoder to scene.FrameWriter.
type frameSink struct{ enc *render.VideoEncoder }

func (f frameSink) WriteFrame(img image.Image) error { return f.enc.WriteFrame(img) }

func must(err error, msg string) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", msg, err)
		os.Exit(1)
	}
}

// asRevealer reflects a Mobject as a Revealer (icons all satisfy this).
func asRevealer(m mobject.Mobject) animation.Revealer {
	if r, ok := m.(animation.Revealer); ok {
		return r
	}
	panic("animated_system_diagram: mobject does not implement Revealer")
}

// positionOf reflects a Positioner; falls back to bounding-box center.
func positionOf(m mobject.Mobject) (float64, float64) {
	if p, ok := m.(interface{ Position() (float64, float64) }); ok {
		return p.Position()
	}
	c := m.Bounds().Center()
	return c.X, c.Y
}
