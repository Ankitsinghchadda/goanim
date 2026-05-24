// Command showcase renders goanim's promotional video — a tour of
// the library's capabilities. The README embeds the output.
//
// Chapters:
//
//  1. Title card    — Manim-for-engineers opening
//  2. Style parade  — the same diagram in 3 presets
//  3. System build  — staggered icons + arrows + camera focus
//  4. Math graphs   — Axes + plotted curves + shaded region
//  5. Math reveal   — LaTeX equation handwriting
//  6. Outro         — go get URL
//
// Run:
//
//	go run ./examples/showcase
//
// Output: out_showcase.mp4 (1920×1080, 60fps, ~50 s)
package main

import (
	"fmt"
	"image"
	"image/color"
	"math"
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
	"github.com/ankitsinghchadda/goanim/mobjects/mathx"
	"github.com/ankitsinghchadda/goanim/mobjects/systemdesign"
)

const (
	W   = 1920
	H   = 1080
	FPS = 60
)

func main() {
	hand, err := render.Excalifont()
	must(err)
	sans, err := render.Inter()
	must(err)

	r := render.NewCanvasRenderer(render.Options{DefaultFont: hand})
	cam := direction.NewCamera()
	s := scene.NewScene(W, H).
		WithRenderer(r).
		WithDefaultStyle(style.PresetSketchy).
		WithFont(style.FontHandDrawn, hand).
		WithFont(style.FontSans, sans).
		WithCamera(cam)
	s.BgColor = color.RGBA{0xFD, 0xF6, 0xE3, 0xFF}
	s.FPS = FPS

	enc, err := render.NewVideoEncoder("out_showcase.mp4", render.VideoOptions{
		Width: W, Height: H, FPS: FPS, CRF: 20, Preset: "fast",
	})
	must(err)
	defer enc.Close()
	sink := frameSink{enc: enc}

	chapterTitle(s, sink)
	chapterStyles(s, sink)
	chapterSystem(s, sink, cam)
	chapterGraphs(s, sink)
	chapterMath(s, sink)
	chapterOutro(s, sink)

	fmt.Println("wrote out_showcase.mp4")
}

// ---------- chapter 1 : title ----------------------------------------------

func chapterTitle(s *scene.Scene, sink scene.FrameWriter) {
	s.Mobjects = s.Mobjects[:0]

	// All text in this chapter renders in the scene's default
	// handwritten (Excalifont) style — no explicit FontFamily.

	// Opening hook: positions the library upfront.
	hook := mobject.NewText(0, "A Go library").MoveTo(0, 330)
	hook.SetStyle(style.Style{FontSize: style.FontHuge})

	// Wordmark — display-sized handwritten.
	title := mobject.NewText(1, "for animated videos").MoveTo(0, 80)
	title.SetStyle(style.Style{FontSize: style.FontDisplay})

	// What the videos cover.
	tagline := mobject.NewText(2, "system design + math, rendered from code").MoveTo(0, -200)
	tagline.SetStyle(style.Style{FontSize: style.FontHuge})

	// Why this library (vs Manim).
	subTagline := mobject.NewText(3, "no Python · no TeX install · just  go run").MoveTo(0, -360)
	subTagline.SetStyle(style.Style{FontSize: style.FontXLarge})

	zero := 0.0
	for _, t := range []*mobject.Text{hook, title, tagline, subTagline} {
		st := t.Style()
		st.Opacity = &zero
	}
	s.Add(hook, title, tagline, subTagline)

	_, _ = s.Play(sink, animation.Sequence(
		animation.FadeIn(hook, 600*time.Millisecond),
		animation.FadeIn(title, 900*time.Millisecond),
		direction.Pause(200*time.Millisecond),
		animation.FadeIn(tagline, 700*time.Millisecond),
		animation.FadeIn(subTagline, 500*time.Millisecond),
	))
	_, _ = s.Play(sink, direction.Pause(2200*time.Millisecond))
	_, _ = s.Play(sink, animation.Parallel(
		animation.FadeOut(hook, 500*time.Millisecond),
		animation.FadeOut(title, 500*time.Millisecond),
		animation.FadeOut(tagline, 500*time.Millisecond),
		animation.FadeOut(subTagline, 500*time.Millisecond),
	))
}

// ---------- chapter 2 : style parade ---------------------------------------

