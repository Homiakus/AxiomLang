//go:build !amd64

package data

func sumInt64Asm(data []int64) int64 {
	var sum int64
	for _, v := range data {
		sum += v
	}
	return sum
}
