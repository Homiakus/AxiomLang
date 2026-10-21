package math

import "math"

// Sin calculates sine of angle in radians
func Sin(x float64) float64 {
	return math.Sin(x)
}

// Cos calculates cosine of angle in radians
func Cos(x float64) float64 {
	return math.Cos(x)
}

// Sqrt returns the square root of x
func Sqrt(x float64) float64 {
	return math.Sqrt(x)
}

// Clamp restricts value between min and max
func Clamp(val, minVal, maxVal float64) float64 {
	if val < minVal {
		return minVal
	}
	if val > maxVal {
		return maxVal
	}
	return val
}

// Abs returns absolute value
func Abs(x float64) float64 {
	return math.Abs(x)
}

// Min returns smaller of two float values
func Min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

// Max returns larger of two float values
func Max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
