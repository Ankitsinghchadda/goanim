// Command icon_contact_sheet renders every generic icon under every
// sloppiness level, producing a single PNG grid for visual review.
//
// This is the design-iteration tool — re-run after touching any icon
// to confirm the change reads well in all three style modes.
package main

import (
	"fmt"
	"image/color"
	"os"

	"github.com/ankitsinghchadda/goanim/core/mobject"
	"github.com/ankitsinghchadda/goanim/core/render"
	"github.com/ankitsinghchadda/goanim/core/scene"
	"github.com/ankitsinghchadda/goanim/core/style"
	"github.com/ankitsinghchadda/goanim/mobjects/icons"
)

type iconFactory struct {
	name string
	make func(seed int64) mobject.Mobject
}

func factories() []iconFactory {
	return []iconFactory{
		// --- existing 15 ---
		{"client", func(s int64) mobject.Mobject { return icons.NewClient(s, "Client") }},
		{"server", func(s int64) mobject.Mobject { return icons.NewServer(s, "Server") }},
		{"database", func(s int64) mobject.Mobject { return icons.NewDatabase(s, "Database") }},
		{"queue", func(s int64) mobject.Mobject { return icons.NewQueue(s, "Queue") }},
		{"stack", func(s int64) mobject.Mobject { return icons.NewStack(s, "Stack") }},
		{"cache", func(s int64) mobject.Mobject { return icons.NewCache(s, "Cache") }},
		{"loadbalancer", func(s int64) mobject.Mobject { return icons.NewLoadBalancer(s, "LB") }},
		{"broker", func(s int64) mobject.Mobject { return icons.NewMessageBroker(s, "Broker") }},
		{"worker", func(s int64) mobject.Mobject { return icons.NewWorker(s, "Worker") }},
		{"apigateway", func(s int64) mobject.Mobject { return icons.NewAPIGateway(s, "Gateway") }},
		{"cdn", func(s int64) mobject.Mobject { return icons.NewCDN(s, "CDN") }},
		{"user", func(s int64) mobject.Mobject { return icons.NewUser(s, "User") }},
		{"service", func(s int64) mobject.Mobject { return icons.NewService(s, "Service") }},
		{"function", func(s int64) mobject.Mobject { return icons.NewFunction(s, "λ") }},
		{"storage", func(s int64) mobject.Mobject { return icons.NewStorage(s, "Bucket") }},
		// --- Batch 1: compute variants ---
		{"container", func(s int64) mobject.Mobject { return icons.NewContainer(s, "Container") }},
		{"pod", func(s int64) mobject.Mobject { return icons.NewPod(s, "Pod") }},
		{"cluster", func(s int64) mobject.Mobject { return icons.NewCluster(s, "Cluster") }},
		{"vm", func(s int64) mobject.Mobject { return icons.NewVM(s, "VM") }},
		{"edgefunc", func(s int64) mobject.Mobject { return icons.NewEdgeFunction(s, "Edge λ") }},
		// --- Batch 2: storage variants ---
		{"relationaldb", func(s int64) mobject.Mobject { return icons.NewRelationalDB(s, "SQL") }},
		{"nosqldb", func(s int64) mobject.Mobject { return icons.NewNoSQLDB(s, "NoSQL") }},
		{"kvstore", func(s int64) mobject.Mobject { return icons.NewKeyValueStore(s, "K:V") }},
		{"objectstorage", func(s int64) mobject.Mobject { return icons.NewObjectStorage(s, "Objects") }},
		{"blockstorage", func(s int64) mobject.Mobject { return icons.NewBlockStorage(s, "Blocks") }},
		// --- Batch 3: database specialty ---
		{"warehouse", func(s int64) mobject.Mobject { return icons.NewDataWarehouse(s, "Warehouse") }},
		{"searchindex", func(s int64) mobject.Mobject { return icons.NewSearchIndex(s, "Search") }},
		{"timeseries", func(s int64) mobject.Mobject { return icons.NewTimeSeriesDB(s, "TSDB") }},
		{"graphdb", func(s int64) mobject.Mobject { return icons.NewGraphDB(s, "Graph") }},
		// --- Batch 4: networking + messaging ---
		{"firewall", func(s int64) mobject.Mobject { return icons.NewFirewall(s, "Firewall") }},
		{"revproxy", func(s int64) mobject.Mobject { return icons.NewReverseProxy(s, "RevProxy") }},
		{"dns", func(s int64) mobject.Mobject { return icons.NewDNS(s, "DNS") }},
		{"eventstream", func(s int64) mobject.Mobject { return icons.NewEventStream(s, "Stream") }},
		{"pubsub", func(s int64) mobject.Mobject { return icons.NewPubSubTopic(s, "Topic") }},
		// --- Batch 5: observability + IoT + specialty ---
		{"metrics", func(s int64) mobject.Mobject { return icons.NewMetrics(s, "Metrics") }},
		{"logs", func(s int64) mobject.Mobject { return icons.NewLogs(s, "Logs") }},
		{"tracing", func(s int64) mobject.Mobject { return icons.NewTracing(s, "Tracing") }},
		{"iotdevice", func(s int64) mobject.Mobject { return icons.NewIoTDevice(s, "IoT") }},
		{"mobile", func(s int64) mobject.Mobject { return icons.NewMobileClient(s, "Mobile") }},
	}
}

