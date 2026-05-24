// Command bench is the Phase-8 benchmark runner. It builds each of
// the scenes from cmd/bench/scenes, plays them through the real
// ffmpeg encoder (output goes to a discardable /tmp path), and
// records wall time + allocations + peak RSS. Each scene runs N
// times (default 3) and the report keeps the best-of-N (which most
// closely reflects the library's actual throughput; the other runs
// have noise from OS background activity).
//
// Output is a JSON file (bench_baseline.json by default — change
// with --output) and a Markdown report on stdout that highlights
// regressions vs an existing baseline.json if one exists.
//
// Usage:
//
//	go run ./cmd/bench --runs 3 --output bench_baseline.json
//	go run ./cmd/bench --compare bench_baseline.json   # quick compare
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"image"
	"os"
	"runtime"
	"sort"
	"time"

	"github.com/ankitsinghchadda/goanim/cmd/bench/scenes"
	"github.com/ankitsinghchadda/goanim/core/render"
	"github.com/ankitsinghchadda/goanim/core/scene"
)

// Result is one (scene, style) measurement. Embedded in the JSON
// report and also written as a row in the Markdown table.
type Result struct {
	Scene        string  `json:"scene"`
	Style        string  `json:"style"`
	WallTimeMs   int64   `json:"wall_time_ms"`
	Frames       int     `json:"frames"`
	FPS          float64 `json:"fps"`
	AllocBytes   uint64  `json:"alloc_bytes"`
	NumAllocs    uint64  `json:"num_allocs"`
	PeakRSSBytes uint64  `json:"peak_rss_bytes"`
}

// Report is the JSON shape committed as bench_baseline.json.
type Report struct {
	Generated time.Time `json:"generated"`
	Runs      int       `json:"runs_per_scene"`
	Results   []Result  `json:"results"`
	Notes     string    `json:"notes,omitempty"`
}

func main() {
	output := flag.String("output", "bench_report.json", "Where to write the JSON report. Use bench_baseline.json when establishing baseline.")
	compareTo := flag.String("compare", "", "Path to an existing report to compare against. Defaults to none.")
	runs := flag.Int("runs", 3, "Runs per scene (best-of-N is reported)")
	only := flag.String("only", "", "Optional filter: only run scenes whose name contains this substring (e.g. 'small' or 'sketchy')")
	notes := flag.String("notes", "", "Free-text notes recorded in the report (e.g. 'after-opt-1')")
	flag.Parse()

	hand, err := render.Excalifont()
	must(err)
	sans, err := render.Inter()
	must(err)

	specs := scenes.All()
	results := make([]Result, 0, len(specs))

	for _, sp := range specs {
		if *only != "" && !contains(sp.Name+"-"+string(sp.Style), *only) {
			continue
		}
		fmt.Printf("=== %s / %s ===\n", sp.Name, sp.Style)
		best := runScene(sp, hand, sans, *runs)
		results = append(results, best)
		fmt.Printf("  wall=%v  frames=%d  fps=%.1f  allocs=%d  rss=%s\n",
			time.Duration(best.WallTimeMs)*time.Millisecond,
			best.Frames, best.FPS, best.NumAllocs, humanBytes(best.PeakRSSBytes))
	}

	rep := Report{
		Generated: time.Now(),
		Runs:      *runs,
		Results:   results,
		Notes:     *notes,
	}
	mustWriteJSON(*output, rep)
	fmt.Printf("wrote %s\n", *output)

	if *compareTo != "" {
		base, err := readReport(*compareTo)
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: could not read --compare baseline %q: %v\n", *compareTo, err)
		} else {
			fmt.Println()
			printMarkdownDelta(base, rep)
		}
	} else {
		fmt.Println()
		printMarkdownSummary(rep)
	}
}

// runScene runs a scene N times and returns the best-of-N. "Best" is
// minimum wall time — allocations and RSS tend to vary less, but we
// take them from the same run that produced the best wall time so
// the report is internally consistent.
func runScene(sp scenes.SceneSpec, hand, sans render.FontFace, runs int) Result {
	var all []Result
	for i := 0; i < runs; i++ {
		runtime.GC()
		var before runtime.MemStats
		runtime.ReadMemStats(&before)
		rssBefore := peakRSSBytes()

		s, anim, err := sp.Build(hand, sans)
		if err != nil {
			fmt.Fprintf(os.Stderr, "build %s/%s: %v\n", sp.Name, sp.Style, err)
			continue
		}

		outPath := fmt.Sprintf("/tmp/goanim-bench-%s-%s.mp4", sp.Name, sp.Style)
		enc, err := render.NewVideoEncoder(outPath, render.VideoOptions{
			Width: 1920, Height: 1080, FPS: 60, CRF: 23, Preset: "ultrafast",
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "encoder %s/%s: %v\n", sp.Name, sp.Style, err)
			continue
		}
		sink := encSink{enc: enc}

		start := time.Now()
		nFrames, err := s.Play(sink, anim)
		_ = enc.Close()
		elapsed := time.Since(start)
		if err != nil {
			fmt.Fprintf(os.Stderr, "play %s/%s: %v\n", sp.Name, sp.Style, err)
			continue
		}

		var after runtime.MemStats
		runtime.ReadMemStats(&after)
		rssAfter := peakRSSBytes()
		peak := rssAfter
		if rssBefore > peak {
			peak = rssBefore
		}

		fps := 0.0
		if elapsed > 0 {
			fps = float64(nFrames) / elapsed.Seconds()
		}
		all = append(all, Result{
			Scene:        sp.Name,
			Style:        string(sp.Style),
			WallTimeMs:   elapsed.Milliseconds(),
			Frames:       nFrames,
			FPS:          fps,
			AllocBytes:   after.TotalAlloc - before.TotalAlloc,
			NumAllocs:    after.Mallocs - before.Mallocs,
			PeakRSSBytes: peak,
		})
	}
	if len(all) == 0 {
		return Result{Scene: sp.Name, Style: string(sp.Style)}
	}
	sort.Slice(all, func(i, j int) bool { return all[i].WallTimeMs < all[j].WallTimeMs })
	return all[0]
}

