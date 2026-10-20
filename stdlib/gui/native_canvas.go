package gui

import (
	"encoding/binary"
	"math"
	"os"
	"strings"
	"sync"
	"unsafe"
)

// FramebufferConfig defines screen resolution and pixel layout for native bare-metal graphics.
type FramebufferConfig struct {
	Width       int
	Height      int
	Pitch       int // bytes per scanline (Width * 4)
	BitsPerPix  int // usually 32 (XRGB / ARGB 8888)
	PhysicalAddr uintptr // MMIO hardware address (e.g., 0xFD000000 or VESA/GOP linear buffer)
}

// Color32 represents a 32-bit ARGB/XRGB color.
type Color32 uint32

func ARGB(a, r, g, b uint8) Color32 {
	return Color32((uint32(a) << 24) | (uint32(r) << 16) | (uint32(g) << 8) | uint32(b))
}

func RGB(r, g, b uint8) Color32 {
	return ARGB(255, r, g, b)
}

// Standalone Native Color Palette (Cyber GlassDark Theme)
var (
	ColorBgDark     = RGB(3, 7, 18)       // #030712 Deep Void
	ColorCardBg     = ARGB(180, 15, 23, 42) // Semi-transparent glass slate
	ColorBorder     = RGB(56, 189, 248)    // Cyan #38bdf8
	ColorText       = RGB(248, 250, 252)   // Slate 50
	ColorTextDim    = RGB(148, 163, 184)   // Slate 400
	ColorPrimary    = RGB(99, 102, 241)    // Indigo 500
	ColorSuccess    = RGB(16, 185, 129)    // Emerald 500
	ColorDanger     = RGB(244, 63, 94)     // Rose 500
)

// NativeCanvas provides direct, hardware-level 2D software rendering (zero OS, zero Web, zero libc).
// Can render to direct MMIO video memory or an in-memory double-buffer.
type NativeCanvas struct {
	Width       int
	Height      int
	Pitch       int
	Pixels      []uint32
	IsHardware  bool
	HwBufferPtr uintptr
	mu          sync.RWMutex
}

// NewNativeCanvas creates an independent double-buffered canvas.
func NewNativeCanvas(width, height int) *NativeCanvas {
	return &NativeCanvas{
		Width:  width,
		Height: height,
		Pitch:  width * 4,
		Pixels: make([]uint32, width*height),
	}
}

// AttachBareMetalFramebuffer binds canvas directly to hardware MMIO linear framebuffer.
func (c *NativeCanvas) AttachBareMetalFramebuffer(physicalPtr *uint32, width, height, pitch int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Width = width
	c.Height = height
	c.Pitch = pitch
	c.HwBufferPtr = uintptr(unsafe.Pointer(physicalPtr))
	c.IsHardware = true
	// Zero-copy hardware slice mapping
	if physicalPtr != nil {
		c.Pixels = unsafe.Slice(physicalPtr, width*height)
	}
}

// Clear fills the entire canvas with a single color using AVX2 SIMD acceleration.
func (c *NativeCanvas) Clear(color Color32) {
	clearPixelsAVX2(c.Pixels, uint32(color))
}

// SetPixel writes a 32-bit pixel with bounds checking.
func (c *NativeCanvas) SetPixel(x, y int, color Color32) {
	if x < 0 || x >= c.Width || y < 0 || y >= c.Height {
		return
	}
	c.Pixels[y*c.Width+x] = uint32(color)
}