// presetForRow picks the style for each contact-sheet row.
func presetForRow(row int) (style.Style, color.Color) {
	switch row {
	case 0:
		// Architect-sloppiness (crisp) on white.
		return style.PresetCrisp, color.RGBA{0xFF, 0xFF, 0xFF, 0xFF}
	case 1:
		// Artist sloppiness (Excalidraw default) on white.
		return style.PresetExcalidraw, color.RGBA{0xFF, 0xFF, 0xFF, 0xFF}
	case 2:
		// Cartoonist sloppiness (sketchy) on cream.
		return style.PresetSketchy, color.RGBA{0xFF, 0xF8, 0xE1, 0xFF}
	}
	return style.PresetExcalidraw, color.RGBA{0xFF, 0xFF, 0xFF, 0xFF}
}

func main() {
	hand, err := render.Excalifont()
	must(err, "excalifont")
	sans, err := render.Inter()
	must(err, "inter")

	const (
		cols   = 5
		gutter = 30
	)
	facts := factories()
	rows := 3
	// Each icon needs ~220x200 (body + label). With 5 cols, total width
	// = 5*220 + 4*gutter = 1100 + 120 = 1220.  With 3 sloppiness rows
	// per icon row, total rows = ceil(15/5) * 3 = 9 rows. We'll lay out
	// in a tall image.
	rowsPerIcon := ceilDiv(len(facts), cols)
	totalRows := rowsPerIcon * rows
	imgW := cols*240 + (cols-1)*gutter + 80
	imgH := totalRows*240 + (totalRows-1)*gutter + 120
	if imgW < 1920 {
		imgW = 1920
	}

	r := render.NewCanvasRenderer(render.Options{Supersample: 2, DefaultFont: hand})
	r.BeginFrame(imgW, imgH, color.RGBA{0xFF, 0xFF, 0xFF, 0xFF})

	// We can't use a single Scene because each cell needs a different
	// style. Instead, render directly through the renderer with three
	// passes (one per style). Each pass uses its own context.
	for row := 0; row < rows; row++ {
		preset, _ := presetForRow(row)
		ctx := style.Context{
			SceneDefault: preset,
			Fonts: map[style.FontFamily]render.FontFace{
				style.FontHandDrawn: hand,
				style.FontSans:      sans,
			},
		}
		for i, f := range facts {
			gridRow := i / cols
			gridCol := i % cols
			// Cell center in user-space (center-origin).
			cellW := float64(240 + gutter)
			cellH := float64(240 + gutter)
			x := -float64(imgW)/2 + float64(gutter) + float64(gridCol)*cellW + cellW/2
			y := +float64(imgH)/2 - float64(gutter) - float64(row*rowsPerIcon+gridRow)*cellH - cellH/2
			ic := f.make(int64(row*100 + i))
			if p, ok := ic.(interface{ SetPosition(float64, float64) }); ok {
				p.SetPosition(x, y)
			}
			ic.Render(r, ctx)
		}
	}

	// Final headers across the top of each block.
	headerCtx := style.Context{
		SceneDefault: style.PresetCrisp,
		Fonts: map[style.FontFamily]render.FontFace{
			style.FontHandDrawn: hand,
			style.FontSans:      sans,
		},
	}
	cellH := float64(240 + gutter)
	for row := 0; row < rows; row++ {
		preset, _ := presetForRow(row)
		var label string
		switch row {
		case 0:
			label = "Architect (crisp)"
		case 1:
			label = "Artist (Excalidraw)"
		case 2:
			label = "Cartoonist (sketchy)"
		}
		_ = preset
		yTop := +float64(imgH)/2 - 40 - float64(row*ceilDiv(len(facts), cols))*cellH
		t := mobject.NewText(int64(9000+row), label).MoveTo(-float64(imgW)/2+80, yTop+40)
		t.SetStyle(style.Style{FontSize: style.FontLarge, FontFamily: style.FontSans})
		_ = scene.NewScene(imgW, imgH) // (unused, but kept for symmetry)
		t.Render(r, headerCtx)
	}

	out, err := os.Create("icon_contact_sheet.png")
	must(err, "create contact sheet")
	defer out.Close()
	must(r.EncodePNG(out), "encode png")

	fmt.Printf("wrote icon_contact_sheet.png (%dx%d, %d icons × 3 styles)\n",
		imgW, imgH, len(facts))
}

func ceilDiv(a, b int) int { return (a + b - 1) / b }

func must(err error, msg string) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", msg, err)
		os.Exit(1)
	}
}