func chapterStyles(s *scene.Scene, sink scene.FrameWriter) {
	heading := mobject.NewText(10, "one diagram, every style").MoveTo(0, 430)
	heading.SetStyle(style.Style{FontSize: style.FontHuge})
	s.Mobjects = s.Mobjects[:0]
	s.Add(heading)
	_, _ = s.Play(sink, animation.FadeIn(heading, 400*time.Millisecond))

	presets := []struct {
		name  string
		st    style.Style
		bgCol color.Color
	}{
		{"Sketchy", style.PresetSketchy, color.RGBA{0xFD, 0xF6, 0xE3, 0xFF}},
		{"Crisp", style.PresetCrisp, color.RGBA{0xFF, 0xFF, 0xFF, 0xFF}},
		{"Excalidraw", style.PresetExcalidraw, color.RGBA{0xFF, 0xFF, 0xFF, 0xFF}},
	}

	for i, p := range presets {
		s.DefaultStyle = p.st
		s.BgColor = p.bgCol

		client := icons.NewClient(int64(100+i*10), "Client").MoveTo(-450, 0)
		server := icons.NewServer(int64(101+i*10), "Server").MoveTo(0, 0)
		database := icons.NewDatabase(int64(102+i*10), "Database").MoveTo(450, 0)
		a1 := systemdesign.NewArrow(int64(150+i*10), client, server)
		a2 := systemdesign.NewArrow(int64(151+i*10), server, database)
		label := mobject.NewText(int64(160+i*10), p.name).MoveTo(0, -360)
		label.SetStyle(style.Style{FontSize: style.FontXLarge})

		for _, m := range []interface{ SetReveal(float64) }{client, server, database, a1, a2} {
			m.SetReveal(0)
		}
		zero := 0.0
		ls := label.Style()
		ls.Opacity = &zero

		s.Add(client, server, database, a1, a2, label)

		_, _ = s.Play(sink, animation.Parallel(
			animation.FadeIn(label, 250*time.Millisecond),
			animation.Stagger(120*time.Millisecond,
				animation.DrawOn(client, 350*time.Millisecond),
				animation.DrawOn(server, 350*time.Millisecond),
				animation.DrawOn(database, 350*time.Millisecond),
				animation.DrawOn(a1, 350*time.Millisecond),
				animation.DrawOn(a2, 350*time.Millisecond),
			),
		))
		_, _ = s.Play(sink, direction.Pause(900*time.Millisecond))

		// Fade everything between presets except the heading.
		_, _ = s.Play(sink, animation.Parallel(
			animation.FadeOut(client, 400*time.Millisecond),
			animation.FadeOut(server, 400*time.Millisecond),
			animation.FadeOut(database, 400*time.Millisecond),
			animation.FadeOut(a1, 400*time.Millisecond),
			animation.FadeOut(a2, 400*time.Millisecond),
			animation.FadeOut(label, 400*time.Millisecond),
		))
		s.Mobjects = []mobject.Mobject{heading}
	}
	_, _ = s.Play(sink, animation.FadeOut(heading, 400*time.Millisecond))

	// Restore sketchy default for next chapters.
	s.DefaultStyle = style.PresetSketchy
	s.BgColor = color.RGBA{0xFD, 0xF6, 0xE3, 0xFF}
}

// ---------- chapter 3 : full system + camera focus -------------------------

