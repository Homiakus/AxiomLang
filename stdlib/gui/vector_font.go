package gui

import (
	"math"
)

// Point represents a 2D floating-point coordinate.
type Point struct {
	X, Y float64
}

// VectorPathSegment represents a path command (MoveTo, LineTo, QuadTo, CubicTo).
type PathCommand int

const (
	CmdMoveTo PathCommand = iota
	CmdLineTo
	CmdQuadTo
	CmdCubicTo
	CmdClose
)

type PathSegment struct {
	Command PathCommand
	Points  []Point
}

// VectorGlyph contains resolution-independent vector outlines for a character.
type VectorGlyph struct {
	Rune     rune
	Advance  float64
	Segments []PathSegment
}

// VectorFont contains scalable vector outline representations for clean antialiased rendering at any size.
type VectorFont struct {
	Name   string
	Glyphs map[rune]VectorGlyph
}

// DefaultVectorFont returns a built-in clean geometric sans-serif vector font.
var DefaultVectorFont = initDefaultVectorFont()

func initDefaultVectorFont() *VectorFont {
	font := &VectorFont{
		Name:   "AxiomSansVector",
		Glyphs: make(map[rune]VectorGlyph),
	}

	// Normalized glyph definitions [0..100] coordinate space
	// A
	font.Glyphs['A'] = VectorGlyph{
		Rune: 'A', Advance: 70,
		Segments: []PathSegment{
			{Command: CmdMoveTo, Points: []Point{{10, 100}}},
			{Command: CmdLineTo, Points: []Point{{35, 10}}},
			{Command: CmdLineTo, Points: []Point{{60, 100}}},
			{Command: CmdMoveTo, Points: []Point{{20, 65}}},
			{Command: CmdLineTo, Points: []Point{{50, 65}}},
		},
	}
	// B
	font.Glyphs['B'] = VectorGlyph{
		Rune: 'B', Advance: 65,
		Segments: []PathSegment{
			{Command: CmdMoveTo, Points: []Point{{15, 100}}},
			{Command: CmdLineTo, Points: []Point{{15, 10}}},
			{Command: CmdLineTo, Points: []Point{{40, 10}}},
			{Command: CmdCubicTo, Points: []Point{{60, 10}, {60, 52}, {35, 52}}},
			{Command: CmdLineTo, Points: []Point{{15, 52}}},
			{Command: CmdMoveTo, Points: []Point{{35, 52}}},
			{Command: CmdCubicTo, Points: []Point{{65, 52}, {65, 100}, {35, 100}}},
			{Command: CmdLineTo, Points: []Point{{15, 100}}},
		},
	}
	// C
	font.Glyphs['C'] = VectorGlyph{
		Rune: 'C', Advance: 65,
		Segments: []PathSegment{
			{Command: CmdMoveTo, Points: []Point{{55, 25}}},
			{Command: CmdCubicTo, Points: []Point{{45, 10}, {20, 10}, {15, 55}}},
			{Command: CmdCubicTo, Points: []Point{{15, 85}, {45, 100}, {55, 85}}},
		},
	}
	// D
	font.Glyphs['D'] = VectorGlyph{
		Rune: 'D', Advance: 70,
		Segments: []PathSegment{
			{Command: CmdMoveTo, Points: []Point{{15, 100}}},
			{Command: CmdLineTo, Points: []Point{{15, 10}}},
			{Command: CmdLineTo, Points: []Point{{35, 10}}},
			{Command: CmdCubicTo, Points: []Point{{65, 10}, {65, 100}, {35, 100}}},
			{Command: CmdLineTo, Points: []Point{{15, 100}}},
		},
	}
	// E
	font.Glyphs['E'] = VectorGlyph{
		Rune: 'E', Advance: 60,
		Segments: []PathSegment{
			{Command: CmdMoveTo, Points: []Point{{50, 10}}},
			{Command: CmdLineTo, Points: []Point{{15, 10}}},
			{Command: CmdLineTo, Points: []Point{{15, 100}}},
			{Command: CmdLineTo, Points: []Point{{50, 100}}},
			{Command: CmdMoveTo, Points: []Point{{15, 55}}},
			{Command: CmdLineTo, Points: []Point{{45, 55}}},
		},
	}
	// F
	font.Glyphs['F'] = VectorGlyph{
		Rune: 'F', Advance: 55,
		Segments: []PathSegment{
			{Command: CmdMoveTo, Points: []Point{{50, 10}}},
			{Command: CmdLineTo, Points: []Point{{15, 10}}},
			{Command: CmdLineTo, Points: []Point{{15, 100}}},
			{Command: CmdMoveTo, Points: []Point{{15, 55}}},
			{Command: CmdLineTo, Points: []Point{{42, 55}}},
		},
	}
	// G
	font.Glyphs['G'] = VectorGlyph{
		Rune: 'G', Advance: 70,
		Segments: []PathSegment{
			{Command: CmdMoveTo, Points: []Point{{60, 25}}},
			{Command: CmdCubicTo, Points: []Point{{45, 10}, {20, 10}, {15, 55}}},
			{Command: CmdCubicTo, Points: []Point{{15, 85}, {45, 100}, {60, 85}}},
			{Command: CmdLineTo, Points: []Point{{60, 55}}},
			{Command: CmdLineTo, Points: []Point{{38, 55}}},
		},
	}
	// H
	font.Glyphs['H'] = VectorGlyph{
		Rune: 'H', Advance: 65,
		Segments: []PathSegment{
			{Command: CmdMoveTo, Points: []Point{{15, 10}}},
			{Command: CmdLineTo, Points: []Point{{15, 100}}},
			{Command: CmdMoveTo, Points: []Point{{55, 10}}},
			{Command: CmdLineTo, Points: []Point{{55, 100}}},
			{Command: CmdMoveTo, Points: []Point{{15, 55}}},
			{Command: CmdLineTo, Points: []Point{{55, 55}}},
		},
	}
	// I
	font.Glyphs['I'] = VectorGlyph{
		Rune: 'I', Advance: 30,
		Segments: []PathSegment{
			{Command: CmdMoveTo, Points: []Point{{15, 10}}},
			{Command: CmdLineTo, Points: []Point{{15, 100}}},
			{Command: CmdMoveTo, Points: []Point{{5, 10}}},
			{Command: CmdLineTo, Points: []Point{{25, 10}}},
			{Command: CmdMoveTo, Points: []Point{{5, 100}}},
			{Command: CmdLineTo, Points: []Point{{25, 100}}},
		},
	}
	// L
	font.Glyphs['L'] = VectorGlyph{
		Rune: 'L', Advance: 55,
		Segments: []PathSegment{
			{Command: CmdMoveTo, Points: []Point{{15, 10}}},
			{Command: CmdLineTo, Points: []Point{{15, 100}}},
			{Command: CmdLineTo, Points: []Point{{50, 100}}},
		},
	}
	// M
	font.Glyphs['M'] = VectorGlyph{
		Rune: 'M', Advance: 80,
		Segments: []PathSegment{
			{Command: CmdMoveTo, Points: []Point{{10, 100}}},
			{Command: CmdLineTo, Points: []Point{{10, 10}}},
			{Command: CmdLineTo, Points: []Point{{40, 60}}},
			{Command: CmdLineTo, Points: []Point{{70, 10}}},
			{Command: CmdLineTo, Points: []Point{{70, 100}}},
		},
	}
	// N
	font.Glyphs['N'] = VectorGlyph{
		Rune: 'N', Advance: 70,
		Segments: []PathSegment{
			{Command: CmdMoveTo, Points: []Point{{15, 100}}},
			{Command: CmdLineTo, Points: []Point{{15, 10}}},
			{Command: CmdLineTo, Points: []Point{{55, 100}}},
			{Command: CmdLineTo, Points: []Point{{55, 10}}},
		},
	}
	// O / 0
	font.Glyphs['O'] = VectorGlyph{
		Rune: 'O', Advance: 70,
		Segments: []PathSegment{
			{Command: CmdMoveTo, Points: []Point{{35, 10}}},
			{Command: CmdCubicTo, Points: []Point{{65, 10}, {65, 100}, {35, 100}}},
			{Command: CmdCubicTo, Points: []Point{{5, 100}, {5, 10}, {35, 10}}},
		},
	}
	font.Glyphs['0'] = font.Glyphs['O']

	// P
	font.Glyphs['P'] = VectorGlyph{
		Rune: 'P', Advance: 60,
		Segments: []PathSegment{
			{Command: CmdMoveTo, Points: []Point{{15, 100}}},
			{Command: CmdLineTo, Points: []Point{{15, 10}}},
			{Command: CmdLineTo, Points: []Point{{35, 10}}},
			{Command: CmdCubicTo, Points: []Point{{60, 10}, {60, 60}, {35, 60}}},
			{Command: CmdLineTo, Points: []Point{{15, 60}}},
		},
	}
	// R
	font.Glyphs['R'] = VectorGlyph{
		Rune: 'R', Advance: 65,
		Segments: []PathSegment{
			{Command: CmdMoveTo, Points: []Point{{15, 100}}},
			{Command: CmdLineTo, Points: []Point{{15, 10}}},
			{Command: CmdLineTo, Points: []Point{{35, 10}}},
			{Command: CmdCubicTo, Points: []Point{{60, 10}, {60, 55}, {35, 55}}},
			{Command: CmdLineTo, Points: []Point{{15, 55}}},
			{Command: CmdMoveTo, Points: []Point{{35, 55}}},
			{Command: CmdLineTo, Points: []Point{{60, 100}}},
		},
	}
	// S
	font.Glyphs['S'] = VectorGlyph{
		Rune: 'S', Advance: 60,
		Segments: []PathSegment{
			{Command: CmdMoveTo, Points: []Point{{50, 25}}},
			{Command: CmdCubicTo, Points: []Point{{45, 10}, {15, 10}, {20, 45}}},
			{Command: CmdCubicTo, Points: []Point{{25, 60}, {55, 60}, {50, 85}}},
			{Command: CmdCubicTo, Points: []Point{{45, 100}, {15, 100}, {10, 80}}},
		},
	}
	// T
	font.Glyphs['T'] = VectorGlyph{
		Rune: 'T', Advance: 60,
		Segments: []PathSegment{
			{Command: CmdMoveTo, Points: []Point{{10, 10}}},
			{Command: CmdLineTo, Points: []Point{{55, 10}}},
			{Command: CmdMoveTo, Points: []Point{{32, 10}}},
			{Command: CmdLineTo, Points: []Point{{32, 100}}},
		},
	}
	// U
	font.Glyphs['U'] = VectorGlyph{
		Rune: 'U', Advance: 65,
		Segments: []PathSegment{
			{Command: CmdMoveTo, Points: []Point{{15, 10}}},
			{Command: CmdLineTo, Points: []Point{{15, 75}}},
			{Command: CmdCubicTo, Points: []Point{{15, 100}, {55, 100}, {55, 75}}},
			{Command: CmdLineTo, Points: []Point{{55, 10}}},
		},
	}
	// V
	font.Glyphs['V'] = VectorGlyph{
		Rune: 'V', Advance: 65,
		Segments: []PathSegment{
			{Command: CmdMoveTo, Points: []Point{{10, 10}}},
			{Command: CmdLineTo, Points: []Point{{35, 100}}},
			{Command: CmdLineTo, Points: []Point{{60, 10}}},
		},
	}
	// W
	font.Glyphs['W'] = VectorGlyph{
		Rune: 'W', Advance: 80,
		Segments: []PathSegment{
			{Command: CmdMoveTo, Points: []Point{{10, 10}}},
			{Command: CmdLineTo, Points: []Point{{25, 100}}},
			{Command: CmdLineTo, Points: []Point{{40, 40}}},
			{Command: CmdLineTo, Points: []Point{{55, 100}}},
			{Command: CmdLineTo, Points: []Point{{70, 10}}},
		},
	}
	// X
	font.Glyphs['X'] = VectorGlyph{
		Rune: 'X', Advance: 60,
		Segments: []PathSegment{
			{Command: CmdMoveTo, Points: []Point{{10, 10}}},
			{Command: CmdLineTo, Points: []Point{{50, 100}}},
			{Command: CmdMoveTo, Points: []Point{{50, 10}}},
			{Command: CmdLineTo, Points: []Point{{10, 100}}},
		},
	}
	// Y
	font.Glyphs['Y'] = VectorGlyph{
		Rune: 'Y', Advance: 60,
		Segments: []PathSegment{
			{Command: CmdMoveTo, Points: []Point{{10, 10}}},
			{Command: CmdLineTo, Points: []Point{{32, 55}}},
			{Command: CmdLineTo, Points: []Point{{55, 10}}},
			{Command: CmdMoveTo, Points: []Point{{32, 55}}},
			{Command: CmdLineTo, Points: []Point{{32, 100}}},
		},
	}
	// Z
	font.Glyphs['Z'] = VectorGlyph{
		Rune: 'Z', Advance: 60,
		Segments: []PathSegment{
			{Command: CmdMoveTo, Points: []Point{{10, 10}}},
			{Command: CmdLineTo, Points: []Point{{50, 10}}},
			{Command: CmdLineTo, Points: []Point{{10, 100}}},
			{Command: CmdLineTo, Points: []Point{{50, 100}}},
		},
	}
	// 1
	font.Glyphs['1'] = VectorGlyph{
		Rune: '1', Advance: 45,
		Segments: []PathSegment{
			{Command: CmdMoveTo, Points: []Point{{15, 25}}},
			{Command: CmdLineTo, Points: []Point{{28, 10}}},
			{Command: CmdLineTo, Points: []Point{{28, 100}}},
			{Command: CmdMoveTo, Points: []Point{{15, 100}}},
			{Command: CmdLineTo, Points: []Point{{40, 100}}},
		},
	}
	// 2
	font.Glyphs['2'] = VectorGlyph{
		Rune: '2', Advance: 60,
		Segments: []PathSegment{
			{Command: CmdMoveTo, Points: []Point{{12, 30}}},
			{Command: CmdCubicTo, Points: []Point{{15, 10}, {45, 10}, {50, 30}}},
			{Command: CmdCubicTo, Points: []Point{{50, 50}, {15, 75}, {10, 100}}},
			{Command: CmdLineTo, Points: []Point{{52, 100}}},
		},
	}
	// Space
	font.Glyphs[' '] = VectorGlyph{
		Rune: ' ', Advance: 35,
		Segments: []PathSegment{},
	}
	// -
	font.Glyphs['-'] = VectorGlyph{
		Rune: '-', Advance: 40,
		Segments: []PathSegment{
			{Command: CmdMoveTo, Points: []Point{{8, 55}}},
			{Command: CmdLineTo, Points: []Point{{32, 55}}},
		},
	}
	// :
	font.Glyphs[':'] = VectorGlyph{
		Rune: ':', Advance: 30,
		Segments: []PathSegment{
			{Command: CmdMoveTo, Points: []Point{{15, 35}}},
			{Command: CmdLineTo, Points: []Point{{15, 40}}},
			{Command: CmdMoveTo, Points: []Point{{15, 75}}},
			{Command: CmdLineTo, Points: []Point{{15, 80}}},
		},
	}
	// .
	font.Glyphs['.'] = VectorGlyph{
		Rune: '.', Advance: 25,
		Segments: []PathSegment{
			{Command: CmdMoveTo, Points: []Point{{12, 90}}},
			{Command: CmdLineTo, Points: []Point{{12, 95}}},
		},
	}

	return font
}

