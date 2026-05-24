// Command hello_world is the smallest possible goanim program.
// It draws a Client → Server diagram and writes it to a PNG.
//
// Run it:
//
//	go run ./examples/01_hello_world
//
// Output: out_hello_world.png in the current directory.
package main

import (
	"fmt"
	"image/color"
	"os"

	"github.com/ankitsinghchadda/goanim/core/render"
	"github.com/ankitsinghchadda/goanim/core/scene"
	"github.com/ankitsinghchadda/goanim/core/style"
	"github.com/ankitsinghchadda/goanim/mobjects/icons"
	"github.com/ankitsinghchadda/goanim/mobjects/systemdesign"
)

func main() {
	// 1. Load the bundled fonts.
	hand, err := render.Excalifont()
	must(err)
	sans, err := render.Inter()
	must(err)

	// 2. Create a renderer + scene. The scene defines the canvas size,
	//    the active style preset, and the fonts.
	const W, H = 1920, 1080
	r := render.NewCanvasRenderer(render.Options{Supersample: 2, DefaultFont: hand})
	r.BeginFrame(W, H, color.RGBA{0xFD, 0xF6, 0xE3, 0xFF})

	s := scene.NewScene(W, H).
		WithRenderer(r).
		WithDefaultStyle(style.PresetSketchy).
		WithFont(style.FontHandDrawn, hand).
		WithFont(style.FontSans, sans)

	// 3. Add some icons and an arrow between them. Coordinates are
	//    center-origin (0,0 is the canvas center; +Y is up).
	client := icons.NewClient(1, "Client").MoveTo(-400, 0)
	server := icons.NewServer(2, "Server").MoveTo(400, 0)
	arrow := systemdesign.NewArrow(3, client, server).WithLabel("request")
	s.Add(client, server, arrow)

	// 4. Render a single frame and write it to disk.
	s.RenderFrame()
	out, err := os.Create("out_hello_world.png")
	must(err)
	defer out.Close()
	must(r.EncodePNG(out))

	fmt.Println("wrote out_hello_world.png")
}

func must(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
