package gui

import (
	"testing"
)

func BenchmarkCanvas_Clear_ScalarVsAVX2(b *testing.B) {
	canvas := NewNativeCanvas(1920, 1080)
	color := Color32(0xFF1E1E2E)
	b.ResetTimer()
	b.SetBytes(int64(1920 * 1080 * 4))

	for i := 0; i < b.N; i++ {
		canvas.Clear(color)
	}
}

func BenchmarkCanvas_RenderFullScene(b *testing.B) {
	model := NewAXLUIModel("bench.axlui", "BenchmarkUI")
	model.Theme = "cyber-dark"
	canvas := NewNativeCanvas(1920, 1080)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		RenderSceneToCanvas(canvas, model)
	}
}