// DrawVectorLineAA draws an antialiased line segment with subpixel width using Wu's antialiasing algorithm.
func (c *NativeCanvas) DrawVectorLineAA(x0, y0, x1, y1 float64, color Color32, strokeWidth float64) {
	dx := x1 - x0
	dy := y1 - y0
	length := math.Hypot(dx, dy)
	if length <= 0.0001 {
		return
	}

	// Step count based on length for smooth sampling
	steps := int(length * 2.0)
	if steps < 1 {
		steps = 1
	}

	stepX := dx / float64(steps)
	stepY := dy / float64(steps)
	halfW := strokeWidth * 0.5

	baseA := float64(color>>24) / 255.0
	r := uint8((color >> 16) & 0xFF)
	g := uint8((color >> 8) & 0xFF)
	b := uint8(color & 0xFF)

	cx := x0
	cy := y0
	for s := 0; s <= steps; s++ {
		ix0 := int(math.Floor(cx - halfW))
		ix1 := int(math.Ceil(cx + halfW))
		iy0 := int(math.Floor(cy - halfW))
		iy1 := int(math.Ceil(cy + halfW))

		for py := iy0; py <= iy1; py++ {
			for px := ix0; px <= ix1; px++ {
				dist := math.Hypot(float64(px)-cx, float64(py)-cy)
				if dist <= halfW {
					alphaCoverage := 1.0 - (dist / (halfW + 0.5))
					if alphaCoverage > 0 {
						pixAlpha := uint8(alphaCoverage * baseA * 255.0)
						c.BlendPixel(px, py, ARGB(pixAlpha, r, g, b))
					}
				}
			}
		}

		cx += stepX
		cy += stepY
	}
}