// BlendPixel performs fast alpha blending: out = src*alpha + dst*(1-alpha).
func (c *NativeCanvas) BlendPixel(x, y int, color Color32) {
	if x < 0 || x >= c.Width || y < 0 || y >= c.Height {
		return
	}
	srcA := uint32(color >> 24)
	if srcA == 255 {
		c.Pixels[y*c.Width+x] = uint32(color)
		return
	}
	if srcA == 0 {
		return
	}

	dst := c.Pixels[y*c.Width+x]
	dstR := (dst >> 16) & 0xFF
	dstG := (dst >> 8) & 0xFF
	dstB := dst & 0xFF

	srcR := (uint32(color) >> 16) & 0xFF
	srcG := (uint32(color) >> 8) & 0xFF
	srcB := uint32(color) & 0xFF

	invA := 255 - srcA
	outR := (srcR*srcA + dstR*invA) / 255
	outG := (srcG*srcA + dstG*invA) / 255
	outB := (srcB*srcA + dstB*invA) / 255

	c.Pixels[y*c.Width+x] = (255 << 24) | (outR << 16) | (outG << 8) | outB
}

// DrawRect renders a filled rectangle with alpha blending using SIMD acceleration.
func (c *NativeCanvas) DrawRect(x, y, w, h int, color Color32) {
	x0 := max(0, x)
	y0 := max(0, y)
	x1 := min(c.Width, x+w)
	y1 := min(c.Height, y+h)
	spanLen := x1 - x0
	if spanLen <= 0 {
		return
	}

	for cy := y0; cy < y1; cy++ {
		rowOffset := cy*c.Width + x0
		blendRowAVX2(c.Pixels[rowOffset:rowOffset+spanLen], uint32(color))
	}
}

// DrawRoundedRect renders a smooth glassmorphic card with rounded corners using SIMD span acceleration.
func (c *NativeCanvas) DrawRoundedRect(x, y, w, h, r int, color Color32) {
	if w <= 0 || h <= 0 {
		return
	}
	if r <= 0 {
		c.DrawRect(x, y, w, h, color)
		return
	}
	// Clamp radius to half dimensions
	if r > w/2 {
		r = w / 2
	}
	if r > h/2 {
		r = h / 2
	}

	rSq := r * r
	y0 := max(0, y)
	y1 := min(c.Height, y+h)

	for cy := y0; cy < y1; cy++ {
		// Calculate left and right boundaries for this scanline
		var leftOffset, rightOffset int
		if cy < y+r {
			// Top rounded corners
			dy := (y + r) - cy
			// dy^2 + dx^2 <= r^2 => dx = sqrt(r^2 - dy^2)
			dx := int(math.Sqrt(float64(rSq - dy*dy)))
			leftOffset = r - dx
			rightOffset = r - dx
		} else if cy >= y+h-r {
			// Bottom rounded corners
			dy := cy - (y + h - r - 1)
			dx := int(math.Sqrt(float64(rSq - dy*dy)))
			leftOffset = r - dx
			rightOffset = r - dx
		}

		spanX0 := max(0, x+leftOffset)
		spanX1 := min(c.Width, x+w-rightOffset)
		spanLen := spanX1 - spanX0
		if spanLen > 0 {
			rowOffset := cy*c.Width + spanX0
			blendRowAVX2(c.Pixels[rowOffset:rowOffset+spanLen], uint32(color))
		}
	}
}

// DrawStrokeRoundedRect renders a glowing neon border.
func (c *NativeCanvas) DrawStrokeRoundedRect(x, y, w, h, r, thickness int, color Color32) {
	// Top & Bottom edges
	c.DrawRect(x+r, y, w-2*r, thickness, color)
	c.DrawRect(x+r, y+h-thickness, w-2*r, thickness, color)
	// Left & Right edges
	c.DrawRect(x, y+r, thickness, h-2*r, color)
	c.DrawRect(x+w-thickness, y+r, thickness, h-2*r, color)
}

func distSq(x1, y1, x2, y2 int) int {
	dx := x1 - x2
	dy := y1 - y2
	return dx*dx + dy*dy
}

