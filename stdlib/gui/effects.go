package gui

import (
	"math"
)

// EasingFunction calculates transition factor [0..1] given time progress t in [0..1].
type EasingFunction func(t float64) float64

var (
	// Linear interpolation.
	EaseLinear EasingFunction = func(t float64) float64 {
		return t
	}

	// EaseInQuad (quadratic acceleration).
	EaseInQuad EasingFunction = func(t float64) float64 {
		return t * t
	}

	// EaseOutQuad (quadratic deceleration).
	EaseOutQuad EasingFunction = func(t float64) float64 {
		return t * (2 - t)
	}

	// EaseInOutCubic (smooth S-curve).
	EaseInOutCubic EasingFunction = func(t float64) float64 {
		if t < 0.5 {
			return 4 * t * t * t
		}
		p := 2*t - 2
		return 0.5*p*p*p + 1
	}

	// EaseOutElastic (spring overshoot effect).
	EaseOutElastic EasingFunction = func(t float64) float64 {
		if t <= 0 {
			return 0
		}
		if t >= 1 {
			return 1
		}
		p := 0.3
		s := p / 4
		return math.Pow(2, -10*t)*math.Sin((t-s)*(2*math.Pi)/p) + 1
	}

	// EaseOutBounce (glass bounce effect).
	EaseOutBounce EasingFunction = func(t float64) float64 {
		if t < (1 / 2.75) {
			return 7.5625 * t * t
		} else if t < (2 / 2.75) {
			t -= (1.5 / 2.75)
			return 7.5625*t*t + 0.75
		} else if t < (2.5 / 2.75) {
			t -= (2.25 / 2.75)
			return 7.5625*t*t + 0.9375
		} else {
			t -= (2.625 / 2.75)
			return 7.5625*t*t + 0.984375
		}
	}
)

// Transition calculates interpolated scalar value between start and end.
func Transition(start, end float64, progress float64, easing EasingFunction) float64 {
	if progress <= 0 {
		return start
	}
	if progress >= 1 {
		return end
	}
	if easing == nil {
		easing = EaseInOutCubic
	}
	k := easing(progress)
	return start + (end-start)*k
}

// TransitionColor calculates interpolated Color32 between start and end.
func TransitionColor(start, end Color32, progress float64, easing EasingFunction) Color32 {
	if easing == nil {
		easing = EaseInOutCubic
	}
	k := easing(progress)
	return LerpColor(start, end, k)
}

// DrawGlow renders a smooth radial/box aura around a rectangle.
func (c *NativeCanvas) DrawGlow(x, y, w, h, radius int, glowColor Color32) {
	if radius <= 0 || w <= 0 || h <= 0 {
		return
	}
	glowAlpha := float64(glowColor>>24) / 255.0
	r := uint8((glowColor >> 16) & 0xFF)
	g := uint8((glowColor >> 8) & 0xFF)
	b := uint8(glowColor & 0xFF)

	minX := max(0, x-radius)
	maxX := min(c.Width, x+w+radius)
	minY := max(0, y-radius)
	maxY := min(c.Height, y+h+radius)

	invRadius := 1.0 / float64(radius)

	for cy := minY; cy < maxY; cy++ {
		for cx := minX; cx < maxX; cx++ {
			// Find distance to the rectangle edges
			dx := 0
			if cx < x {
				dx = x - cx
			} else if cx >= x+w {
				dx = cx - (x + w - 1)
			}

			dy := 0
			if cy < y {
				dy = y - cy
			} else if cy >= y+h {
				dy = cy - (y + h - 1)
			}

			dist := math.Hypot(float64(dx), float64(dy))
			if dist > 0 && dist <= float64(radius) {
				factor := 1.0 - (dist * invRadius)
				// Quadratic falloff for natural light diffusion
				alpha := uint8(factor * factor * glowAlpha * 255.0)
				c.BlendPixel(cx, cy, ARGB(alpha, r, g, b))
			}
		}
	}
}

// DrawBoxShadow renders a soft drop-shadow with offset (offsetX, offsetY) and blur radius.
func (c *NativeCanvas) DrawBoxShadow(x, y, w, h, offsetX, offsetY, blurRadius int, shadowColor Color32) {
	shadowX := x + offsetX
	shadowY := y + offsetY
	c.DrawGlow(shadowX, shadowY, w, h, blurRadius, shadowColor)
}

// ApplyGaussianBlur applies a 2-pass separable box/gaussian approximation blur to a sub-region.
func (c *NativeCanvas) ApplyGaussianBlur(x, y, w, h, radius int) {
	if radius <= 0 || w <= 0 || h <= 0 {
		return
	}
	x0 := max(0, x)
	y0 := max(0, y)
	x1 := min(c.Width, x+w)
	y1 := min(c.Height, y+h)
	subW := x1 - x0
	subH := y1 - y0
	if subW <= 1 || subH <= 1 {
		return
	}

	temp := make([]uint32, subW*subH)

	// Horizontal Pass
	for row := 0; row < subH; row++ {
		origY := y0 + row
		for col := 0; col < subW; col++ {
			var rSum, gSum, bSum, count uint32
			for k := -radius; k <= radius; k++ {
				sampX := x0 + col + k
				if sampX >= x0 && sampX < x1 {
					pix := c.Pixels[origY*c.Width+sampX]
					rSum += (pix >> 16) & 0xFF
					gSum += (pix >> 8) & 0xFF
					bSum += pix & 0xFF
					count++
				}
			}
			if count > 0 {
				r := uint8(rSum / count)
				g := uint8(gSum / count)
				b := uint8(bSum / count)
				temp[row*subW+col] = (255 << 24) | (uint32(r) << 16) | (uint32(g) << 8) | uint32(b)
			}
		}
	}

	// Vertical Pass
	for col := 0; col < subW; col++ {
		destX := x0 + col
		for row := 0; row < subH; row++ {
			destY := y0 + row
			var rSum, gSum, bSum, count uint32
			for k := -radius; k <= radius; k++ {
				sampRow := row + k
				if sampRow >= 0 && sampRow < subH {
					pix := temp[sampRow*subW+col]
					rSum += (pix >> 16) & 0xFF
					gSum += (pix >> 8) & 0xFF
					bSum += pix & 0xFF
					count++
				}
			}
			if count > 0 {
				r := uint8(rSum / count)
				g := uint8(gSum / count)
				b := uint8(bSum / count)
				c.BlendPixel(destX, destY, ARGB(200, r, g, b))
			}
		}
	}
}