func chapterSystem(s *scene.Scene, sink scene.FrameWriter, cam *direction.Camera) {
	s.Mobjects = s.Mobjects[:0]

	user := icons.NewUser(200, "User").MoveTo(-700, 200)
	gw := icons.NewAPIGateway(201, "Gateway").MoveTo(-300, 200)
	lb := icons.NewLoadBalancer(202, "LB").MoveTo(100, 200)
	api1 := icons.NewServer(210, "API-1").MoveTo(500, 340)
	api2 := icons.NewServer(211, "API-2").MoveTo(500, 60)
	cache := icons.NewCache(220, "Cache").MoveTo(820, 340)
	db := icons.NewDatabase(221, "Database").MoveTo(820, 60)

	all := []mobject.Mobject{user, gw, lb, api1, api2, cache, db}
	for _, m := range all {
		if r, ok := m.(interface{ SetReveal(float64) }); ok {
			r.SetReveal(0)
		}
	}

	a := []*systemdesign.Arrow{
		systemdesign.NewArrow(300, user, gw),
		systemdesign.NewArrow(301, gw, lb),
		systemdesign.NewArrow(302, lb, api1),
		systemdesign.NewArrow(303, lb, api2),
		systemdesign.NewArrow(304, api1, cache),
		systemdesign.NewArrow(305, api2, db),
	}
	for _, ar := range a {
		ar.SetReveal(0)
	}

	s.Add(all...)
	for _, ar := range a {
		s.Add(ar)
	}

	// Build the diagram with a staggered draw-on.
	revAnims := make([]animation.Animation, 0, len(all)+len(a))
	for _, m := range all {
		if rv, ok := m.(animation.Revealer); ok {
			revAnims = append(revAnims, animation.DrawOn(rv, 350*time.Millisecond))
		}
	}
	for _, ar := range a {
		revAnims = append(revAnims, animation.DrawOn(ar, 350*time.Millisecond))
	}
	_, _ = s.Play(sink, animation.Stagger(110*time.Millisecond, revAnims...))
	_, _ = s.Play(sink, direction.Pause(700*time.Millisecond))

	// Caption while we focus on the cache layer.
	_, _ = s.Play(sink, direction.Caption(
		"Camera.Focus dims everything else",
		2500*time.Millisecond))

	_, _ = s.Play(sink, cam.Focus(cache, 1.8, 900*time.Millisecond))
	_, _ = s.Play(sink, direction.Pause(1100*time.Millisecond))
	_, _ = s.Play(sink, cam.UnFocus(900*time.Millisecond))

	// Flow a request packet along the path with a laser pointer.
	_, _ = s.Play(sink, direction.LaserPointer(
		direction.PathThrough(user, gw, lb, api2, db),
		2200*time.Millisecond,
	))
	_, _ = s.Play(sink, direction.Pause(600*time.Millisecond))

	// Wipe.
	wipe := make([]animation.Animation, 0, len(all)+len(a))
	for _, m := range all {
		wipe = append(wipe, animation.FadeOut(m, 500*time.Millisecond))
	}
	for _, ar := range a {
		wipe = append(wipe, animation.FadeOut(ar, 500*time.Millisecond))
	}
	_, _ = s.Play(sink, animation.Parallel(wipe...))
}

// ---------- chapter 4 : math graphs ----------------------------------------

func chapterGraphs(s *scene.Scene, sink scene.FrameWriter) {
	s.Mobjects = s.Mobjects[:0]

	// Graphs need to read cleanly: switch the scene to Crisp + white
	// background for this chapter, restore Sketchy on exit.
	prevStyle := s.DefaultStyle
	prevBg := s.BgColor
	s.DefaultStyle = style.PresetCrisp
	s.BgColor = color.RGBA{0xFF, 0xFF, 0xFF, 0xFF}
	defer func() {
		s.DefaultStyle = prevStyle
		s.BgColor = prevBg
	}()

	// Heading keeps the handwritten font (set explicitly because the
	// scene default just flipped to Crisp/Sans for the graph itself).
	heading := mobject.NewText(380, "Plot a function. Shade a region.").MoveTo(0, 430)
	heading.SetStyle(style.Style{FontSize: style.FontHuge, FontFamily: style.FontHandDrawn})
	zero := 0.0
	hs := heading.Style()
	hs.Opacity = &zero

	// Coordinate plane — Crisp tokens give clean axes / tick labels.
	axes := mathx.NewAxes(-4, 4, -1.5, 4).
		WithSize(1200, 600).
		WithSteps(1, 1).
		WithGrid(true).
		WithLabels(true).
		MoveTo(0, -50)
	axes.SetReveal(0)

	// Plot a parabola and a sine curve.
	parabola := axes.Plot(func(x float64) float64 { return 0.5 * x * x }).
		WithRange(-3, 3).
		WithSamples(120).
		WithColor(color.RGBA{0xE0, 0x4A, 0x59, 0xFF}) // red
	parabola.SetReveal(0)

	sine := axes.Plot(func(x float64) float64 { return 2 * math.Sin(x) }).
		WithRange(-math.Pi, math.Pi).
		WithSamples(160).
		WithColor(color.RGBA{0x2D, 0x6A, 0xDF, 0xFF}) // blue
	sine.SetReveal(0)

	// Shade the area under the parabola from x = -2 to x = 2.
	shade := mathx.NewShade(parabola, -2, 2).WithSamples(80)
	shade.SetReveal(0)

	s.Add(heading, axes, shade, parabola, sine)

	_, _ = s.Play(sink, animation.FadeIn(heading, 400*time.Millisecond))
	_, _ = s.Play(sink, animation.DrawOn(axes, 1000*time.Millisecond))
	_, _ = s.Play(sink, animation.DrawOn(parabola, 1100*time.Millisecond))
	_, _ = s.Play(sink, animation.DrawOn(sine, 1100*time.Millisecond))
	_, _ = s.Play(sink, animation.DrawOn(shade, 700*time.Millisecond))
	_, _ = s.Play(sink, direction.Pause(1400*time.Millisecond))

	_, _ = s.Play(sink, animation.Parallel(
		animation.FadeOut(heading, 500*time.Millisecond),
		animation.FadeOut(axes, 500*time.Millisecond),
		animation.FadeOut(parabola, 500*time.Millisecond),
		animation.FadeOut(sine, 500*time.Millisecond),
		animation.FadeOut(shade, 500*time.Millisecond),
	))
}

