//go:build amd64

package data

//go:noescape
func sumInt64Asm(data []int64) int64
