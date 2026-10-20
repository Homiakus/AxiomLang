package lexer

import "testing"

func BenchmarkLexer_FullProgram(b *testing.B) {
	code := `
требовать Capability[StdIO]
требовать Capability[FileSystem.Read]

структура Config {
    порт: int,
    хост: string,
    активен: bool,
}

намерение StartServer {
    пусть порт = 8080
    пусть изм счетчик = 0
    для i в [1, 2, 3, 4, 5] {
        если i > 2 {
            счетчик = счетчик + i * 2
        } иначе {
            счетчик = счетчик + 1
        }
    }
    вернуть счетчик
}
`
	b.ResetTimer()
	b.SetBytes(int64(len(code)))

	for i := 0; i < b.N; i++ {
		l := New(code)
		for {
			tok := l.NextToken()
			if tok.Type == EOF {
				break
			}
		}
	}
}

func BenchmarkFastCRC32_Hardware(b *testing.B) {
	str := "требовать_Capability_FileSystem_Read_Write_Execute_LongIdentifier"
	b.ResetTimer()
	b.SetBytes(int64(len(str)))

	for i := 0; i < b.N; i++ {
		_ = fastCRC32String(str)
	}
}