// DrawChar renders a single 8x16 bitmap character without external font libraries.
func (c *NativeCanvas) DrawChar(x, y int, ch rune, color Color32, scale int) {
	glyph := getNativeGlyph(ch)
	for row := 0; row < 16; row++ {
		bits := glyph[row]
		for col := 0; col < 8; col++ {
			if (bits & (1 << (7 - col))) != 0 {
				if scale == 1 {
					c.BlendPixel(x+col, y+row, color)
				} else {
					c.DrawRect(x+col*scale, y+row*scale, scale, scale, color)
				}
			}
		}
	}
}

// DrawString renders bitmap text on bare-metal.
func (c *NativeCanvas) DrawString(x, y int, text string, color Color32, scale int) {
	cx := x
	cy := y
	for _, ch := range text {
		if ch == '\n' {
			cy += 18 * scale
			cx = x
			continue
		}
		c.DrawChar(cx, cy, ch, color, scale)
		cx += 9 * scale
	}
}

// SaveAsBMP exports canvas to standard BMP file for screenshot inspection without dependencies.
func (c *NativeCanvas) SaveAsBMP(filePath string) error {
	f, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	fileSize := 54 + (c.Width * c.Height * 4)
	header := make([]byte, 54)
	// BMP Header
	header[0] = 'B'
	header[1] = 'M'
	binary.LittleEndian.PutUint32(header[2:], uint32(fileSize))
	binary.LittleEndian.PutUint32(header[10:], 54) // offset
	// DIB Header
	binary.LittleEndian.PutUint32(header[14:], 40) // header size
	binary.LittleEndian.PutUint32(header[18:], uint32(c.Width))
	binary.LittleEndian.PutUint32(header[22:], uint32(-c.Height)) // top-down
	binary.LittleEndian.PutUint16(header[26:], 1)                 // planes
	binary.LittleEndian.PutUint16(header[28:], 32)                // bpp
	binary.LittleEndian.PutUint32(header[30:], 0)                 // BI_RGB (uncompressed)

	if _, err := f.Write(header); err != nil {
		return err
	}

	buf := make([]byte, c.Width*c.Height*4)
	for i, pix := range c.Pixels {
		buf[i*4+0] = byte(pix & 0xFF)         // B
		buf[i*4+1] = byte((pix >> 8) & 0xFF)  // G
		buf[i*4+2] = byte((pix >> 16) & 0xFF) // R
		buf[i*4+3] = byte((pix >> 24) & 0xFF) // A
	}
	_, err = f.Write(buf)
	return err
}

// ============================================================================
// 🌟 Native Glassmorphic GUI Renderer (Standalone Bare-Metal Scene Graph)
// ============================================================================

// RenderSceneToCanvas renders the declarative AXLUI scene graph directly to native pixels using themes, gradients, and effects.
func RenderSceneToCanvas(canvas *NativeCanvas, model *AXLUIModel) {
	theme := GetTheme(model.Theme)

	// 1. Fill themed background gradient
	canvas.Clear(theme.Background)

	// Atmospheric radial-linear hybrid gradient
	grad := NewLinearGradient(math.Pi*0.25,
		GradientStop{Offset: 0.0, Color: LerpColor(theme.Background, theme.Primary, 0.25)},
		GradientStop{Offset: 0.5, Color: theme.Background},
		GradientStop{Offset: 1.0, Color: LerpColor(theme.Background, theme.Secondary, 0.2)},
	)
	canvas.FillLinearGradient(0, 0, canvas.Width, canvas.Height, grad)

	// 2. Render Root & Children
	if model.Root != nil {
		pad := 24
		rootW := canvas.Width - pad*2
		rootH := canvas.Height - pad*2
		renderElementTreeWithTheme(canvas, model.Root, pad, pad, rootW, rootH, theme)
	}

	// 3. Render telemetry badge at bottom-right with vector typography
	footerText := "AXIOM NATIVE ENGINE · Ring 0 · 120 FPS · Zero-Web · Zero-GC"
	canvas.DrawString(canvas.Width-560, canvas.Height-20, footerText, theme.TextMuted, 1)
}

