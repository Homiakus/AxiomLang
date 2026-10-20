package gui

import (
	"math"
	"sort"
)

// GradientStop defines a color stop along a normalized gradient axis [0.0, 1.0].
type GradientStop struct {
	Offset float64 // 0.0 to 1.0
	Color  Color32
}

// GradientType specifies the interpolation geometry.
type GradientType int

const (
	GradientLinear GradientType = iota
	GradientRadial
	GradientAngular
)

// Gradient represents a multi-stop color gradient.
type Gradient struct {
	Type   GradientType
	Stops  []GradientStop
	Angle  float64 // angle in radians (for Linear and Angular)
	CenterX float64 // center ratio [0.0, 1.0] for Radial and Angular (default 0.5)
	CenterY float64 // center ratio [0.0, 1.0] for Radial and Angular (default 0.5)
}

// NewLinearGradient creates a linear gradient given an angle in radians and color stops.
func NewLinearGradient(angleRad float64, stops ...GradientStop) Gradient {
	sortedStops := make([]GradientStop, len(stops))
	copy(sortedStops, stops)
	sort.Slice(sortedStops, func(i, j int) bool {
		return sortedStops[i].Offset < sortedStops[j].Offset
	})
	return Gradient{
		Type:    GradientLinear,
		Angle:   angleRad,
		Stops:   sortedStops,
		CenterX: 0.5,
		CenterY: 0.5,
	}
}

// NewRadialGradient creates a radial gradient centered at (cx, cy) [0..1] with stops.
func NewRadialGradient(cx, cy float64, stops ...GradientStop) Gradient {
	sortedStops := make([]GradientStop, len(stops))
	copy(sortedStops, stops)
	sort.Slice(sortedStops, func(i, j int) bool {
		return sortedStops[i].Offset < sortedStops[j].Offset
	})
	return Gradient{
		Type:    GradientRadial,
		Stops:   sortedStops,
		CenterX: cx,
		CenterY: cy,
	}
}

// NewAngularGradient creates a sweep / conical / angular gradient.
func NewAngularGradient(cx, cy, startAngle float64, stops ...GradientStop) Gradient {
	sortedStops := make([]GradientStop, len(stops))
	copy(sortedStops, stops)
	sort.Slice(sortedStops, func(i, j int) bool {
		return sortedStops[i].Offset < sortedStops[j].Offset
	})
	return Gradient{
		Type:    GradientAngular,
		Angle:   startAngle,
		Stops:   sortedStops,
		CenterX: cx,
		CenterY: cy,
	}
}

// Sample interpolates color at position t in [0.0, 1.0].
func (g Gradient) Sample(t float64) Color32 {
	if len(g.Stops) == 0 {
		return ARGB(255, 0, 0, 0)
	}
	if len(g.Stops) == 1 || t <= g.Stops[0].Offset {
		return g.Stops[0].Color
	}
	if t >= g.Stops[len(g.Stops)-1].Offset {
		return g.Stops[len(g.Stops)-1].Color
	}

	for i := 0; i < len(g.Stops)-1; i++ {
		s0 := g.Stops[i]
		s1 := g.Stops[i+1]
		if t >= s0.Offset && t <= s1.Offset {
			span := s1.Offset - s0.Offset
			if span <= 0.00001 {
				return s0.Color
			}
			factor := (t - s0.Offset) / span
			return LerpColor(s0.Color, s1.Color, factor)
		}
	}
	return g.Stops[len(g.Stops)-1].Color
}

// LerpColor smoothly linearly interpolates between two Color32 values.
func LerpColor(c1, c2 Color32, t float64) Color32 {
	if t <= 0 {
		return c1
	}
	if t >= 1 {
		return c2
	}
	f := float32(t)
	invF := 1.0 - f

	a1, r1, g1, b1 := float32((c1>>24)&0xFF), float32((c1>>16)&0xFF), float32((c1>>8)&0xFF), float32(c1&0xFF)
	a2, r2, g2, b2 := float32((c2>>24)&0xFF), float32((c2>>16)&0xFF), float32((c2>>8)&0xFF), float32(c2&0xFF)

	outA := uint8(a1*invF + a2*f)
	outR := uint8(r1*invF + r2*f)
	outG := uint8(g1*invF + g2*f)
	outB := uint8(b1*invF + b2*f)

	return ARGB(outA, outR, outG, outB)
}

