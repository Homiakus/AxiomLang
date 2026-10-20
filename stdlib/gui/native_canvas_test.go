package gui_test

import (
	"os"
	"testing"

	"AxiomLang/stdlib/gui"
)

func TestNativeCanvas_DirectRendering(t *testing.T) {
	width := 800
	height := 600
	canvas := gui.NewNativeCanvas(width, height)

	// 1. Verify canvas initialization
	if len(canvas.Pixels) != width*height {
		t.Fatalf("expected %d pixels, got %d", width*height, len(canvas.Pixels))
	}

	// 2. Clear background to dark void
	canvas.Clear(gui.ColorBgDark)
	if canvas.Pixels[0] != uint32(gui.ColorBgDark) {
		t.Fatalf("expected pixel[0] to be ColorBgDark (%x), got %x", gui.ColorBgDark, canvas.Pixels[0])
	}

	// 3. Draw direct primitive rectangle
	canvas.DrawRect(50, 50, 200, 100, gui.ColorPrimary)
	if canvas.Pixels[60*width+60] != uint32(gui.ColorPrimary) {
		t.Fatalf("expected pixel inside rect to match ColorPrimary")
	}

	// 4. Draw rounded glass card
	canvas.DrawRoundedRect(300, 50, 400, 200, 16, gui.ColorCardBg)

	// 5. Draw bitmap text without OS/font dependencies
	canvas.DrawString(320, 80, "AXIOM BARE-METAL KERNEL", gui.ColorText, 1)

	// 6. Test BMP Export
	tmpBMP := "test_native_framebuffer.bmp"
	err := canvas.SaveAsBMP(tmpBMP)
	if err != nil {
		t.Fatalf("failed to save native BMP: %v", err)
	}
	defer os.Remove(tmpBMP)

	fi, err := os.Stat(tmpBMP)
	if err != nil || fi.Size() < 1000 {
		t.Fatalf("BMP file is invalid or too small: %v", err)
	}
}

func TestNativeCanvas_RenderAXLUIModel(t *testing.T) {
	canvas := gui.NewNativeCanvas(1280, 720)

	app := gui.Window("Axiom Quantum OS",
		gui.Header("🛡️ Quantum OS Monitor", "120 FPS BARE-METAL"),
		gui.Row(
			gui.MetricBadge("CPU Core 0", "99.1%", "+1.2%"),
			gui.MetricBadge("Framebuffer", "1920x1080", "DIRECT-MMIO"),
		),
		gui.Card("Hardware Telemetry", "Прямой рендеринг в видеопамять без ОС и веб-серверов."),
	)

	// Direct scene graph render to memory buffer
	gui.RenderSceneToCanvas(canvas, app)

	// Verify top-left has cosmic background
	if canvas.Pixels[0] == 0 {
		t.Fatalf("expected non-zero cosmic pixel in rendered scene")
	}
}