// DrawCubicBezierAA draws an antialiased cubic bezier curve: P0 -> P1 -> P2 -> P3.
func (c *NativeCanvas) DrawCubicBezierAA(p0, p1, p2, p3 Point, color Color32, strokeWidth float64) {
	// Approximate curve length to choose subdivision steps
	chord := math.Hypot(p3.X-p0.X, p3.Y-p0.Y)
	poly := math.Hypot(p1.X-p0.X, p1.Y-p0.Y) + math.Hypot(p2.X-p1.X, p2.Y-p1.Y) + math.Hypot(p3.X-p2.X, p3.Y-p2.Y)
	steps := int((chord+poly)*0.5 + 4)
	if steps > 64 {
		steps = 64
	}

	dt := 1.0 / float64(steps)
	prev := p0
	for i := 1; i <= steps; i++ {
		t := float64(i) * dt
		invT := 1.0 - t
		// B(t) = (1-t)^3*P0 + 3(1-t)^2*t*P1 + 3(1-t)*t^2*P2 + t^3*P3
		currX := invT*invT*invT*p0.X + 3*invT*invT*t*p1.X + 3*invT*t*t*p2.X + t*t*t*p3.X
		currY := invT*invT*invT*p0.Y + 3*invT*invT*t*p1.Y + 3*invT*t*t*p2.Y + t*t*t*p3.Y
		c.DrawVectorLineAA(prev.X, prev.Y, currX, currY, color, strokeWidth)
		prev = Point{X: currX, Y: currY}
	}
}