// FillLinearGradient fills a rectangle with a high-performance linear gradient.
func (c *NativeCanvas) FillLinearGradient(x, y, w, h int, grad Gradient) {
	if w <= 0 || h <= 0 {
		return
	}
	cosA := math.Cos(grad.Angle)
	sinA := math.Sin(grad.Angle)

	x0 := max(0, x)
	y0 := max(0, y)
	x1 := min(c.Width, x+w)
	y1 := min(c.Height, y+h)

	// Length projection along direction vector
	diag := math.Abs(float64(w)*cosA) + math.Abs(float64(h)*sinA)
	if diag <= 0.0001 {
		diag = 1.0
	}
	invDiag := 1.0 / diag

	// Precompute LUT for faster per-pixel lookup if span is large
	const lutSize = 512
	var lut [lutSize]Color32
	for i := 0; i < lutSize; i++ {
		lut[i] = grad.Sample(float64(i) / float64(lutSize-1))
	}

	for cy := y0; cy < y1; cy++ {
		dy := float64(cy - y)
		rowOffset := cy * c.Width
		for cx := x0; cx < x1; cx++ {
			dx := float64(cx - x)
			// Projection onto direction
			proj := (dx*cosA + dy*sinA) * invDiag
			if proj < 0 {
				proj = 0
			} else if proj > 1 {
				proj = 1
			}
			lutIdx := int(proj * float64(lutSize-1))
			c.BlendPixel(cx, cy, lut[lutIdx])
		}
		_ = rowOffset
	}
}

// FillRadialGradient fills a rectangle with a radial gradient centered at (cx, cy).
func (c *NativeCanvas) FillRadialGradient(x, y, w, h int, grad Gradient) {
	if w <= 0 || h <= 0 {
		return
	}
	centerX := float64(x) + grad.CenterX*float64(w)
	centerY := float64(y) + grad.CenterY*float64(h)
	maxRadius := math.Hypot(math.Max(grad.CenterX, 1.0-grad.CenterX)*float64(w), math.Max(grad.CenterY, 1.0-grad.CenterY)*float64(h))
	if maxRadius <= 0.0001 {
		maxRadius = 1.0
	}
	invRadius := 1.0 / maxRadius

	x0 := max(0, x)
	y0 := max(0, y)
	x1 := min(c.Width, x+w)
	y1 := min(c.Height, y+h)

	const lutSize = 512
	var lut [lutSize]Color32
	for i := 0; i < lutSize; i++ {
		lut[i] = grad.Sample(float64(i) / float64(lutSize-1))
	}

	for cy := y0; cy < y1; cy++ {
		dy := float64(cy) - centerY
		dySq := dy * dy
		for cx := x0; cx < x1; cx++ {
			dx := float64(cx) - centerX
			dist := math.Sqrt(dx*dx + dySq)
			t := dist * invRadius
			if t > 1.0 {
				t = 1.0
			}
			lutIdx := int(t * float64(lutSize-1))
			c.BlendPixel(cx, cy, lut[lutIdx])
		}
	}
}

// FillRoundedGradient fills a rounded rectangle with a linear gradient.
func (c *NativeCanvas) FillRoundedGradient(x, y, w, h, r int, grad Gradient) {
	if w <= 0 || h <= 0 {
		return
	}
	if r <= 0 {
		c.FillLinearGradient(x, y, w, h, grad)
		return
	}
	if r > w/2 {
		r = w / 2
	}
	if r > h/2 {
		r = h / 2
	}

	cosA := math.Cos(grad.Angle)
	sinA := math.Sin(grad.Angle)
	diag := math.Abs(float64(w)*cosA) + math.Abs(float64(h)*sinA)
	if diag <= 0.0001 {
		diag = 1.0
	}
	invDiag := 1.0 / diag

	const lutSize = 512
	var lut [lutSize]Color32
	for i := 0; i < lutSize; i++ {
		lut[i] = grad.Sample(float64(i) / float64(lutSize-1))
	}

	rSq := r * r
	y0 := max(0, y)
	y1 := min(c.Height, y+h)

	for cy := y0; cy < y1; cy++ {
		var leftOffset, rightOffset int
		if cy < y+r {
			dy := (y + r) - cy
			dx := int(math.Sqrt(float64(rSq - dy*dy)))
			leftOffset = r - dx
			rightOffset = r - dx
		} else if cy >= y+h-r {
			dy := cy - (y + h - r - 1)
			dx := int(math.Sqrt(float64(rSq - dy*dy)))
			leftOffset = r - dx
			rightOffset = r - dx
		}

		spanX0 := max(0, x+leftOffset)
		spanX1 := min(c.Width, x+w-rightOffset)
		dyGrad := float64(cy - y)

		for cx := spanX0; cx < spanX1; cx++ {
			dxGrad := float64(cx - x)
			proj := (dxGrad*cosA + dyGrad*sinA) * invDiag
			if proj < 0 {
				proj = 0
			} else if proj > 1 {
				proj = 1
			}
			lutIdx := int(proj * float64(lutSize-1))
			c.BlendPixel(cx, cy, lut[lutIdx])
		}
	}
}
