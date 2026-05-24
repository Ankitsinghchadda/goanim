// Package color provides the Excalidraw color palette used throughout goanim.
//
// Colors are pre-parsed image/color.RGBA values matching Excalidraw's
// "Hand-drawn" theme defaults. Light-mode background is white; dark-mode
// is the near-black Excalidraw uses. Stroke colors are the saturated set
// used for outlines and text; fill colors are pale tints typically used
// as shape backgrounds.
package color

import "image/color"

// hex parses a 6-digit hex string (#RRGGBB) into an RGBA at full opacity.
// Panics on malformed input — only used with compile-time constants.
func hex(s string) color.RGBA {
	if len(s) != 7 || s[0] != '#' {
		panic("invalid hex: " + s)
	}
	parse := func(c byte) uint8 {
		switch {
		case c >= '0' && c <= '9':
			return c - '0'
		case c >= 'a' && c <= 'f':
			return c - 'a' + 10
		case c >= 'A' && c <= 'F':
			return c - 'A' + 10
		}
		panic("invalid hex digit: " + string(c))
	}
	return color.RGBA{
		R: parse(s[1])<<4 | parse(s[2]),
		G: parse(s[3])<<4 | parse(s[4]),
		B: parse(s[5])<<4 | parse(s[6]),
		A: 0xFF,
	}
}

// Background colors.
var (
	BackgroundLight = hex("#FFFFFF")
	BackgroundDark  = hex("#121212")
)

// Stroke palette — the saturated outline colors.
var (
	StrokeInk    = hex("#1E1E1E") // default near-black
	StrokeRed    = hex("#E03131")
	StrokeGreen  = hex("#2F9E44")
	StrokeBlue   = hex("#1971C2")
	StrokeOrange = hex("#F08C00")
	StrokeViolet = hex("#6741D9")
	StrokeTeal   = hex("#0B7285")
	StrokeYellow = hex("#F59F00")
)

// Fill palette — pale tints that pair with the stroke colors.
var (
	FillRed     = hex("#FFC9C9")
	FillGreen   = hex("#B2F2BB")
	FillBlue    = hex("#A5D8FF")
	FillYellow  = hex("#FFEC99")
	FillOrange  = hex("#FFD8A8")
	FillViolet  = hex("#D0BFFF")
	FillTeal    = hex("#99E9F2")
	FillNeutral = hex("#E9ECEF")
)
