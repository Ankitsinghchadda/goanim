package layout

// VerticalAlign positions a row's children along the cross axis.
type VerticalAlign uint8

const (
	VTop    VerticalAlign = iota // align top edges
	VMiddle                      // align vertical centers (default for HBox)
	VBottom                      // align bottom edges
)

// HorizontalAlign positions a column's children along the cross axis.
type HorizontalAlign uint8

const (
	HLeft   HorizontalAlign = iota
	HCenter                 // default for VBox
	HRight
)

// Anchor names a relative position between two mobjects (used by AlignTo).
type Anchor uint8

const (
	AnchorAbove Anchor = iota
	AnchorBelow
	AnchorLeftOf
	AnchorRightOf
	AnchorCenter
)
