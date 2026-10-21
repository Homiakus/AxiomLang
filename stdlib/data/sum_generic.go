//go:build !amd64

package data

func sumInt64Asm(data []int64) int64 {
	var sum int64
	for _, v := range data {
		sum += v
	}
	return sum
}

func minMaxFloat64Asm(data []float64) (float64, float64) {
	if len(data) == 0 {
		return 0, 0
	}
	minVal, maxVal := data[0], data[0]
	for _, v := range data[1:] {
		if v < minVal {
			minVal = v
		}
		if v > maxVal {
			maxVal = v
		}
	}
	return minVal, maxVal
}