// DrawVectorChar renders a high-res vector glyph with custom pixel size and antialiasing.
func (c *NativeCanvas) DrawVectorChar(x, y float64, ch rune, color Color32, pixelSize float64) float64 {
	glyph, ok := DefaultVectorFont.Glyphs[ch]
	if !ok {
		// Uppercase fallback
		if ch >= 'a' && ch <= 'z' {
			glyph, ok = DefaultVectorFont.Glyphs[ch-32]
		}
	}
	if !ok {
		// Fallback to bitmap font rendering if glyph not defined in vector font
		c.DrawChar(int(x), int(y), ch, color, int(pixelSize/16)+1)
		return pixelSize * 0.6
	}

	scale := pixelSize / 100.0
	strokeW := math.Max(1.5, pixelSize*0.08)

	var lastPoint Point
	for _, seg := range glyph.Segments {
		switch seg.Command {
		case CmdMoveTo:
			if len(seg.Points) > 0 {
				lastPoint = Point{
					X: x + seg.Points[0].X*scale,
					Y: y + seg.Points[0].Y*scale,
				}
			}
		case CmdLineTo:
			if len(seg.Points) > 0 {
				nextPoint := Point{
					X: x + seg.Points[0].X*scale,
					Y: y + seg.Points[0].Y*scale,
				}
				c.DrawVectorLineAA(lastPoint.X, lastPoint.Y, nextPoint.X, nextPoint.Y, color, strokeW)
				lastPoint = nextPoint
			}
		case CmdCubicTo:
			if len(seg.Points) >= 3 {
				p1 := Point{X: x + seg.Points[0].X*scale, Y: y + seg.Points[0].Y*scale}
				p2 := Point{X: x + seg.Points[1].X*scale, Y: y + seg.Points[1].Y*scale}
				p3 := Point{X: x + seg.Points[2].X*scale, Y: y + seg.Points[2].Y*scale}
				c.DrawCubicBezierAA(lastPoint, p1, p2, p3, color, strokeW)
				lastPoint = p3
			}
		}
	}

	return glyph.Advance * scale
}

// DrawVectorString renders smooth antialiased vector typography at any arbitrary size.
func (c *NativeCanvas) DrawVectorString(x, y float64, text string, color Color32, pixelSize float64) {
	cx := x
	cy := y
	lineHeight := pixelSize * 1.25

	for _, ch := range text {
		if ch == '\n' {
			cy += lineHeight
			cx = x
			continue
		}
		advance := c.DrawVectorChar(cx, cy, ch, color, pixelSize)
		cx += advance + pixelSize*0.05
	}
}
