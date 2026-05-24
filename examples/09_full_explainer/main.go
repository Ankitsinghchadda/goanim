// Command url_shortener is the Phase-7 integration test: a 90-second
// scripted teaching video that exercises every direction-layer
// primitive in real-world combinations. If every feature works
// individually but they don't compose into this video, the phase
// isn't done. (Quoting the prompt.)
//
// Renders:
//
//	out_url_shortener_crisp.mp4
//	out_url_shortener_sketchy.mp4
//
// Both at 1920×1080, 60fps. The script:
//
//  1. Title              — fade in "Designing a URL Shortener"
//  2. The problem         — long URL → short URL example
//  3. Naive approach      — Client/Server/Database with laser + callout
//  4. Add redundancy      — LB + multi-server + replicas + spotlight
//  5. Read-heavy opt      — Cache + laser + Replay
//  6. Hot URLs            — zoom on Cache + pulse + callout
//  7. Summary             — pulse cascade across the full architecture
package main

import (
	"fmt"
	"image"
	"image/color"
	"os"
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
		out := "out_url_shortener_" + v.name + ".mp4"
		if err := renderVideo(v, hand, sans, out); err != nil {
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

func renderVideo(v variant, hand, sans render.FontFace, outPath string) error {
	r := render.NewCanvasRenderer(render.Options{Supersample: 2, DefaultFont: hand})
	cam := direction.NewCamera()
	s := scene.NewScene(W, H).
		WithRenderer(r).
		WithDefaultStyle(v.style).
		WithFont(style.FontHandDrawn, hand).
		WithFont(style.FontSans, sans).
		WithCamera(cam)
	s.BgColor = v.bg
	s.FPS = FPS

	enc, err := render.NewVideoEncoder(outPath, render.VideoOptions{
		Width: W, Height: H, FPS: FPS, CRF: 18, Preset: "fast",
	})
	if err != nil {
		return err
	}
	defer enc.Close()
	sink := frameSink{enc: enc}

	if err := sceneTitle(s, sink); err != nil {
		return err
	}
	if err := sceneProblem(s, sink); err != nil {
		return err
	}
	if err := sceneNaive(s, sink, cam); err != nil {
		return err
	}
	if err := sceneRedundancy(s, sink, cam); err != nil {
		return err
	}
	return nil
}

// 0:00–0:05 — Title.
func sceneTitle(s *scene.Scene, sink scene.FrameWriter) error {
	// Position above frame center; Inter at XLarge has a tall ascent
	// so y=180 puts the visual middle near frame center.
	title := mobject.NewText(1, "Designing a URL Shortener").MoveTo(0, 180)
	zero := 0.0
	title.SetStyle(style.Style{FontSize: style.FontXLarge, FontFamily: style.FontSans, Opacity: &zero})
	subtitle := mobject.NewText(2, "system design walkthrough").MoveTo(0, 60)
	subtitle.SetStyle(style.Style{FontSize: style.FontMedium, FontFamily: style.FontSans, Opacity: &zero})
	s.Add(title, subtitle)

	if _, err := s.Play(sink, animation.Parallel(
		animation.FadeIn(title, 800*time.Millisecond),
		animation.Sequence(direction.Pause(200*time.Millisecond), animation.FadeIn(subtitle, 600*time.Millisecond)),
	)); err != nil {
		return err
	}
	if _, err := s.Play(sink, direction.Pause(2*time.Second)); err != nil {
		return err
	}
	if _, err := s.Play(sink, animation.Parallel(
		animation.FadeOut(title, 500*time.Millisecond),
		animation.FadeOut(subtitle, 500*time.Millisecond),
	)); err != nil {
		return err
	}
	clearScene(s)
	return nil
}

// clearScene wipes all mobjects from the scene. Used between scenes
// so leftover Text/Arrow labels (which don't always honor the parent
// mobject's Opacity) don't bleed across cuts.
func clearScene(s *scene.Scene) {
	s.Mobjects = s.Mobjects[:0]
}

// 0:05–0:15 — The problem: long URL → short URL.
func sceneProblem(s *scene.Scene, sink scene.FrameWriter) error {
	zero := 0.0
	longURL := mobject.NewText(10, "https://www.example.com/articles/2024/...").MoveTo(-500, 0)
	longURL.SetStyle(style.Style{FontSize: style.FontMedium, FontFamily: style.FontSans, Opacity: &zero})
	shortURL := mobject.NewText(11, "exa.mpl/k7d9").MoveTo(560, 0)
	shortURL.SetStyle(style.Style{FontSize: style.FontLarge, FontFamily: style.FontSans, Opacity: &zero})
	s.Add(longURL, shortURL)

	if _, err := s.Play(sink, animation.FadeIn(longURL, 600*time.Millisecond)); err != nil {
		return err
	}
	// Arrow between them.
	arrow := systemdesign.NewArrow(20, longURL, shortURL).WithLabel("shorten")
	arrow.SetReveal(0)
	s.Add(arrow)
	defer clearScene(s)
	if _, err := s.Play(sink, animation.Parallel(
		animation.DrawOn(arrow, 800*time.Millisecond),
		animation.FadeIn(shortURL, 600*time.Millisecond),
	)); err != nil {
		return err
	}
	if _, err := s.Play(sink, direction.Pause(2*time.Second)); err != nil {
		return err
	}
	// Clear for next scene.
	if _, err := s.Play(sink, animation.Parallel(
		animation.FadeOut(longURL, 500*time.Millisecond),
		animation.FadeOut(shortURL, 500*time.Millisecond),
		animation.FadeOut(arrow, 500*time.Millisecond),
	)); err != nil {
		return err
	}
	return nil
}

// 0:15–0:30 — Naive approach: Client → Server → Database, laser
// trace, zoom on Database with callout.
func sceneNaive(s *scene.Scene, sink scene.FrameWriter, cam *direction.Camera) error {
	caption := direction.Caption("First attempt: simple service", 4*time.Second)

	client := icons.NewClient(100, "Client")
	server := icons.NewServer(101, "Server")
	db := icons.NewDatabase(102, "Database")
	row := layout.NewHBox(client, server, db).WithSpacing(180).MoveTo(0, -20)
	s.Add(row)
	_ = row.Bounds()

	a1 := systemdesign.NewArrow(110, client, server)
	a2 := systemdesign.NewArrow(111, server, db)
	s.Add(a1, a2)
	defer clearScene(s)

	for _, m := range []interface{ SetReveal(float64) }{client, server, db, a1, a2} {
		m.SetReveal(0)
	}

	// Caption runs in parallel with stagger-in.
	if _, err := s.Play(sink, animation.Parallel(
		caption,
		animation.Stagger(180*time.Millisecond,
			animation.DrawOn(asRev(client), 500*time.Millisecond),
			animation.DrawOn(asRev(server), 500*time.Millisecond),
			animation.DrawOn(asRev(db), 500*time.Millisecond),
			animation.DrawOn(a1, 500*time.Millisecond),
			animation.DrawOn(a2, 500*time.Millisecond),
		),
	)); err != nil {
		return err
	}

	if _, err := s.Play(sink, direction.Pause(1*time.Second)); err != nil {
		return err
	}

	// Laser trace: Client → Server → Database → Server → Client.
	if _, err := s.Play(sink, direction.LaserPointer(
		direction.PathThrough(client, server, db, server, client),
		3*time.Second,
	)); err != nil {
		return err
	}

	// Zoom on Database with a Callout "single point of failure".
	if _, err := s.Play(sink, cam.ZoomTo(db, 1.5, 1*time.Second)); err != nil {
		return err
	}
	if _, err := s.Play(sink, direction.Callout(db,
		"single point of failure!", direction.CalloutBelow,
		2500*time.Millisecond,
	)); err != nil {
		return err
	}
	if _, err := s.Play(sink, cam.Reset(1*time.Second)); err != nil {
		return err
	}
	return nil
}

// 0:30–0:50 — Add redundancy: LB + 3 Servers + replicated DB.
func sceneRedundancy(s *scene.Scene, sink scene.FrameWriter, cam *direction.Camera) error {
	// Find the existing client/server/db by snapshotting before adding
	// new nodes. The scene mobject order is: ... (title/subtitle/long/
	// short/arrow all faded; then row, a1, a2 from sceneNaive). We don't
	// reach into them by index — instead we add fresh icons offset
	// horizontally and rebuild the layout.
	//
	// For simplicity, fade out the naive scene and build the redundant
	// design fresh.
	caption := direction.Caption("Add redundancy", 4*time.Second)

	// Cross-fade single server → three servers.
	client := icons.NewClient(200, "Client")
	lb := icons.NewLoadBalancer(201, "LB")
	api1 := icons.NewServer(211, "API-1")
	api2 := icons.NewServer(212, "API-2")
	api3 := icons.NewServer(213, "API-3")
	apiCol := layout.NewVBox(api1, api2, api3).WithSpacing(30)
	primaryDB := icons.NewDatabase(221, "primary")
	replica1 := icons.NewDatabase(222, "replica")
	replica2 := icons.NewDatabase(223, "replica")
	dbCol := layout.NewVBox(primaryDB, replica1, replica2).WithSpacing(20)
	row := layout.NewHBox(client, lb, apiCol, dbCol).WithSpacing(120).MoveTo(0, 0)
	s.Add(row)
	_ = row.Bounds()

	for _, m := range []interface{ SetReveal(float64) }{client, lb, api1, api2, api3, primaryDB, replica1, replica2} {
		m.SetReveal(0)
	}

	// Stagger appear all components alongside caption.
	if _, err := s.Play(sink, animation.Parallel(
		caption,
		animation.Stagger(120*time.Millisecond,
			animation.DrawOn(asRev(client), 500*time.Millisecond),
			animation.DrawOn(asRev(lb), 500*time.Millisecond),
			animation.DrawOn(asRev(api1), 500*time.Millisecond),
			animation.DrawOn(asRev(api2), 500*time.Millisecond),
			animation.DrawOn(asRev(api3), 500*time.Millisecond),
			animation.DrawOn(asRev(primaryDB), 500*time.Millisecond),
			animation.DrawOn(asRev(replica1), 500*time.Millisecond),
			animation.DrawOn(asRev(replica2), 500*time.Millisecond),
		),
	)); err != nil {
		return err
	}

	// Arrows: client → LB, LB → each API, each API → primary DB,
	// primary → each replica (curved).
	a1 := systemdesign.NewArrow(230, client, lb)
	a2 := systemdesign.NewArrow(231, lb, api1).WithRouting(systemdesign.RoutingOrthogonal)
	a3 := systemdesign.NewArrow(232, lb, api2).WithRouting(systemdesign.RoutingOrthogonal)
	a4 := systemdesign.NewArrow(233, lb, api3).WithRouting(systemdesign.RoutingOrthogonal)
	a5 := systemdesign.NewArrow(234, api2, primaryDB).WithRouting(systemdesign.RoutingOrthogonal)
	a6 := systemdesign.NewArrow(235, primaryDB, replica1).WithRouting(systemdesign.RoutingCurved).WithLabel("repl")
	a7 := systemdesign.NewArrow(236, primaryDB, replica2).WithRouting(systemdesign.RoutingCurved).WithLabel("repl")
	for _, a := range []*systemdesign.Arrow{a1, a2, a3, a4, a5, a6, a7} {
		a.SetReveal(0)
		s.Add(a)
	}
	if _, err := s.Play(sink, animation.Parallel(
		animation.DrawOn(a1, 500*time.Millisecond),
		animation.DrawOn(a2, 500*time.Millisecond),
		animation.DrawOn(a3, 500*time.Millisecond),
		animation.DrawOn(a4, 500*time.Millisecond),
		animation.DrawOn(a5, 500*time.Millisecond),
		animation.DrawOn(a6, 600*time.Millisecond),
		animation.DrawOn(a7, 600*time.Millisecond),
	)); err != nil {
		return err
	}

	// Spotlight on LB.
	if _, err := s.Play(sink, direction.Spotlight(cam, lb, 500*time.Millisecond)); err != nil {
		return err
	}
	if _, err := s.Play(sink, direction.Pause(2*time.Second)); err != nil {
		return err
	}
	if _, err := s.Play(sink, direction.RemoveSpotlight(cam, 500*time.Millisecond)); err != nil {
		return err
	}

	// Hold the final redundancy state for a beat. (Subsequent scenes
	// not implemented this session — see status report. The video
	// gracefully ends here.)
	if _, err := s.Play(sink, direction.Pause(2*time.Second)); err != nil {
		return err
	}
	return nil
}

// --------- plumbing ---------

type frameSink struct{ enc *render.VideoEncoder }

func (f frameSink) WriteFrame(img image.Image) error { return f.enc.WriteFrame(img) }

func must(err error, msg string) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", msg, err)
		os.Exit(1)
	}
}

func asRev(m interface{}) animation.Revealer {
	if r, ok := m.(animation.Revealer); ok {
		return r
	}
	panic("url_shortener: not a Revealer")
}
