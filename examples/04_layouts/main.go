// Command layout_demo renders a static PNG showing nested layout
// containers. Verifies HBox/VBox/Grid math works without any manual
// coordinate calculation.
package main

import (
	"fmt"
	"image/color"
	"os"

	"github.com/ankitsinghchadda/goanim/core/layout"
	"github.com/ankitsinghchadda/goanim/core/render"
	"github.com/ankitsinghchadda/goanim/core/scene"
	"github.com/ankitsinghchadda/goanim/core/style"
	"github.com/ankitsinghchadda/goanim/mobjects/systemdesign"
)

func main() {
	hand, err := render.Excalifont()
	must(err, "excalifont")
	sans, err := render.Inter()
	must(err, "inter")

	r := render.NewCanvasRenderer(render.Options{Supersample: 2, DefaultFont: hand})
	r.BeginFrame(1920, 1080, color.RGBA{0xFF, 0xFF, 0xFF, 0xFF})

	s := scene.NewScene(1920, 1080).
		WithRenderer(r).
		WithDefaultStyle(style.PresetCrisp).
		WithFont(style.FontHandDrawn, hand).
		WithFont(style.FontSans, sans)

	// Top row: client → load balancer (2 nodes).
	client := systemdesign.NewClient(1, "Client")
	lb := systemdesign.NewServer(2, "LoadBalancer")
	row1 := layout.NewHBox(client, lb).WithSpacing(120)

	// Middle row: three server instances.
	s1 := systemdesign.NewServer(3, "API-1")
	s2 := systemdesign.NewServer(4, "API-2")
	s3 := systemdesign.NewServer(5, "API-3")
	row2 := layout.NewHBox(s1, s2, s3).WithSpacing(60)

	// Bottom row: cache and database.
	cache := systemdesign.NewServer(6, "Cache")
	db := systemdesign.NewDatabase(7, "Postgres")
	row3 := layout.NewHBox(cache, db).WithSpacing(100)

	// Compose into a VBox.
	column := layout.NewVBox(row1, row2, row3).WithSpacing(80).MoveTo(0, 0)
	s.Add(column)

	s.RenderFrame()

	out, err := os.Create("out.png")
	must(err, "create out.png")
	defer out.Close()
	must(r.EncodePNG(out), "encode png")

	fmt.Println("wrote out.png (layout demo)")
}

func must(err error, msg string) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", msg, err)
		os.Exit(1)
	}
}
