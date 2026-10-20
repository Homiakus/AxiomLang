package gui

import (
	"strings"
)

// Theme represents a visual design system with coherent colors, glass effects, borders, and typography.
type Theme struct {
	Name         string  `json:"name"`
	Background   Color32 `json:"background"`
	CardBg       Color32 `json:"card_bg"`
	Border       Color32 `json:"border"`
	BorderGlow   Color32 `json:"border_glow"`
	TextPrimary  Color32 `json:"text_primary"`
	TextMuted    Color32 `json:"text_muted"`
	Primary      Color32 `json:"primary"`
	PrimaryEnd   Color32 `json:"primary_end"`
	Secondary    Color32 `json:"secondary"`
	Success      Color32 `json:"success"`
	Warning      Color32 `json:"warning"`
	Danger       Color32 `json:"danger"`
	Accent       Color32 `json:"accent"`
	BackdropBlur float64 `json:"backdrop_blur"`
	CornerRadius int     `json:"corner_radius"`
}

var (
	// ThemeGlassDark is the default deep cyberpunk / space glass theme.
	ThemeGlassDark = Theme{
		Name:         "GlassDark",
		Background:   RGB(3, 7, 18),           // #030712 Deep Void
		CardBg:       ARGB(190, 15, 23, 42),   // Semi-transparent glass slate
		Border:       RGB(56, 189, 248),       // Cyan #38bdf8
		BorderGlow:   ARGB(100, 56, 189, 248), // Glowing cyan
		TextPrimary:  RGB(248, 250, 252),      // Slate 50
		TextMuted:    RGB(148, 163, 184),      // Slate 400
		Primary:      RGB(99, 102, 241),       // Indigo 500
		PrimaryEnd:   RGB(14, 165, 233),       // Sky 500
		Secondary:    RGB(139, 92, 246),       // Purple 500
		Success:      RGB(16, 185, 129),       // Emerald 500
		Warning:      RGB(245, 158, 11),       // Amber 500
		Danger:       RGB(244, 63, 94),        // Rose 500
		Accent:       RGB(236, 72, 153),       // Pink 500
		BackdropBlur: 24,
		CornerRadius: 12,
	}

	// ThemeCyberpunk is a high-contrast neon theme.
	ThemeCyberpunk = Theme{
		Name:         "Cyberpunk",
		Background:   RGB(10, 10, 15),         // Dark neon matrix
		CardBg:       ARGB(210, 24, 18, 43),   // Deep neon purple slate
		Border:       RGB(255, 0, 128),        // Neon Pink
		BorderGlow:   ARGB(160, 0, 240, 255),  // Neon Cyan glow
		TextPrimary:  RGB(255, 255, 255),
		TextMuted:    RGB(165, 180, 252),
		Primary:      RGB(255, 0, 128),        // Neon Pink
		PrimaryEnd:   RGB(255, 230, 0),        // Neon Yellow
		Secondary:    RGB(0, 240, 255),        // Neon Cyan
		Success:      RGB(0, 255, 159),        // Matrix Green
		Warning:      RGB(255, 215, 0),        // Bright Gold
		Danger:       RGB(255, 33, 79),        // Crimson Neon
		Accent:       RGB(189, 0, 255),        // Electric Violet
		BackdropBlur: 32,
		CornerRadius: 8,
	}

	// ThemeNord is an Arctic, north-bluish clean theme.
	ThemeNord = Theme{
		Name:         "Nord",
		Background:   RGB(46, 52, 64),         // Nord0 Polar Night
		CardBg:       ARGB(210, 59, 66, 82),   // Nord1
		Border:       RGB(136, 192, 208),      // Nord8 Frost
		BorderGlow:   ARGB(120, 129, 161, 193),
		TextPrimary:  RGB(236, 239, 244),      // Nord6 Snow Storm
		TextMuted:    RGB(216, 222, 233),      // Nord4
		Primary:      RGB(136, 192, 208),      // Nord8 Frost
		PrimaryEnd:   RGB(94, 129, 172),       // Nord10 Frost Deep
		Secondary:    RGB(180, 142, 173),      // Nord15 Aurora Purple
		Success:      RGB(163, 190, 140),      // Nord14 Aurora Green
		Warning:      RGB(235, 203, 139),      // Nord13 Aurora Yellow
		Danger:       RGB(191, 97, 106),       // Nord11 Aurora Red
		Accent:       RGB(143, 188, 187),      // Nord7
		BackdropBlur: 16,
		CornerRadius: 10,
	}

	// ThemeLight is a sleek modern light mode with subtle borders and shadows.
	ThemeLight = Theme{
		Name:         "Light",
		Background:   RGB(248, 250, 252),      // Slate 50
		CardBg:       ARGB(220, 255, 255, 255),// Pure frosted white
		Border:       RGB(203, 213, 225),      // Slate 300
		BorderGlow:   ARGB(80, 99, 102, 241),  // Indigo glow
		TextPrimary:  RGB(15, 23, 42),         // Slate 900
		TextMuted:    RGB(100, 116, 139),      // Slate 500
		Primary:      RGB(79, 70, 229),        // Indigo 600
		PrimaryEnd:   RGB(37, 99, 235),        // Blue 600
		Secondary:    RGB(124, 58, 237),       // Violet 600
		Success:      RGB(22, 163, 74),        // Green 600
		Warning:      RGB(217, 119, 6),        // Amber 600
		Danger:       RGB(225, 29, 72),        // Rose 600
		Accent:       RGB(13, 148, 136),       // Teal 600
		BackdropBlur: 12,
		CornerRadius: 12,
	}

	// ThemeSolarized is Ethan Schoonover's warm precision solarized dark palette.
	ThemeSolarized = Theme{
		Name:         "Solarized",
		Background:   RGB(0, 43, 54),          // Base03
		CardBg:       ARGB(210, 7, 54, 66),    // Base02
		Border:       RGB(38, 139, 210),       // Blue
		BorderGlow:   ARGB(100, 42, 161, 152), // Cyan glow
		TextPrimary:  RGB(253, 246, 227),      // Base3
		TextMuted:    RGB(147, 161, 161),      // Base1
		Primary:      RGB(38, 139, 210),       // Blue
		PrimaryEnd:   RGB(42, 161, 152),       // Cyan
		Secondary:    RGB(108, 113, 196),      // Violet
		Success:      RGB(133, 153, 0),        // Green
		Warning:      RGB(181, 137, 0),        // Yellow
		Danger:       RGB(220, 50, 47),        // Red
		Accent:       RGB(203, 75, 22),        // Orange
		BackdropBlur: 16,
		CornerRadius: 8,
	}

	// ThemeSunset is a vibrant warm dusk gradient theme.
	ThemeSunset = Theme{
		Name:         "Sunset",
		Background:   RGB(24, 12, 38),         // Deep twilight
		CardBg:       ARGB(200, 48, 22, 70),   // Violet twilight
		Border:       RGB(251, 113, 133),      // Coral Rose
		BorderGlow:   ARGB(120, 249, 115, 22), // Orange glow
		TextPrimary:  RGB(255, 245, 245),
		TextMuted:    RGB(203, 178, 203),
		Primary:      RGB(244, 63, 94),        // Rose
		PrimaryEnd:   RGB(249, 115, 22),       // Orange
		Secondary:    RGB(168, 85, 247),       // Purple
		Success:      RGB(52, 211, 153),       // Emerald
		Warning:      RGB(251, 191, 36),       // Amber
		Danger:       RGB(239, 68, 68),        // Red
		Accent:       RGB(236, 72, 153),       // Pink
		BackdropBlur: 20,
		CornerRadius: 14,
	}

	// themesRegistry stores all themes indexed by name.
	themesRegistry = map[string]Theme{
		"glassdark": ThemeGlassDark,
		"cyberpunk": ThemeCyberpunk,
		"nord":      ThemeNord,
		"light":     ThemeLight,
		"solarized": ThemeSolarized,
		"sunset":    ThemeSunset,
	}
)

// GetTheme retrieves a theme by name (case-insensitive), defaulting to ThemeGlassDark if not found.
func GetTheme(name string) Theme {
	clean := strings.ToLower(strings.TrimSpace(name))
	if t, ok := themesRegistry[clean]; ok {
		return t
	}
	return ThemeGlassDark
}

// RegisterTheme adds or replaces a theme in the runtime registry.
func RegisterTheme(t Theme) {
	clean := strings.ToLower(strings.TrimSpace(t.Name))
	themesRegistry[clean] = t
}

// ListThemes returns the names of all registered themes.
func ListThemes() []string {
	names := make([]string, 0, len(themesRegistry))
	for _, t := range themesRegistry {
		names = append(names, t.Name)
	}
	return names
}
