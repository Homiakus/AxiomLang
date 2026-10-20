//go:build !amd64

package lexer

func findNonWhitespaceAVX2(s string, offset int) int {
	for i := offset; i < len(s); i++ {
		ch := s[i]
		if ch != ' ' && ch != '\t' && ch != '\r' && ch != '\n' {
			return i
		}
	}
	return len(s)
}

func fastCRC32String(s string) uint32 {
	var crc uint32 = 0xFFFFFFFF
	for i := 0; i < len(s); i++ {
		crc ^= uint32(s[i])
		for j := 0; j < 8; j++ {
			if (crc & 1) != 0 {
				crc = (crc >> 1) ^ 0xEDB88320
			} else {
				crc >>= 1
			}
		}
	}
	return ^crc
}
