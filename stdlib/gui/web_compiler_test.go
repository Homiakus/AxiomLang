package gui

import (
	"os"
	"testing"
)

func TestWebCompiler_ComplexDashboard(t *testing.T) {
	htmlPath := "../../examples/gui_spatial_dashboard.html"
	if _, err := os.Stat(htmlPath); err != nil {
		t.Skip("gui_spatial_dashboard.html not found, skipping test")
	}

	model, err := LoadHTMLFile(htmlPath)
	if err != nil {
		t.Fatalf("failed to compile web interface: %v", err)
	}

	if model == nil {
		t.Fatal("expected non-nil compiled model")
	}

	if len(model.Root.Children) == 0 {
		t.Fatal("expected compiled DOM tree elements, got 0")
	}

	// Verify rendering to native canvas
	canvas := NewNativeCanvas(1920, 1080)
	RenderSceneToCanvas(canvas, model)

	outBMP := "/tmp/test_web_compiled.bmp"
	defer os.Remove(outBMP)

	if err := canvas.SaveAsBMP(outBMP); err != nil {
		t.Fatalf("failed to save compiled web BMP: %v", err)
	}

	info, err := os.Stat(outBMP)
	if err != nil || info.Size() == 0 {
		t.Fatalf("compiled web BMP file is invalid or empty: %v", err)
	}
}
