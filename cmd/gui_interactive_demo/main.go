package main

/*
#cgo LDFLAGS: -lX11
#include <X11/Xlib.h>
#include <X11/Xutil.h>
#include <stdlib.h>
#include <string.h>

static inline void put_pixel_buffer(Display *d, Window w, GC gc, XImage *img, int width, int height) {
    XPutImage(d, w, gc, img, 0, 0, 0, 0, width, height);
    XFlush(d);
}
*/
import "C"
import (
	"fmt"
	"math"
	"os"
	"time"
	"unsafe"

	"AxiomLang/stdlib/gui"
)

func main() {
	fmt.Println("================================================================================")
	fmt.Println("  🚀 AxiomLang Realtime Interactive GUI Engine (120 FPS / Native Window)")
	fmt.Println("================================================================================")
	fmt.Println("  Управление:")
	fmt.Println("    [SPACE]   - Переключить тему (GlassDark -> Cyberpunk -> Nord -> Light -> Solarized -> Sunset)")
	fmt.Println("    [E]       - Вкл/выкл неоновый Glow эффект")
	fmt.Println("    [B]       - Вкл/выкл Frosted Glass размытие (Gaussian Blur)")
	fmt.Println("    [Q / ESC] - Выход из программы")
	fmt.Println("================================================================================")

	width := 1280
	height := 720

	// 1. Инициализация нативного X11 окна
	display := C.XOpenDisplay(nil)
	if display == nil {
		fmt.Println("[FATAL] Не удалось открыть соединение с X11 сервером ($DISPLAY).")
		os.Exit(1)
	}
	defer C.XCloseDisplay(display)

	screen := C.XDefaultScreen(display)
	rootWindow := C.XRootWindow(display, screen)

	window := C.XCreateSimpleWindow(
		display,
		rootWindow,
		100, 100,
		C.uint(width), C.uint(height),
		1,
		C.XBlackPixel(display, screen),
		C.XBlackPixel(display, screen),
	)

	titleStr := C.CString("AxiomLang Next-Gen Graphics Showcase · Realtime Interactive Window")
	defer C.free(unsafe.Pointer(titleStr))
	C.XStoreName(display, window, titleStr)

	// Подписываемся на события клавиатуры и закрытия окна
	C.XSelectInput(display, window, C.ExposureMask|C.KeyPressMask|C.StructureNotifyMask)
	C.XMapWindow(display, window)

	gc := C.XDefaultGC(display, screen)
	visual := C.XDefaultVisual(display, screen)
	depth := C.XDefaultDepth(display, screen)

	// Создаем double-buffer холст NativeCanvas
	canvas := gui.NewNativeCanvas(width, height)

	// Создаем XImage, разделяющий пиксельный буфер напрямую с NativeCanvas
	ximg := C.XCreateImage(
		display,
		visual,
		C.uint(depth),
		C.ZPixmap,
		0,
		(*C.char)(unsafe.Pointer(&canvas.Pixels[0])),
		C.uint(width),
		C.uint(height),
		32,
		C.int(width*4),
	)
	if ximg == nil {
		fmt.Println("[FATAL] Не удалось создать аппаратный XImage буфер.")
		os.Exit(1)
	}

	themeNames := []string{"GlassDark", "Cyberpunk", "Nord", "Light", "Solarized", "Sunset"}
	currentThemeIdx := 0
	enableGlow := true
	enableBlur := true

	startTime := time.Now()
	ticker := time.NewTicker(8 * time.Millisecond) // ~120 FPS
	defer ticker.Stop()

	running := true

	for running {
		// Обработка очереди событий X11
		for C.XPending(display) > 0 {
			var event C.XEvent
			C.XNextEvent(display, &event)
			evType := *(*C.int)(unsafe.Pointer(&event))

			switch evType {
			case C.KeyPress:
				xkey := (*C.XKeyEvent)(unsafe.Pointer(&event))
				keysym := C.XLookupKeysym(xkey, 0)
				switch keysym {
				case 0x20: // SPACE
					currentThemeIdx = (currentThemeIdx + 1) % len(themeNames)
					fmt.Printf(" [UI] Смена темы: %s\n", themeNames[currentThemeIdx])
				case 0x65, 0x45: // 'e' / 'E'
					enableGlow = !enableGlow
					fmt.Printf(" [FX] Неоновый Glow: %v\n", enableGlow)
				case 0x62, 0x42: // 'b' / 'B'
					enableBlur = !enableBlur
					fmt.Printf(" [FX] Frosted Glass Blur: %v\n", enableBlur)
				case 0x71, 0x51, 0xff1b: // 'q' / 'Q' / ESC
					fmt.Println(" [UI] Закрытие графического окна...")
					running = false
				}
			case C.DestroyNotify:
				running = false
			}
		}

		if !running {
			break
		}

		// Вычисляем анимационное время
		elapsed := time.Since(startTime).Seconds()
		activeTheme := gui.GetTheme(themeNames[currentThemeIdx])

		// 1. Очистка и сложный динамический градиентный фон
		canvas.Clear(activeTheme.Background)
		gradAngle := elapsed * 0.4
		bgGrad := gui.NewLinearGradient(gradAngle,
			gui.GradientStop{Offset: 0.0, Color: activeTheme.Background},
			gui.GradientStop{Offset: 0.5, Color: gui.LerpColor(activeTheme.Background, activeTheme.Primary, 0.35)},
			gui.GradientStop{Offset: 1.0, Color: gui.LerpColor(activeTheme.Background, activeTheme.Secondary, 0.25)},
		)
		canvas.FillLinearGradient(0, 0, width, height, bgGrad)

		// 2. Декоративные вращающиеся радиальные ауры
		auraX1 := 200 + math.Cos(elapsed*1.2)*80
		auraY1 := 160 + math.Sin(elapsed*1.2)*60
		rad1 := gui.NewRadialGradient(0.5, 0.5,
			gui.GradientStop{Offset: 0.0, Color: activeTheme.BorderGlow},
			gui.GradientStop{Offset: 1.0, Color: gui.ARGB(0, 0, 0, 0)},
		)
		canvas.FillRadialGradient(int(auraX1), int(auraY1), 400, 400, rad1)

		// 3. Заголовок векторным шрифтом (Vector Typography)
		titleText := fmt.Sprintf("AXIOM ENGINE · %s", activeTheme.Name)
		canvas.DrawVectorString(60, 60, titleText, activeTheme.TextPrimary, 38)
		subtitleText := "120 FPS REALTIME · VECTOR FONTS · SHADERS · EASING TRANSITIONS · ZERO-WEB"
		canvas.DrawVectorString(60, 110, subtitleText, activeTheme.Border, 16)

		// 4. Левая интерактивная карточка (Градиенты + Векторная графика)
		card1X, card1Y, card1W, card1H := 60, 150, 560, 480
		canvas.DrawBoxShadow(card1X, card1Y, card1W, card1H, 0, 8, 20, gui.ARGB(120, 0, 0, 0))
		canvas.DrawRoundedRect(card1X, card1Y, card1W, card1H, activeTheme.CornerRadius, activeTheme.CardBg)
		canvas.DrawStrokeRoundedRect(card1X, card1Y, card1W, card1H, activeTheme.CornerRadius, 1, activeTheme.BorderGlow)

		canvas.DrawVectorString(float64(card1X+24), float64(card1Y+30), "MULTI-STOP GRADIENTS & VECTOR LINES", activeTheme.TextPrimary, 20)

		// Живые градиентные полосы
		pulseGrad1 := gui.NewLinearGradient(elapsed,
			gui.GradientStop{Offset: 0.0, Color: activeTheme.Primary},
			gui.GradientStop{Offset: 0.5, Color: activeTheme.Accent},
			gui.GradientStop{Offset: 1.0, Color: activeTheme.PrimaryEnd},
		)
		canvas.FillRoundedGradient(card1X+24, card1Y+70, card1W-48, 36, 8, pulseGrad1)

		pulseGrad2 := gui.NewLinearGradient(-elapsed*0.8,
			gui.GradientStop{Offset: 0.0, Color: activeTheme.Secondary},
			gui.GradientStop{Offset: 0.5, Color: activeTheme.Success},
			gui.GradientStop{Offset: 1.0, Color: activeTheme.Warning},
		)
		canvas.FillRoundedGradient(card1X+24, card1Y+120, card1W-48, 36, 8, pulseGrad2)

		// Векторные субпиксельные линии Безье (Wu-AA)
		canvas.DrawVectorString(float64(card1X+24), float64(card1Y+190), "DYNAMIC BEZIER CURVES & WU-AA", activeTheme.TextMuted, 16)
		p0 := gui.Point{X: float64(card1X + 30), Y: float64(card1Y + 340)}
		p1 := gui.Point{X: float64(card1X + 180), Y: float64(card1Y+220) + math.Sin(elapsed*3)*70}
		p2 := gui.Point{X: float64(card1X + 360), Y: float64(card1Y+420) + math.Cos(elapsed*3)*70}
		p3 := gui.Point{X: float64(card1X + 530), Y: float64(card1Y + 280)}
		canvas.DrawCubicBezierAA(p0, p1, p2, p3, activeTheme.Primary, 3.0)
		canvas.DrawCubicBezierAA(p0, p2, p1, p3, activeTheme.Accent, 1.8)

		// 5. Правая интерактивная карточка (Эффекты, Glow, Blur, Transitions)
		card2X, card2Y, card2W, card2H := 660, 150, 560, 480
		canvas.DrawBoxShadow(card2X, card2Y, card2W, card2H, 0, 8, 20, gui.ARGB(120, 0, 0, 0))
		canvas.DrawRoundedRect(card2X, card2Y, card2W, card2H, activeTheme.CornerRadius, activeTheme.CardBg)
		canvas.DrawStrokeRoundedRect(card2X, card2Y, card2W, card2H, activeTheme.CornerRadius, 1, activeTheme.BorderGlow)

		canvas.DrawVectorString(float64(card2X+24), float64(card2Y+30), "VISUAL EFFECTS & EASING TRANSITIONS", activeTheme.TextPrimary, 20)

		// Пульсирующая кнопка с Glow эффектом
		btnW := int(gui.Transition(220, 480, (math.Sin(elapsed*2)+1)*0.5, gui.EaseInOutCubic))
		btnCol := gui.TransitionColor(activeTheme.Primary, activeTheme.Accent, (math.Sin(elapsed*2)+1)*0.5, gui.EaseInOutCubic)

		if enableGlow {
			canvas.DrawGlow(card2X+30, card2Y+75, btnW, 50, 18, btnCol)
		}
		canvas.DrawRoundedRect(card2X+30, card2Y+75, btnW, 50, 10, btnCol)
		canvas.DrawStrokeRoundedRect(card2X+30, card2Y+75, btnW, 50, 10, 1, gui.ARGB(200, 255, 255, 255))
		canvas.DrawVectorString(float64(card2X+50), float64(card2Y+108), "PULSING NEON BUTTON", gui.RGB(255, 255, 255), 18)

		// Область Frosted Glass с динамическим размытием
		blurW, blurH := 500, 120
		canvas.DrawVectorString(float64(card2X+24), float64(card2Y+160), "FROSTED GLASS PANEL [B to toggle]", activeTheme.TextMuted, 16)
		canvas.DrawRoundedRect(card2X+30, card2Y+185, blurW, blurH, 12, gui.ARGB(140, 255, 255, 255))
		if enableBlur {
			canvas.ApplyGaussianBlur(card2X+30, card2Y+185, blurW, blurH, 2)
		}
		canvas.DrawStrokeRoundedRect(card2X+30, card2Y+185, blurW, blurH, 12, 1, gui.ARGB(200, 255, 255, 255))
		canvas.DrawVectorString(float64(card2X+60), float64(card2Y+250), "REALTIME GAUSSIAN SEPARABLE BLUR", activeTheme.TextPrimary, 18)

		// Анимированные дорожки функций плавности
		canvas.DrawVectorString(float64(card2X+24), float64(card2Y+335), "ANIMATED EASING TIMELINES", activeTheme.TextMuted, 16)
		easings := []struct {
			name string
			fn   gui.EasingFunction
			col  gui.Color32
		}{
			{"EaseLinear", gui.EaseLinear, activeTheme.TextMuted},
			{"EaseInOutCubic", gui.EaseInOutCubic, activeTheme.Primary},
			{"EaseOutElastic", gui.EaseOutElastic, activeTheme.Secondary},
		}

		loopProgress := math.Mod(elapsed, 2.0) / 2.0
		for idx, es := range easings {
			trackY := card2Y + 360 + idx*35
			canvas.DrawVectorString(float64(card2X+30), float64(trackY+12), es.name, activeTheme.TextPrimary, 14)
			canvas.DrawRoundedRect(card2X+170, trackY, 320, 14, 7, gui.ARGB(120, 20, 30, 45))

			pos := gui.Transition(0, 306, loopProgress, es.fn)
			ballX := card2X + 170 + int(pos)
			if enableGlow {
				canvas.DrawGlow(ballX, trackY-3, 20, 20, 6, es.col)
			}
			canvas.DrawRoundedRect(ballX, trackY-3, 20, 20, 10, es.col)
		}

		// Нижняя панель подсказок
		helpText := "[SPACE] Переключить тему  |  [E] Вкл/выкл Glow  |  [B] Вкл/выкл Blur  |  [Q/ESC] Выход"
		canvas.DrawString(60, height-25, helpText, activeTheme.TextMuted, 1)

		// 6. Вывод кадра в аппаратное X11 окно
		C.put_pixel_buffer(display, window, gc, ximg, C.int(width), C.int(height))

		<-ticker.C
	}

	fmt.Println("[OK] Графическое приложение успешно завершило работу.")
}