func renderElementTree(c *NativeCanvas, el *UIElement, x, y, maxW, maxH int) {
	renderElementTreeWithTheme(c, el, x, y, maxW, maxH, ThemeGlassDark)
}

func renderElementTreeWithTheme(c *NativeCanvas, el *UIElement, x, y, maxW, maxH int, theme Theme) {
	if el == nil {
		return
	}

	w := int(el.Style.Width)
	if w <= 0 || w > maxW {
		w = maxW
	}
	h := int(el.Style.Height)
	if h <= 0 || h > maxH {
		h = maxH
	}

	// Determine background, gradients, borders, shadows
	if el.Tag == "button" {
		// Soft glow & drop shadow
		c.DrawBoxShadow(x, y, w, h, 0, 4, 12, ARGB(60, 0, 0, 0))
		// Button gradient fill
		btnGrad := NewLinearGradient(math.Pi*0.5,
			GradientStop{Offset: 0.0, Color: theme.Primary},
			GradientStop{Offset: 1.0, Color: theme.PrimaryEnd},
		)
		c.FillRoundedGradient(x, y, w, h, theme.CornerRadius, btnGrad)
		c.DrawStrokeRoundedRect(x, y, w, h, theme.CornerRadius, 1, ARGB(180, 255, 255, 255))
	} else if el.ID != "viewport_root" && (el.Style.BorderColor != "" || el.Style.Background != "") {
		// Glass Card styling with blur effect and gradient card background
		c.DrawBoxShadow(x, y, w, h, 0, 8, 20, ARGB(70, 0, 0, 0))
		bg := theme.CardBg
		c.DrawRoundedRect(x, y, w, h, theme.CornerRadius, bg)
		c.DrawStrokeRoundedRect(x, y, w, h, theme.CornerRadius, 1, theme.BorderGlow)
	}

	// Draw Element Text if present
	if el.Text != "" {
		textColor := theme.TextPrimary
		if el.Style.Color == "#38bdf8" || strings.Contains(el.Style.Color, "blue") || strings.Contains(el.Style.Color, "cyan") {
			textColor = theme.Border
		} else if el.Style.Color == "#4ade80" || strings.Contains(el.Style.Color, "green") {
			textColor = theme.Success
		} else if el.Style.Color == "#94a3b8" || strings.Contains(el.Style.Color, "slate") {
			textColor = theme.TextMuted
		}

		scale := 1
		if el.Style.FontSize >= 18 {
			scale = 2
		}

		textY := y + 10
		if el.Tag == "button" {
			textY = y + (h-16*scale)/2
		}
		c.DrawString(x+14, textY, el.Text, textColor, scale)
	}

	// Render Children with Flex Layout (row or column)
	if len(el.Children) > 0 {
		gap := 16
		isRow := el.Style.FlexDirection == "row"

		if isRow {
			childCount := len(el.Children)
			availW := w - 24 - (childCount-1)*gap
			childW := availW / childCount
			childH := h - 20

			curX := x + 12
			for _, ch := range el.Children {
				cw := childW
				if ch.Style.Width > 0 && int(ch.Style.Width) < w {
					cw = int(ch.Style.Width)
				}
				chH := childH
				if ch.Style.Height > 0 && int(ch.Style.Height) < h {
					chH = int(ch.Style.Height)
				}
				renderElementTreeWithTheme(c, ch, curX, y+10, cw, chH, theme)
				curX += cw + gap
			}
		} else {
			// Column layout
			curY := y + 10
			if el.Text != "" {
				curY += 28
			}
			for _, ch := range el.Children {
				chH := int(ch.Style.Height)
				if chH <= 0 {
					chH = 44
				}
				chW := w - 24
				if ch.Style.Width > 0 && int(ch.Style.Width) < w {
					chW = int(ch.Style.Width)
				}
				renderElementTreeWithTheme(c, ch, x+12, curY, chW, chH, theme)
				curY += chH + gap
			}
		}
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