// encSink adapts the VideoEncoder to scene.FrameWriter.
type encSink struct{ enc *render.VideoEncoder }

func (e encSink) WriteFrame(img image.Image) error { return e.enc.WriteFrame(img) }

// reading / writing reports ---------------------------------------------------

func mustWriteJSON(path string, rep Report) {
	f, err := os.Create(path)
	must(err)
	defer f.Close()
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	must(enc.Encode(rep))
}

func readReport(path string) (Report, error) {
	var rep Report
	f, err := os.Open(path)
	if err != nil {
		return rep, err
	}
	defer f.Close()
	return rep, json.NewDecoder(f).Decode(&rep)
}

// printMarkdownSummary prints a single-table summary of the current
// report (no comparison) to stdout. Suitable for embedding in a
// benchmark write-up.
func printMarkdownSummary(rep Report) {
	fmt.Println("| scene | style | wall | frames | fps | allocs | peak rss |")
	fmt.Println("|---|---|---|---|---|---|---|")
	for _, r := range rep.Results {
		fmt.Printf("| %s | %s | %v | %d | %.1f | %d | %s |\n",
			r.Scene, r.Style,
			time.Duration(r.WallTimeMs)*time.Millisecond,
			r.Frames, r.FPS, r.NumAllocs, humanBytes(r.PeakRSSBytes))
	}
}

// printMarkdownDelta prints a side-by-side comparison between the
// current report and a previously-committed baseline.
func printMarkdownDelta(baseline, current Report) {
	idx := func(rep Report) map[string]Result {
		m := make(map[string]Result, len(rep.Results))
		for _, r := range rep.Results {
			m[r.Scene+"-"+r.Style] = r
		}
		return m
	}
	bi := idx(baseline)

	fmt.Println("| scene | style | wall before → after | Δ% | allocs before → after | Δ% |")
	fmt.Println("|---|---|---|---|---|---|")
	for _, c := range current.Results {
		b, ok := bi[c.Scene+"-"+c.Style]
		if !ok {
			fmt.Printf("| %s | %s | (new) %dms | — | %d | — |\n",
				c.Scene, c.Style, c.WallTimeMs, c.NumAllocs)
			continue
		}
		dW := pctDelta(b.WallTimeMs, c.WallTimeMs)
		dA := pctDelta(int64(b.NumAllocs), int64(c.NumAllocs))
		fmt.Printf("| %s | %s | %dms → %dms | %s | %d → %d | %s |\n",
			c.Scene, c.Style, b.WallTimeMs, c.WallTimeMs, dW,
			b.NumAllocs, c.NumAllocs, dA)
	}
}

func pctDelta(before, after int64) string {
	if before == 0 {
		return "—"
	}
	d := float64(after-before) / float64(before) * 100
	switch {
	case d <= -0.5:
		return fmt.Sprintf("**%+.1f%%**", d) // improvement
	case d >= 0.5:
		return fmt.Sprintf("`%+.1f%%`", d) // regression
	default:
		return fmt.Sprintf("%+.1f%%", d)
	}
}

func humanBytes(n uint64) string {
	const (
		KB = 1 << 10
		MB = 1 << 20
		GB = 1 << 30
	)
	switch {
	case n >= GB:
		return fmt.Sprintf("%.1fGB", float64(n)/GB)
	case n >= MB:
		return fmt.Sprintf("%.1fMB", float64(n)/MB)
	case n >= KB:
		return fmt.Sprintf("%.1fKB", float64(n)/KB)
	default:
		return fmt.Sprintf("%dB", n)
	}
}

func contains(haystack, needle string) bool {
	for i := 0; i+len(needle) <= len(haystack); i++ {
		if haystack[i:i+len(needle)] == needle {
			return true
		}
	}
	return false
}

func must(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// scene.FrameWriter is satisfied by encSink above; the import is so
// it's referenced from this file for go vet's sake.
var _ scene.FrameWriter = encSink{}
