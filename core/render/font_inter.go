package render

import (
	_ "embed"
	"fmt"
)

//go:embed embed/Inter-Regular.ttf
var interTTF []byte

// Inter returns a FontFace backed by Inter (OFL 1.1) — the bundled
// sans-serif used for the Architect style.
func Inter() (FontFace, error) {
	if len(interTTF) == 0 {
		return nil, fmt.Errorf("render: Inter not embedded")
	}
	return LoadFont("Inter", interTTF)
}
