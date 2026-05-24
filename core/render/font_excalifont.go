package render

import (
	_ "embed"
	"fmt"
)

//go:embed embed/Excalifont-Regular.ttf
var excalifontTTF []byte

// Excalifont returns a FontFace backed by Excalidraw's bundled Excalifont
// (OFL 1.1). The TTF is embedded at compile time.
func Excalifont() (FontFace, error) {
	if len(excalifontTTF) == 0 {
		return nil, fmt.Errorf("render: Excalifont not embedded")
	}
	return LoadFont("Excalifont", excalifontTTF)
}
