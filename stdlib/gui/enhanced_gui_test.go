package gui

import (
	"os"
	"testing"
)

func TestThemes(t *testing.T) {
	themes := ListThemes()
	if len(themes) < 6 {
		t.Fatalf("expected at least 6 themes, got %d", len(themes))
	}

	for _, name := range []string{"GlassDark", "Cyberpunk", "Nord", "Light", "Solarized", "Sunset"} {
		th := GetTheme(name)
		if th.Name != name {
			t.Errorf("expected theme %s, got %s", name, th.Name)
		}
	}

	// Fallback check
	fallback := GetTheme("NonExistentThemeXYZ")
	if fallback.Name != "GlassDark" {
		t.Errorf("expected fallback to GlassDark, got %s", fallback.Name)
	}
}

func TestGradients(t *testing.T) {
	c := NewNativeCanvas(200, 200)

	// Linear gradient
	linGrad := NewLinearGradient(0,
		GradientStop{Offset: 0.0, Color: RGB(255, 0, 0)},
		GradientStop{Offset: 1.0, Color: RGB(0, 0, 255)},
	)
	c.FillLinearGradient(10, 10, 80, 80, linGrad)

	// Radial gradient
	radGrad := NewRadialGradient(0.5, 0.5,
		GradientStop{Offset: 0.0, Color: RGB(0, 255, 0)},
		GradientStop{Offset: 1.0, Color: RGB(0, 0, 0)},
	)
	c.FillRadialGradient(100, 10, 80, 80, radGrad)

	// Rounded gradient
	c.FillRoundedGradient(10, 100, 180, 80, 12, linGrad)

	// Sample check
	sampleMid := linGrad.Sample(0.5)
	if sampleMid == 0 {
		t.Error("expected non-zero interpolated gradient color")
	}
}

func TestVectorFont(t *testing.T) {
	c := NewNativeCanvas(400, 100)
	c.Clear(RGB(10, 10, 20))

	// Draw antialiased vector string
	c.DrawVectorString(20, 30, "AXIOM 2026", RGB(0, 240, 255), 32)

	// Check that pixels were actually blended
	hasRenderedPixels := false
	for _, pix := range c.Pixels {
		if pix != uint32(RGB(10, 10, 20)) {
			hasRenderedPixels = true
			break
		}
	}
	if !hasRenderedPixels {
		t.Error("expected rendered pixels from DrawVectorString")
	}
}

func TestEffectsAndTransitions(t *testing.T) {
	c := NewNativeCanvas(300, 200)
	c.Clear(RGB(5, 5, 10))

	// Drop shadow and glow
	c.DrawBoxShadow(50, 50, 100, 60, 0, 4, 10, ARGB(120, 0, 0, 0))
	c.DrawGlow(50, 50, 100, 60, 8, ARGB(150, 56, 189, 248))

	// Gaussian blur
	c.ApplyGaussianBlur(40, 40, 120, 80, 2)

	// Easing transitions
	p0 := Transition(0, 100, 0.5, EaseInOutCubic)
	if p0 != 50.0 {
		t.Errorf("expected 50.0 for EaseInOutCubic at 0.5, got %f", p0)
	}

	colTrans := TransitionColor(RGB(0, 0, 0), RGB(200, 100, 50), 0.5, EaseLinear)
	r := (colTrans >> 16) & 0xFF
	if r != 100 {
		t.Errorf("expected r=100, got %d", r)
	}
}

func TestSceneRendering_AllThemes(t *testing.T) {
	themes := []string{"GlassDark", "Cyberpunk", "Nord", "Light", "Solarized", "Sunset"}
	for _, th := range themes {
		app := Window("Axiom Precision Control",
			Header("System Core", "ONLINE"),
			Row(
				MetricBadge("TELEMETRY", "99.98%", "+0.02%"),
				MetricBadge("FREQUENCY", "4.80 GHz", "NOMINAL"),
			),
			Button("DEPLOY HYDROPILOT", "linear-gradient(135deg, #0ea5e9, #6366f1)", "onDeployClick"),
		)
		app.Theme = th

		c := NewNativeCanvas(800, 600)
		RenderSceneToCanvas(c, app)

		tmpFile := "/tmp/test_theme_" + th + ".bmp"
		err := c.SaveAsBMP(tmpFile)
		if err != nil {
			t.Fatalf("failed to save theme %s BMP: %v", th, err)
		}
		_ = os.Remove(tmpFile)
	}
}
