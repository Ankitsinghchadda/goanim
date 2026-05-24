package geometry

// CmdKind identifies a path command.
type CmdKind uint8

const (
	// CmdMove starts a new subpath at P0.
	CmdMove CmdKind = iota
	// CmdLine draws a straight line from the current point to P0.
	CmdLine
	// CmdCurve draws a cubic Bézier from the current point with controls
	// P0, P1 to endpoint P2.
	CmdCurve
	// CmdClose closes the current subpath with a straight line to its start.
	CmdClose
)

// Cmd is a single path command. Only the fields relevant to Kind are populated.
//
//	CmdMove   — P0 = target
//	CmdLine   — P0 = target
//	CmdCurve  — P0 = ctrl1, P1 = ctrl2, P2 = end
//	CmdClose  — no points
type Cmd struct {
	Kind       CmdKind
	P0, P1, P2 Point
}

// Path is a sequence of drawing commands. A Path may contain multiple
// subpaths (a new CmdMove starts a new subpath). Paths are pure data —
// they don't know how to render themselves; renderers consume them.
type Path struct {
	Cmds []Cmd
}

// NewPath constructs an empty path. Use the builder methods to append.
func NewPath() *Path { return &Path{} }

// MoveTo appends a move command.
func (p *Path) MoveTo(x, y float64) *Path {
	p.Cmds = append(p.Cmds, Cmd{Kind: CmdMove, P0: Point{x, y}})
	return p
}

// LineTo appends a line command.
func (p *Path) LineTo(x, y float64) *Path {
	p.Cmds = append(p.Cmds, Cmd{Kind: CmdLine, P0: Point{x, y}})
	return p
}

// CurveTo appends a cubic Bézier with control points (cx1, cy1), (cx2, cy2)
// and endpoint (x, y).
func (p *Path) CurveTo(cx1, cy1, cx2, cy2, x, y float64) *Path {
	p.Cmds = append(p.Cmds, Cmd{
		Kind: CmdCurve,
		P0:   Point{cx1, cy1},
		P1:   Point{cx2, cy2},
		P2:   Point{x, y},
	})
	return p
}

// Close appends a close command.
func (p *Path) Close() *Path {
	p.Cmds = append(p.Cmds, Cmd{Kind: CmdClose})
	return p
}

// Append concatenates the commands of q onto p.
func (p *Path) Append(q *Path) *Path {
	if q != nil {
		p.Cmds = append(p.Cmds, q.Cmds...)
	}
	return p
}

// Transform returns a new Path with t applied to every point.
func (p *Path) Transform(t Transform) *Path {
	out := &Path{Cmds: make([]Cmd, len(p.Cmds))}
	for i, c := range p.Cmds {
		nc := Cmd{Kind: c.Kind}
		switch c.Kind {
		case CmdMove, CmdLine:
			nc.P0 = t.Apply(c.P0)
		case CmdCurve:
			nc.P0 = t.Apply(c.P0)
			nc.P1 = t.Apply(c.P1)
			nc.P2 = t.Apply(c.P2)
		}
		out.Cmds[i] = nc
	}
	return out
}

// Bounds returns the axis-aligned bounding box of all anchor and control
// points in the path. (For tight bounds of curves, callers should flatten
// or compute extrema — this is sufficient for layout, not for hit-testing.)
func (p *Path) Bounds() Rect {
	if len(p.Cmds) == 0 {
		return Rect{Min: Point{1, 1}, Max: Point{-1, -1}}
	}
	var pts []Point
	for _, c := range p.Cmds {
		switch c.Kind {
		case CmdMove, CmdLine:
			pts = append(pts, c.P0)
		case CmdCurve:
			pts = append(pts, c.P0, c.P1, c.P2)
		}
	}
	return RectFromPoints(pts...)
}
