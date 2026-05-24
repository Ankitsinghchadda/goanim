// Command animations is a guided tour of goanim's core animation
// primitives. The output video labels each beat on screen so you can
// see what each primitive does.
//
// Beats:
//
//  1. DrawOn        — reveal a mobject as if hand-drawn
//  2. FadeIn        — fade opacity 0 → 1
//  3. MoveTo        — translate to a new position
//  4. PopIn         — scale 0 → 1 with overshoot
//  5. Flash         — briefly tint a mobject
//  6. Sequence      — run animations back-to-back
//  7. Parallel      — run animations simultaneously
//  8. Stagger       — start each child after a delay
//
// Run:
//
//	go run ./examples/02_animations
//
// Output: out_animations.mp4 (1920×1080, 60fps, ~22 s)
package main

import (
	"fmt"
	"image"
	"image/color"
	"os"
	"time"

	"github.com/ankitsinghchadda/goanim/core/animation"
	"github.com/ankitsinghchadda/goanim/core/direction"
	"github.com/ankitsinghchadda/goanim/core/mobject"
	"github.com/ankitsinghchadda/goanim/core/render"
	"github.com/ankitsinghchadda/goanim/core/scene"
	"github.com/ankitsinghchadda/goanim/core/style"
	"github.com/ankitsinghchadda/goanim/mobjects/icons"
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
	s := scene.NewScene(W, H).
		WithRenderer(r).
		WithDefaultStyle(style.PresetCrisp).
		WithFont(style.FontHandDrawn, hand).
		WithFont(style.FontSans, sans)
	s.BgColor = color.RGBA{0xFF, 0xFF, 0xFF, 0xFF}
	s.FPS = FPS

	enc, err := render.NewVideoEncoder("out_animations.mp4", render.VideoOptions{
		Width: W, Height: H, FPS: FPS, CRF: 20, Preset: "fast",
	})
	must(err)
	defer enc.Close()
	sink := frameSink{enc: enc}

	// Each beat owns its own heading (Text content is immutable after
	// construction, so we make a fresh heading per beat).
	beats := []struct {
		title string
		fn    func(*scene.Scene, scene.FrameWriter)
	}{
		{"DrawOn — reveal as if hand-drawn", beatDrawOn},
		{"FadeIn — opacity 0 → 1", beatFadeIn},
		{"MoveTo — translate", beatMoveTo},
		{"PopIn — scale 0 → 1 with overshoot", beatPopIn},
		{"Flash — briefly tint", beatFlash},
		{"Sequence — back-to-back", beatSequence},
		{"Parallel — simultaneous", beatParallel},
		{"Stagger — delayed start per child", beatStagger},
	}
	for i, beat := range beats {
		s.Mobjects = s.Mobjects[:0]
		heading := mobject.NewText(int64(1000+i), beat.title).MoveTo(0, 380)
		heading.SetStyle(style.Style{FontSize: style.FontXLarge, FontFamily: style.FontSans})
		s.Add(heading)
		_, _ = s.Play(sink, animation.FadeIn(heading, 250*time.Millisecond))
		beat.fn(s, sink)
		_, _ = s.Play(sink, direction.Pause(700*time.Millisecond))
	}

	fmt.Println("wrote out_animations.mp4")
}

func beatDrawOn(s *scene.Scene, sink scene.FrameWriter) {
	box := icons.NewServer(1, "Server").MoveTo(0, 0)
	box.SetReveal(0)
	s.Add(box)
	_, _ = s.Play(sink, animation.DrawOn(box, 1200*time.Millisecond))
}

func beatFadeIn(s *scene.Scene, sink scene.FrameWriter) {
	zero := 0.0
	box := icons.NewDatabase(1, "Database").MoveTo(0, 0)
	box.SetStyle(style.Style{Opacity: &zero})
	s.Add(box)
	_, _ = s.Play(sink, animation.FadeIn(box, 800*time.Millisecond))
}

func beatMoveTo(s *scene.Scene, sink scene.FrameWriter) {
	box := icons.NewClient(1, "Client").MoveTo(-600, 0)
	s.Add(box)
	_, _ = s.Play(sink, animation.MoveTo(box, 600, 0, 1000*time.Millisecond))
}

func beatPopIn(s *scene.Scene, sink scene.FrameWriter) {
	box := icons.NewCache(1, "Cache").MoveTo(0, 0)
	box.SetVisualScale(0)
	s.Add(box)
	_, _ = s.Play(sink, animation.PopIn(box, 700*time.Millisecond))
}

func beatFlash(s *scene.Scene, sink scene.FrameWriter) {
	box := icons.NewLoadBalancer(1, "LB").MoveTo(0, 0)
	s.Add(box)
	_, _ = s.Play(sink, animation.Flash(box,
		color.RGBA{0xF5, 0x9E, 0x0B, 0xFF}, 600*time.Millisecond))
}

func beatSequence(s *scene.Scene, sink scene.FrameWriter) {
	a := icons.NewClient(1, "A").MoveTo(-500, 0)
	b := icons.NewServer(2, "B").MoveTo(0, 0)
	c := icons.NewDatabase(3, "C").MoveTo(500, 0)
	for _, m := range []interface{ SetReveal(float64) }{a, b, c} {
		m.SetReveal(0)
	}
	s.Add(a, b, c)
	_, _ = s.Play(sink, animation.Sequence(
		animation.DrawOn(a, 400*time.Millisecond),
		animation.DrawOn(b, 400*time.Millisecond),
		animation.DrawOn(c, 400*time.Millisecond),
	))
}

func beatParallel(s *scene.Scene, sink scene.FrameWriter) {
	a := icons.NewClient(1, "A").MoveTo(-500, 0)
	b := icons.NewServer(2, "B").MoveTo(0, 0)
	c := icons.NewDatabase(3, "C").MoveTo(500, 0)
	for _, m := range []interface{ SetReveal(float64) }{a, b, c} {
		m.SetReveal(0)
	}
	s.Add(a, b, c)
	_, _ = s.Play(sink, animation.Parallel(
		animation.DrawOn(a, 600*time.Millisecond),
		animation.DrawOn(b, 600*time.Millisecond),
		animation.DrawOn(c, 600*time.Millisecond),
	))
}

func beatStagger(s *scene.Scene, sink scene.FrameWriter) {
	xs := []float64{-600, -300, 0, 300, 600}
	mobs := make([]mobject.Mobject, len(xs))
	revealers := make([]animation.Animation, len(xs))
	for i, x := range xs {
		m := icons.NewServer(int64(i+1), fmt.Sprintf("S%d", i+1)).MoveTo(x, 0)
		m.SetReveal(0)
		mobs[i] = m
		revealers[i] = animation.DrawOn(m, 500*time.Millisecond)
	}
	s.Add(mobs...)
	_, _ = s.Play(sink, animation.Stagger(150*time.Millisecond, revealers...))
}

// frameSink adapts a *VideoEncoder to scene.FrameWriter.
type frameSink struct{ enc *render.VideoEncoder }

func (f frameSink) WriteFrame(img image.Image) error { return f.enc.WriteFrame(img) }

func must(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
