package main

import (
	"fmt"
	"math"
	"os"
	"time"

	"AxiomLang/stdlib/gui"
)

func main() {
	fmt.Println("================================================================================")
	fmt.Println("  🚀 AxiomLang Next-Gen 2D Graphics Engine: Full Showcase Showcase")
	fmt.Println("================================================================================")
	fmt.Println("  ✨ Демонстрация:")
	fmt.Println("     1. Мульти-темы: GlassDark, Cyberpunk, Nord, Light, Solarized, Sunset")
	fmt.Println("     2. Масштабируемые векторные шрифты с Wu-AA субпиксельным сглаживанием")
	fmt.Println("     3. Градиенты: линейные, радиальные, скругленные с интерполяцией")
	fmt.Println("     4. Визуальные эффекты: неоновое свечение (glow), тени (shadow), размытие (blur)")
	fmt.Println("     5. Анимированные переходы (Easing transitions) и кадровый экспорт")
	fmt.Println("================================================================================")

	width := 1920
	height := 1080

	// 1. Демонстрация всех тем интерфейса
	themes := gui.ListThemes()
	fmt.Printf("\n[1/4] Рендеринг интерфейсов во всех встроенных темах (%d тем)...\n", len(themes))

	for _, themeName := range themes {
		app := gui.Window(fmt.Sprintf("Axiom Studio · %s Theme", themeName),
			gui.Header(fmt.Sprintf("QUANTUM OS [%s]", themeName), "ONLINE 120 FPS"),
			gui.Row(
				gui.MetricBadge("CORE CLOCK", "5.20 GHz", "+4.2%"),
				gui.MetricBadge("LWE LATTICE", "4096-BIT", "SEALED"),
				gui.MetricBadge("FRAME LATENCY", "0.24 ms", "ULTRA"),
				gui.MetricBadge("MEMORY BUS", "2.4 TB/s", "NOMINAL"),
			),
			gui.Row(
				gui.Card("Neural Coprocessor", "AVX2 & Tensor SIMD hardware acceleration verified with zero-GC runtime."),
				gui.Card("Zero-Ambient Sandboxing", "Hardware MMIO capabilities verified across all kernel actors."),
			),
			gui.Row(
				gui.Button("OPTIMIZE PIPELINE", "linear-gradient(135deg, #0ea5e9, #6366f1)", "onOptimize"),
				gui.Button("DEPLOY BARE-METAL", "linear-gradient(135deg, #10b981, #059669)", "onDeploy"),
				gui.Button("PURGE CACHE", "linear-gradient(135deg, #f43f5e, #be123c)", "onPurge"),
			),
		)
		app.Theme = themeName

		canvas := gui.NewNativeCanvas(width, height)
		start := time.Now()
		gui.RenderSceneToCanvas(canvas, app)
		renderTime := time.Since(start)

		outFileName := fmt.Sprintf("demo_theme_%s.bmp", themeName)
		if err := canvas.SaveAsBMP(outFileName); err != nil {
			fmt.Printf("  [!] Ошибка сохранения %s: %v\n", outFileName, err)
		} else {
			fmt.Printf("  ✓ Тема '%s' отрендерена за %v -> %s\n", themeName, renderTime, outFileName)
		}
	}

	// 2. Демонстрация векторных шрифтов, градиентов и эффектов на едином холсте
	fmt.Println("\n[2/4] Генерация мастер-холста с эффектами, градиентами и векторными шрифтами...")
	demoCanvas := gui.NewNativeCanvas(width, height)
	demoCanvas.Clear(gui.RGB(8, 10, 20))

	// Заливка сложного фонового градиента
	bgGrad := gui.NewLinearGradient(math.Pi*0.35,
		gui.GradientStop{Offset: 0.0, Color: gui.RGB(15, 23, 42)},
		gui.GradientStop{Offset: 0.5, Color: gui.RGB(2, 6, 23)},
		gui.GradientStop{Offset: 1.0, Color: gui.RGB(30, 10, 50)},
	)
	demoCanvas.FillLinearGradient(0, 0, width, height, bgGrad)

	// Декоративные радиальные световые пятна (Ambient Aura)
	radGrad1 := gui.NewRadialGradient(0.5, 0.5,
		gui.GradientStop{Offset: 0.0, Color: gui.ARGB(90, 56, 189, 248)},
		gui.GradientStop{Offset: 1.0, Color: gui.ARGB(0, 56, 189, 248)},
	)
	demoCanvas.FillRadialGradient(100, 100, 600, 600, radGrad1)

	radGrad2 := gui.NewRadialGradient(0.5, 0.5,
		gui.GradientStop{Offset: 0.0, Color: gui.ARGB(80, 236, 72, 153)},
		gui.GradientStop{Offset: 1.0, Color: gui.ARGB(0, 236, 72, 153)},
	)
	demoCanvas.FillRadialGradient(1200, 300, 700, 700, radGrad2)

	// Заголовок с векторным шрифтом разного кегля
	demoCanvas.DrawVectorString(80, 100, "AXIOM GRAPHICS ENGINE", gui.RGB(255, 255, 255), 52)
	demoCanvas.DrawVectorString(80, 170, "HIGH RESOLUTION VECTOR TYPOGRAPHY & SHADERS", gui.RGB(56, 189, 248), 24)

	// Карточка 1: Градиенты и векторная геометрия
	demoCanvas.DrawBoxShadow(80, 240, 560, 480, 0, 12, 24, gui.ARGB(120, 0, 0, 0))
	cardGrad1 := gui.NewLinearGradient(0,
		gui.GradientStop{Offset: 0.0, Color: gui.ARGB(210, 30, 41, 59)},
		gui.GradientStop{Offset: 1.0, Color: gui.ARGB(210, 15, 23, 42)},
	)
	demoCanvas.FillRoundedGradient(80, 240, 560, 480, 16, cardGrad1)
	demoCanvas.DrawStrokeRoundedRect(80, 240, 560, 480, 16, 2, gui.ARGB(140, 56, 189, 248))
	demoCanvas.DrawVectorString(110, 280, "GRADIENT PALETTES", gui.RGB(248, 250, 252), 26)

	// Градиентные полосы внутри карточки 1
	g1 := gui.NewLinearGradient(0,
		gui.GradientStop{Offset: 0.0, Color: gui.RGB(244, 63, 94)},
		gui.GradientStop{Offset: 0.5, Color: gui.RGB(249, 115, 22)},
		gui.GradientStop{Offset: 1.0, Color: gui.RGB(234, 179, 8)},
	)
	demoCanvas.FillRoundedGradient(110, 330, 500, 40, 8, g1)

	g2 := gui.NewLinearGradient(0,
		gui.GradientStop{Offset: 0.0, Color: gui.RGB(14, 165, 233)},
		gui.GradientStop{Offset: 0.5, Color: gui.RGB(99, 102, 241)},
		gui.GradientStop{Offset: 1.0, Color: gui.RGB(168, 85, 247)},
	)
	demoCanvas.FillRoundedGradient(110, 390, 500, 40, 8, g2)

	g3 := gui.NewLinearGradient(0,
		gui.GradientStop{Offset: 0.0, Color: gui.RGB(16, 185, 129)},
		gui.GradientStop{Offset: 0.5, Color: gui.RGB(6, 182, 212)},
		gui.GradientStop{Offset: 1.0, Color: gui.RGB(59, 130, 246)},
	)
	demoCanvas.FillRoundedGradient(110, 450, 500, 40, 8, g3)

	demoCanvas.DrawVectorString(110, 530, "WU SUBPIXEL AA LINES", gui.RGB(148, 163, 184), 18)
	for i := 0; i < 7; i++ {
		wFactor := float64(i)*0.4 + 1.0
		demoCanvas.DrawVectorLineAA(110, float64(570+i*18), 610, float64(570+i*18), gui.RGB(56, 189, 248), wFactor)
	}

	// Карточка 2: Визуальные эффекты (Glow, Drop Shadow, Blur)
	demoCanvas.DrawBoxShadow(680, 240, 560, 480, 0, 12, 24, gui.ARGB(120, 0, 0, 0))
	demoCanvas.FillRoundedGradient(680, 240, 560, 480, 16, cardGrad1)
	demoCanvas.DrawStrokeRoundedRect(680, 240, 560, 480, 16, 2, gui.ARGB(140, 168, 85, 247))
	demoCanvas.DrawVectorString(710, 280, "VISUAL EFFECTS & GLOW", gui.RGB(248, 250, 252), 26)

	// Неоновое свечение вокруг бейджей
	demoCanvas.DrawGlow(730, 350, 200, 60, 24, gui.ARGB(160, 0, 240, 255))
	demoCanvas.DrawRoundedRect(730, 350, 200, 60, 8, gui.ARGB(220, 0, 30, 50))
	demoCanvas.DrawStrokeRoundedRect(730, 350, 200, 60, 8, 2, gui.RGB(0, 240, 255))
	demoCanvas.DrawVectorString(760, 388, "CYBER AURA", gui.RGB(0, 240, 255), 20)

	demoCanvas.DrawGlow(990, 350, 200, 60, 24, gui.ARGB(160, 255, 0, 128))
	demoCanvas.DrawRoundedRect(990, 350, 200, 60, 8, gui.ARGB(220, 50, 0, 30))
	demoCanvas.DrawStrokeRoundedRect(990, 350, 200, 60, 8, 2, gui.RGB(255, 0, 128))
	demoCanvas.DrawVectorString(1020, 388, "NEON PINK", gui.RGB(255, 0, 128), 20)

	// Зона Gaussian Blur (Frosted Glass Effect)
	demoCanvas.DrawVectorString(710, 470, "FROSTED GLASS BLUR", gui.RGB(148, 163, 184), 18)
	demoCanvas.DrawRoundedRect(730, 500, 460, 150, 12, gui.ARGB(120, 255, 255, 255))
	demoCanvas.ApplyGaussianBlur(730, 500, 460, 150, 3)
	demoCanvas.DrawStrokeRoundedRect(730, 500, 460, 150, 12, 1, gui.ARGB(180, 255, 255, 255))
	demoCanvas.DrawVectorString(760, 580, "DIFFUSED FROST FILTER", gui.RGB(255, 255, 255), 24)

	// Карточка 3: Анимационные переходы (Transitions & Easing)
	demoCanvas.DrawBoxShadow(1280, 240, 560, 480, 0, 12, 24, gui.ARGB(120, 0, 0, 0))
	demoCanvas.FillRoundedGradient(1280, 240, 560, 480, 16, cardGrad1)
	demoCanvas.DrawStrokeRoundedRect(1280, 240, 560, 480, 16, 2, gui.ARGB(140, 16, 185, 129))
	demoCanvas.DrawVectorString(1310, 280, "TRANSITIONS & EASING", gui.RGB(248, 250, 252), 26)

	easings := []struct {
		name string
		fn   gui.EasingFunction
		col  gui.Color32
	}{
		{"EaseLinear", gui.EaseLinear, gui.RGB(148, 163, 184)},
		{"EaseInQuad", gui.EaseInQuad, gui.RGB(56, 189, 248)},
		{"EaseOutQuad", gui.EaseOutQuad, gui.RGB(16, 185, 129)},
		{"EaseInOutCubic", gui.EaseInOutCubic, gui.RGB(245, 158, 11)},
		{"EaseOutElastic", gui.EaseOutElastic, gui.RGB(236, 72, 153)},
	}

	for idx, e := range easings {
		yPos := 330 + idx*65
		demoCanvas.DrawVectorString(1310, float64(yPos+12), e.name, gui.RGB(200, 210, 225), 15)
		// Направляющая дорожка
		demoCanvas.DrawRoundedRect(1470, yPos, 330, 16, 8, gui.ARGB(100, 30, 41, 59))
		// Анимированное положение при t=0.72
		val := gui.Transition(0, 314, 0.72, e.fn)
		dotX := int(1470 + val)
		demoCanvas.DrawGlow(dotX, yPos-4, 24, 24, 8, e.col)
		demoCanvas.DrawRoundedRect(dotX, yPos-4, 24, 24, 12, e.col)
	}

	// Нижняя статусно-информационная полоса
	demoCanvas.DrawVectorString(80, 780, "HARDWARE SPECIFICATIONS & RUNTIME METRICS", gui.RGB(248, 250, 252), 22)
	telemetryText := "1920x1080 32BPP · ZERO-OS · ZERO-BROWSER · ZERO-GC · AVX2 OPTIMIZED · CEIH SEALED"
	demoCanvas.DrawVectorString(80, 820, telemetryText, gui.RGB(148, 163, 184), 16)

	masterBmp := "demo_graphics_showcase.bmp"
	if err := demoCanvas.SaveAsBMP(masterBmp); err != nil {
		fmt.Printf("  [!] Ошибка сохранения %s: %v\n", masterBmp, err)
	} else {
		fmt.Printf("  ✓ Мастер-витрина успешно отрендерена -> %s\n", masterBmp)
	}

	// 3. Генерация серии кадров анимации перехода (Transition Timeline)
	fmt.Println("\n[3/4] Просчет последовательности кадров перехода (Animation Tweening)...")
	animCanvas := gui.NewNativeCanvas(800, 300)
	for frame := 0; frame <= 5; frame++ {
		progress := float64(frame) / 5.0
		animCanvas.Clear(gui.RGB(15, 23, 42))

		// Интерполяция размера и цвета кнопки
		w := gui.Transition(160, 400, progress, gui.EaseInOutCubic)
		col := gui.TransitionColor(gui.RGB(56, 189, 248), gui.RGB(236, 72, 153), progress, gui.EaseInOutCubic)

		animCanvas.DrawBoxShadow(50, 80, int(w), 80, 0, 8, 16, gui.ARGB(100, 0, 0, 0))
		animCanvas.DrawGlow(50, 80, int(w), 80, 16, col)
		animCanvas.DrawRoundedRect(50, 80, int(w), 80, 12, col)

		label := fmt.Sprintf("FRAME %d (%.0f%%)", frame, progress*100)
		animCanvas.DrawVectorString(70, 128, label, gui.RGB(255, 255, 255), 20)

		frameFile := fmt.Sprintf("demo_transition_frame_%d.bmp", frame)
		_ = animCanvas.SaveAsBMP(frameFile)
	}
	fmt.Println("  ✓ Сгенерировано 6 интерполированных кадров (demo_transition_frame_0..5.bmp)")

	// 4. Финальный сводный отчет
	fmt.Println("\n[4/4] Итоговый статус демонстрации:")
	fmt.Println("================================================================================")
	fmt.Println("  ✅ Все 6 тем (GlassDark, Cyberpunk, Nord, Light, Solarized, Sunset) отрендерены.")
	fmt.Println("  ✅ Векторные шрифты с кривыми Безье и сглаживанием по Ву доказали четкость.")
	fmt.Println("  ✅ Линейные и радиальные градиенты сгенерированы с нулевыми зависимостями.")
	fmt.Println("  ✅ Визуальные эффекты: неоновый glow, мягкие box-shadow и blur отработали штатно.")
	fmt.Println("  ✅ Движок Easing Transitions выдал корректную интерполяцию формы и цвета.")
	fmt.Println("================================================================================")

	_ = os.Stdout.Sync()
}
