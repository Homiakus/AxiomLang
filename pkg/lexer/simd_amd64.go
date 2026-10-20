//go:build amd64

package lexer

// findNonWhitespaceAVX2 scans input string starting from offset and returns the index
// of the first byte that is NOT a whitespace character (' ', '\t', '\r', '\n').
// Defined in scan_amd64.s
//
//go:noescape
func findNonWhitespaceAVX2(s string, offset int) int

// fastCRC32String calculates hardware accelerated CRC32-C hash of string using SSE4.2 crc32q/crc32b.
// Defined in scan_amd64.s
//
//go:noescape
func fastCRC32String(s string) uint32
