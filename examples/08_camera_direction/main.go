// Command direction_demo is the Phase-7 foundation integration test:
// a short video exercising Pause + Camera (ZoomTo, Focus, Reset) +
// LaserPointer in both crisp and sketchy. It's the gate-check for
// "these primitives compose" without taking on the full URL Shortener
// script.
//
// Topology: Client → Server → Database. The video:
//
//  1. fades the diagram in,
//  2. holds (Pause),
//  3. a laser pointer traces Client → Server → Database,
//  4. camera zooms to Database (with Focus dimming the others),
//  5. holds,
//  6. camera resets and dim releases,
//  7. final hold.
//
// Renders to out_direction_crisp.mp4 and out_direction_sketchy.mp4.
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
		out := "out_direction_" + v.name + ".mp4"
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
	cam := direction.NewCamera()
	s := scene.NewScene(W, H).
		WithRenderer(r).
		WithDefaultStyle(v.style).
		WithFont(style.FontHandDrawn, hand).
		WithFont(style.FontSans, sans).
		WithCamera(cam)
	s.BgColor = v.bg
	s.FPS = FPS

	client := icons.NewClient(1, "Client")
	server := icons.NewServer(2, "Server")
	db := icons.NewDatabase(3, "Database")
	row := layout.NewHBox(client, server, db).WithSpacing(160).MoveTo(0, 0)
	s.Add(row)
	_ = row.Bounds() // trigger layout

	a1 := systemdesign.NewArrow(101, client, server)
	a2 := systemdesign.NewArrow(102, server, db)
	s.Add(a1, a2)

	// Hide everything initially.
	for _, m := range []interface{ SetReveal(float64) }{client, server, db, a1, a2} {
		m.SetReveal(0)
	}

	enc, err := render.NewVideoEncoder(outPath, render.VideoOptions{
		Width: W, Height: H, FPS: FPS, CRF: 18, Preset: "fast",
	})
	if err != nil {
		return err
	}
	defer enc.Close()
	sink := frameSink{enc: enc}

	// 0.0–0.8s: stagger draw the diagram.
	if _, err := s.Play(sink, animation.Stagger(120*time.Millisecond,
		animation.DrawOn(asRev(client), 400*time.Millisecond),
		animation.DrawOn(asRev(server), 400*time.Millisecond),
		animation.DrawOn(asRev(db), 400*time.Millisecond),
		animation.DrawOn(a1, 400*time.Millisecond),
		animation.DrawOn(a2, 400*time.Millisecond),
	)); err != nil {
		return err
	}

	// 0.8–1.8s: hold.
	if _, err := s.Play(sink, direction.Pause(1*time.Second)); err != nil {
		return err
	}

	// 1.8–4.3s: laser pointer traces the data path.
	if _, err := s.Play(sink, direction.LaserPointer(
		direction.PathThrough(client, server, db),
		2500*time.Millisecond,
	)); err != nil {
		return err
	}

	// 4.3–5.3s: zoom in on the Database with Focus dimming the others.
	if _, err := s.Play(sink, cam.Focus(db, 1.8, 1*time.Second)); err != nil {
		return err
	}

	// 5.3–6.8s: hold the focused view.
	if _, err := s.Play(sink, direction.Pause(1500*time.Millisecond, "after-zoom-explanation")); err != nil {
		return err
	}

	// 6.8–7.8s: reset camera and release dim.
	if _, err := s.Play(sink, cam.UnFocus(1*time.Second)); err != nil {
		return err
	}

	// 7.8–8.8s: final hold.
	if _, err := s.PlayStill(sink, 1*time.Second); err != nil {
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

func asRev(m interface{}) animation.Revealer {
	if r, ok := m.(animation.Revealer); ok {
		return r
	}
	panic("direction_demo: not a Revealer")
}
