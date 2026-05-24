// Command math_demo renders the Phase-4 math deliverable: equations,
// number lines, and a coordinate plane with a plotted function.
//
// The video is rendered in sketchy mode by default — to showcase the
// "handwritten math" capability (style-aware LaTeX rendering) that
// makes goanim distinctive. Set VARIANT=crisp for the clean version.
package main

import (
	"fmt"
	"image"
	"image/color"
	"os"
	"time"

	"github.com/ankitsinghchadda/goanim/core/animation"
	"github.com/ankitsinghchadda/goanim/core/mobject"
	"github.com/ankitsinghchadda/goanim/core/render"
	"github.com/ankitsinghchadda/goanim/core/scene"
	"github.com/ankitsinghchadda/goanim/core/style"
	"github.com/ankitsinghchadda/goanim/mobjects/mathx"
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
		out := "out_math_" + v.name + ".mp4"
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
		{"sketchy", style.PresetExcalidraw, color.RGBA{0xFF, 0xFB, 0xEB, 0xFF}},
		{"crisp", style.PresetCrisp, color.RGBA{0xFF, 0xFF, 0xFF, 0xFF}},
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

	enc, err := render.NewVideoEncoder(outPath, render.VideoOptions{
		Width: W, Height: H, FPS: FPS, CRF: 18, Preset: "slow",
	})
	if err != nil {
		return err
	}
	defer enc.Close()
	sink := frameSink{enc: enc}

	// === Scene 1: Equation writing + highlight =============================
	title := mobject.NewText(100, "Energy-Mass Equivalence").MoveTo(0, 380)
	title.SetStyle(style.Style{FontSize: style.FontLarge})
	s.Add(title)

	eq := mathx.NewEquation("E = mc^2").WithHeight(150).MoveTo(0, 0)
	eq.SetReveal(0)
	s.Add(eq)

	// FadeIn the title.
	if _, err := s.Play(sink, animation.FadeIn(title, 500*time.Millisecond)); err != nil {
		return err
	}

	// Write the equation (FadeIn for now; per-symbol stagger arrives in
	// the Write animation below).
	if _, err := s.Play(sink, mathx.Write(eq, 1500*time.Millisecond)); err != nil {
		return err
	}

	// Hold briefly so viewers can absorb.
	if _, err := s.PlayStill(sink, 700*time.Millisecond); err != nil {
		return err
	}

	// === Scene 2: Number line ==============================================
	if _, err := s.Play(sink, animation.FadeOut(eq, 400*time.Millisecond)); err != nil {
		return err
	}
	s.Remove(eq)

	if _, err := s.Play(sink, animation.FadeOut(title, 400*time.Millisecond)); err != nil {
		return err
	}
	s.Remove(title)

	title2 := mobject.NewText(101, "The Number Line").MoveTo(0, 380)
	title2.SetStyle(style.Style{FontSize: style.FontLarge})
	s.Add(title2)
	if _, err := s.Play(sink, animation.FadeIn(title2, 400*time.Millisecond)); err != nil {
		return err
	}

	nl := mathx.NewNumberLine(-5, 5).WithLength(1100).WithStep(1).MoveTo(0, 60)
	nl.SetReveal(0)
	s.Add(nl)
	if _, err := s.Play(sink, animation.DrawOn(nl, 1000*time.Millisecond)); err != nil {
		return err
	}

	dist := mathx.NewEquation(`|3 - (-2)| = 5`).WithHeight(70).MoveTo(0, -260)
	dist.SetReveal(0)
	s.Add(dist)
	if _, err := s.Play(sink, mathx.Write(dist, 1100*time.Millisecond)); err != nil {
		return err
	}

	// === Scene 3: Axes with plot ==========================================
	if _, err := s.Play(sink, animation.FadeOut(nl, 350*time.Millisecond)); err != nil {
		return err
	}
	s.Remove(nl)
	if _, err := s.Play(sink, animation.FadeOut(dist, 350*time.Millisecond)); err != nil {
		return err
	}
	s.Remove(dist)
	if _, err := s.Play(sink, animation.FadeOut(title2, 350*time.Millisecond)); err != nil {
		return err
	}
	s.Remove(title2)

	title3 := mobject.NewText(102, "A Function on the Plane").MoveTo(0, 440)
	title3.SetStyle(style.Style{FontSize: style.FontLarge})
	s.Add(title3)
	if _, err := s.Play(sink, animation.FadeIn(title3, 400*time.Millisecond)); err != nil {
		return err
	}

	axes := mathx.NewAxes(-4, 4, -4, 9).WithSize(820, 560).WithSteps(1, 2).WithGrid(true).MoveTo(0, -20)
	axes.SetReveal(0)
	s.Add(axes)
	if _, err := s.Play(sink, animation.DrawOn(axes, 800*time.Millisecond)); err != nil {
		return err
	}

	graph := axes.Plot(func(x float64) float64 { return x*x - 3 }).WithRange(-3.5, 3.5).WithSamples(400)
	graph.SetReveal(0)
	s.Add(graph)
	if _, err := s.Play(sink, animation.DrawOn(graph, 1200*time.Millisecond)); err != nil {
		return err
	}

	formula := mathx.NewEquation(`f(x) = x^2 - 3`).WithHeight(56).MoveTo(0, -430)
	formula.SetReveal(0)
	s.Add(formula)
	if _, err := s.Play(sink, mathx.Write(formula, 900*time.Millisecond)); err != nil {
		return err
	}

	// Hold final.
	if _, err := s.PlayStill(sink, 1500*time.Millisecond); err != nil {
		return err
	}
	return nil
}

type frameSink struct{ enc *render.VideoEncoder }

func (f frameSink) WriteFrame(img image.Image) error { return f.enc.WriteFrame(img) }

func must(err error, msg string) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", msg, err)
		os.Exit(1)
	}
}