// ---------- chapter 5 : math reveal ----------------------------------------

func chapterMath(s *scene.Scene, sink scene.FrameWriter) {
	s.Mobjects = s.Mobjects[:0]

	heading := mobject.NewText(400, "Pure-Go LaTeX. No TeX install.").MoveTo(0, 430)
	heading.SetStyle(style.Style{FontSize: style.FontHuge})
	zero := 0.0
	hs := heading.Style()
	hs.Opacity = &zero

	eq1 := mathx.NewEquation("E = mc^2").WithHeight(200).MoveTo(0, 100)
	eq2 := mathx.NewEquation("\\int_0^\\infty e^{-x^2}\\,dx = \\frac{\\sqrt{\\pi}}{2}").
		WithHeight(140).MoveTo(0, -180)

	s.Add(heading, eq1, eq2)

	_, _ = s.Play(sink, animation.FadeIn(heading, 400*time.Millisecond))
	_, _ = s.Play(sink, mathx.Write(eq1, 1500*time.Millisecond))
	_, _ = s.Play(sink, direction.Pause(600*time.Millisecond))
	_, _ = s.Play(sink, mathx.Write(eq2, 1800*time.Millisecond))
	_, _ = s.Play(sink, direction.Pause(1200*time.Millisecond))

	_, _ = s.Play(sink, animation.Parallel(
		animation.FadeOut(heading, 500*time.Millisecond),
		animation.FadeOut(eq1, 500*time.Millisecond),
		animation.FadeOut(eq2, 500*time.Millisecond),
	))
}

// ---------- chapter 5 : outro ----------------------------------------------

func chapterOutro(s *scene.Scene, sink scene.FrameWriter) {
	s.Mobjects = s.Mobjects[:0]

	headline := mobject.NewText(500, "Render your first video.").MoveTo(0, 300)
	headline.SetStyle(style.Style{FontSize: style.FontDisplay})

	cmd := mobject.NewText(501, "go get github.com/ankitsinghchadda/goanim").MoveTo(0, 0)
	cmd.SetStyle(style.Style{FontSize: style.FontHuge})

	star := mobject.NewText(502, "if you ship something with it — drop a ⭐").MoveTo(0, -200)
	star.SetStyle(style.Style{FontSize: style.FontXLarge})

	zero := 0.0
	for _, t := range []*mobject.Text{headline, cmd, star} {
		ts := t.Style()
		ts.Opacity = &zero
	}

	// Show a small example layout off to the side.
	mini := layout.NewHBox(
		icons.NewClient(601, "").MoveTo(0, 0),
		icons.NewServer(602, "").MoveTo(0, 0),
		icons.NewDatabase(603, "").MoveTo(0, 0),
	).WithSpacing(60).MoveTo(0, -280)

	s.Add(headline, cmd, star, mini)

	_, _ = s.Play(sink, animation.Sequence(
		animation.FadeIn(headline, 600*time.Millisecond),
		animation.FadeIn(cmd, 600*time.Millisecond),
		animation.FadeIn(star, 500*time.Millisecond),
	))
	_, _ = s.Play(sink, direction.Pause(2800*time.Millisecond))
}

// ---------- plumbing -------------------------------------------------------

type frameSink struct{ enc *render.VideoEncoder }

func (f frameSink) WriteFrame(img image.Image) error { return f.enc.WriteFrame(img) }

func must(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
