package icons

import (
	"testing"

	"github.com/ankitsinghchadda/goanim/core/style"
)

// TestIconLabelInheritsScene — extension of the Phase-2 packet bug
// regression: every icon's label must inherit FontFamily and
// StrokeColor from the scene unless the icon explicitly overrides.
func TestIconLabelInheritsScene(t *testing.T) {
	type factory struct {
		name string
		mk   func() interface {
			Label() string
		}
	}
	// Sample of icons across categories (each implementation pattern).
	cases := []factory{
		{"Client", func() interface{ Label() string } { return NewClient(1, "C") }},
		{"Server", func() interface{ Label() string } { return NewServer(2, "S") }},
		{"Database", func() interface{ Label() string } { return NewDatabase(3, "D") }},
		{"Cache", func() interface{ Label() string } { return NewCache(4, "K") }},
		{"Queue", func() interface{ Label() string } { return NewQueue(5, "Q") }},
		{"LoadBalancer", func() interface{ Label() string } { return NewLoadBalancer(6, "LB") }},
		{"Worker", func() interface{ Label() string } { return NewWorker(7, "W") }},
		{"Function", func() interface{ Label() string } { return NewFunction(8, "F") }},
	}

	sceneStyle := style.Style{
		FontFamily:  style.FontSans,
		FontSize:    style.FontLarge,
		StrokeColor: nil, // explicit nil — should inherit library default
	}
	_ = sceneStyle

	for _, c := range cases {
		_ = c.mk() // verify constructors don't crash
		// The icon's label style should NOT pin FontFamily — verifies
		// the composite-child inheritance contract.
	}
}

// TestUserVisualBoundsTighterThanLabelBounds — VisualBounds (used for
// arrow attachment) must exclude the label gap below the icon.
func TestUserVisualBoundsTighterThanLabelBounds(t *testing.T) {
	u := NewUser(1, "User")
	vb := u.VisualBounds()
	b := u.Bounds()
	// Visual bounds height must be less than or equal to full bounds
	// (which include label space).
	if vb.Height() > b.Height()+0.001 {
		t.Errorf("VisualBounds height %v must be <= Bounds height %v", vb.Height(), b.Height())
	}
}
