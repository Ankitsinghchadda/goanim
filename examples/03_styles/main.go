// Command preset_gallery renders the same Client → Server → Database
// diagram in each of goanim's named style presets, producing one PNG
// per preset. Useful for visual regression and for the README.
package main

import (
	"fmt"
	"image/color"
	"os"

	"github.com/ankitsinghchadda/goanim/core/mobject"
	"github.com/ankitsinghchadda/goanim/core/render"
	"github.com/ankitsinghchadda/goanim/core/scene"
	"github.com/ankitsinghchadda/goanim/core/style"
	"github.com/ankitsinghchadda/goanim/mobjects/systemdesign"
)

type preset struct {
	name  string
	style style.Style
	bg    color.Color
}

func main() {
	hand, err := render.Excalifont()
	must(err, "excalifont")
	sans, err := render.Inter()
	must(err, "inter")

	white := color.RGBA{0xFF, 0xFF, 0xFF, 0xFF}
	cream := color.RGBA{0xFD, 0xF6, 0xE3, 0xFF}
	blueprintBG := color.RGBA{0x1A, 0x4E, 0x8C, 0xFF}

	presets := []preset{
		{"excalidraw", style.PresetExcalidraw, white},
		{"sketchy", style.PresetSketchy, cream},
		{"crisp", style.PresetCrisp, white},
		{"blueprint", style.PresetBlueprint, blueprintBG},
		{"notebook", style.PresetNotebook, white},
	}

	for _, p := range presets {
		path := "out_" + p.name + ".png"
		if err := renderOne(p, hand, sans, path); err != nil {
			fmt.Fprintln(os.Stderr, p.name, "failed:", err)
			os.Exit(1)
		}
		fmt.Println("wrote", path)
	}
}

func renderOne(p preset, hand, sans render.FontFace, outPath string) error {
	r := render.NewCanvasRenderer(render.Options{Supersample: 2, DefaultFont: hand})
	r.BeginFrame(1920, 1080, p.bg)

	s := scene.NewScene(1920, 1080).
		WithRenderer(r).
		WithDefaultStyle(p.style).
		WithFont(style.FontHandDrawn, hand).
		WithFont(style.FontSans, sans)

	client := systemdesign.NewClient(1001, "Client").MoveTo(-600, 40)
	server := systemdesign.NewServer(1002, "Server").MoveTo(0, 40)
	database := systemdesign.NewDatabase(1003, "Database").MoveTo(620, 40)
	arrow1 := systemdesign.NewArrow(2001, client, server)
	arrow2 := systemdesign.NewArrow(2002, server, database)

	s.Add(client, server, database, arrow1, arrow2)

	caption := mobject.NewText(0, p.name).MoveTo(0, -360)
	caption.SetStyle(style.Style{FontSize: style.FontLarge})
	s.Add(caption)

	s.RenderFrame()

	f, err := os.Create(outPath)
	if err != nil {
		return err
	}
	defer f.Close()
	return r.EncodePNG(f)
}

func must(err error, msg string) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", msg, err)
		os.Exit(1)
	}
}
