package gui

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
	"sync"
)

// Unit represents responsive dimension units.
type Unit string

const (
	UnitPx   Unit = "px"
	UnitPct  Unit = "%"
	UnitVw   Unit = "vw"
	UnitVh   Unit = "vh"
	UnitAuto Unit = "auto"
)

// Style defines the universal glassmorphic & spatial design token system.
type Style struct {
	Width          float64 `json:"width,omitempty"`
	Height         float64 `json:"height,omitempty"`
	Background     string  `json:"background,omitempty"`
	BackdropBlur   float64 `json:"backdrop_blur,omitempty"`
	BorderRadius   float64 `json:"border_radius,omitempty"`
	BorderColor    string  `json:"border_color,omitempty"`
	BoxShadow      string  `json:"box_shadow,omitempty"`
	Color          string  `json:"color,omitempty"`
	FontSize       float64 `json:"font_size,omitempty"`
	FontWeight     string  `json:"font_weight,omitempty"`
	Padding        string  `json:"padding,omitempty"`
	Margin         string  `json:"margin,omitempty"`
	FlexDirection  string  `json:"flex_direction,omitempty"`  // "row", "column"
	AlignItems     string  `json:"align_items,omitempty"`      // "center", "flex-start", "flex-end"
	JustifyContent string  `json:"justify_content,omitempty"`  // "center", "space-between", "space-around"
	Gap            string  `json:"gap,omitempty"`
	Cursor         string  `json:"cursor,omitempty"`
	Transition     string  `json:"transition,omitempty"`
}

// UIElement represents a visual component node in the AXLUI scene graph
type UIElement struct {
	ID       string       `json:"id"`
	Tag      string       `json:"tag"` // "div", "button", "text", "canvas", "input", "card", "metric", "chart"
	Text     string       `json:"text,omitempty"`
	Value    string       `json:"value,omitempty"`
	Style    Style        `json:"style"`
	Children []*UIElement `json:"children,omitempty"`
	OnClick  string       `json:"on_click,omitempty"`
}

// AXLUIModel represents a bidirectional editable UI file & live state
type AXLUIModel struct {
	FilePath string       `json:"file_path"`
	Title    string       `json:"title"`
	Theme    string       `json:"theme"`
	Root     *UIElement   `json:"root"`
	mu       sync.RWMutex
}

func NewAXLUIModel(filePath, title string) *AXLUIModel {
	return &AXLUIModel{
		FilePath: filePath,
		Title:    title,
		Theme:    "GlassDark",
		Root: &UIElement{
			ID:   "viewport_root",
			Tag:  "div",
			Text: "",
			Style: Style{
				Width:          1200,
				Height:         800,
				Background:     "radial-gradient(circle at 50% 0%, #1e1b4b 0%, #030712 100%)",
				BackdropBlur:   24,
				BorderRadius:   16,
				BorderColor:    "rgba(255, 255, 255, 0.12)",
				BoxShadow:      "0 25px 50px -12px rgba(0, 0, 0, 0.5)",
				Color:          "#f8fafc",
				Padding:        "32px",
				FlexDirection:  "column",
				AlignItems:     "center",
				JustifyContent: "flex-start",
				Gap:            "24px",
			},
			Children: make([]*UIElement, 0),
		},
	}
}

// ============================================================================
// 🌟 ULTRA-HIGH-LEVEL FLUENT BUILDER API (Создание сложных UI в 1-2 строки)
// ============================================================================

// App builds a root window application with theme & title
func App(title string) *AXLUIModel {
	return NewAXLUIModel("app.axlui", title)
}

// Window creates an expressive spatial glass container
func Window(title string, children ...*UIElement) *AXLUIModel {
	app := NewAXLUIModel("app.axlui", title)
	app.Root.Children = append(app.Root.Children, children...)
	return app
}

// Header creates a spatial frosted header with icon, title and status badge
func Header(title string, statusText string) *UIElement {
	return &UIElement{
		ID:  "header_" + sanitize(title),
		Tag: "div",
		Style: Style{
			Width:          1120,
			Height:         72,
			Background:     "rgba(15, 23, 42, 0.65)",
			BackdropBlur:   16,
			BorderRadius:   12,
			BorderColor:    "rgba(56, 189, 248, 0.25)",
			BoxShadow:      "0 10px 15px -3px rgba(0, 0, 0, 0.3)",
			Padding:        "16px 24px",
			FlexDirection:  "row",
			AlignItems:     "center",
			JustifyContent: "space-between",
		},
		Children: []*UIElement{
			{
				ID:   "header_title",
				Tag:  "text",
				Text: title,
				Style: Style{
					FontSize:   22,
					FontWeight: "700",
					Color:      "#38bdf8",
				},
			},
			{
				ID:   "header_badge",
				Tag:  "button",
				Text: statusText,
				Style: Style{
					Background:   "linear-gradient(135deg, #10b981, #059669)",
					BorderRadius: 8,
					Padding:      "6px 14px",
					Color:        "#ffffff",
					FontSize:     12,
					FontWeight:   "600",
				},
			},
		},
	}
}

// Card creates a glassmorphic content container
func Card(title, content string) *UIElement {
	return &UIElement{
		ID:  "card_" + sanitize(title),
		Tag: "div",
		Style: Style{
			Width:        350,
			Height:       240,
			Background:   "rgba(30, 41, 59, 0.7)",
			BackdropBlur: 20,
			BorderRadius: 16,
			BorderColor:  "rgba(255, 255, 255, 0.1)",
			BoxShadow:    "0 15px 25px rgba(0,0,0,0.4)",
			Padding:      "24px",
			FlexDirection: "column",
			Gap:          "12px",
		},
		Children: []*UIElement{
			{
				ID:   "card_title_" + sanitize(title),
				Tag:  "text",
				Text: title,
				Style: Style{
					FontSize:   18,
					FontWeight: "600",
					Color:      "#f1f5f9",
				},
			},
			{
				ID:   "card_body_" + sanitize(title),
				Tag:  "text",
				Text: content,
				Style: Style{
					FontSize: 14,
					Color:    "#94a3b8",
				},
			},
		},
	}
}

// MetricBadge creates a futuristic cyber-metric display widget
func MetricBadge(label string, value string, delta string) *UIElement {
	return &UIElement{
		ID:  "metric_" + sanitize(label),
		Tag: "div",
		Style: Style{
			Width:        260,
			Height:       130,
			Background:   "rgba(15, 23, 42, 0.8)",
			BackdropBlur: 24,
			BorderRadius: 14,
			BorderColor:  "rgba(168, 85, 247, 0.3)",
			BoxShadow:    "0 10px 20px rgba(0,0,0,0.35)",
			Padding:      "18px",
			FlexDirection: "column",
			JustifyContent: "space-between",
		},
		Children: []*UIElement{
			{
				ID:   "m_lbl_" + sanitize(label),
				Tag:  "text",
				Text: strings.ToUpper(label),
				Style: Style{
					FontSize:   11,
					FontWeight: "700",
					Color:      "#a855f7",
				},
			},
			{
				ID:   "m_val_" + sanitize(label),
				Tag:  "text",
				Text: value,
				Style: Style{
					FontSize:   26,
					FontWeight: "800",
					Color:      "#f8fafc",
				},
			},
			{
				ID:   "m_dlt_" + sanitize(label),
				Tag:  "text",
				Text: delta,
				Style: Style{
					FontSize:   12,
					Color:      "#4ade80",
					FontWeight: "600",
				},
			},
		},
	}
}

// Button creates an animated gradient interactive button
func Button(text, colorGradient string, onClickHandler string) *UIElement {
	bg := "linear-gradient(135deg, #6366f1, #a855f7)"
	if colorGradient == "cyan" {
		bg = "linear-gradient(135deg, #0ea5e9, #2563eb)"
	} else if colorGradient == "emerald" {
		bg = "linear-gradient(135deg, #10b981, #059669)"
	} else if colorGradient == "rose" {
		bg = "linear-gradient(135deg, #f43f5e, #e11d48)"
	}

	return &UIElement{
		ID:      "btn_" + sanitize(text),
		Tag:     "button",
		Text:    text,
		OnClick: onClickHandler,
		Style: Style{
			Width:        220,
			Height:       46,
			Background:   bg,
			BorderRadius: 10,
			Color:        "#ffffff",
			FontSize:     15,
			FontWeight:   "600",
			BoxShadow:    "0 10px 20px -3px rgba(99, 102, 241, 0.4)",
			Cursor:       "pointer",
			Transition:   "transform 0.2s ease, box-shadow 0.2s ease",
		},
	}
}

// Row creates a flex horizontal layout row
func Row(children ...*UIElement) *UIElement {
	return &UIElement{
		ID:  "row_container",
		Tag: "div",
		Style: Style{
			Width:          1120,
			FlexDirection:  "row",
			AlignItems:     "center",
			JustifyContent: "space-between",
			Gap:            "20px",
		},
		Children: children,
	}
}

// Grid creates a responsive auto-wrapping grid
func Grid(children ...*UIElement) *UIElement {
	return &UIElement{
		ID:  "grid_container",
		Tag: "div",
		Style: Style{
			Width:          1120,
			FlexDirection:  "row",
			AlignItems:     "center",
			JustifyContent: "space-between",
			Gap:            "24px",
		},
		Children: children,
	}
}

func sanitize(s string) string {
	reg := regexp.MustCompile("[^a-zA-Z0-9_]+")
	return strings.ToLower(reg.ReplaceAllString(s, "_"))
}

// SaveAXLUI serializes the model to .axlui format
func (m *AXLUIModel) SaveAXLUI() error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(m.FilePath, data, 0644)
}

// LoadAXLUI reads and deserializes an .axlui file into AXLUIModel
func LoadAXLUI(filePath string) (*AXLUIModel, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	var model AXLUIModel
	if err := json.Unmarshal(data, &model); err != nil {
		return nil, err
	}
	model.FilePath = filePath
	return &model, nil
}

// ImportHTMLCSS parses raw HTML/CSS and converts it cleanly into AXLUI scene graph
func (m *AXLUIModel) ImportHTMLCSS(rawHTML string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	reButton := regexp.MustCompile(`<button[^>]*>(.*?)</button>`)
	reText := regexp.MustCompile(`<(?:h[1-6]|p|span)[^>]*>(.*?)</(?:h[1-6]|p|span)>`)

	newChildren := []*UIElement{}

	textMatches := reText.FindAllStringSubmatch(rawHTML, -1)
	for i, match := range textMatches {
		if len(match) > 1 {
			newChildren = append(newChildren, &UIElement{
				ID:   fmt.Sprintf("imported_text_%d", i+1),
				Tag:  "text",
				Text: strings.TrimSpace(match[1]),
				Style: Style{
					FontSize: 20,
					Color:    "#f1f5f9",
				},
			})
		}
	}

	btnMatches := reButton.FindAllStringSubmatch(rawHTML, -1)
	for i, match := range btnMatches {
		if len(match) > 1 {
			newChildren = append(newChildren, &UIElement{
				ID:   fmt.Sprintf("imported_btn_%d", i+1),
				Tag:  "button",
				Text: strings.TrimSpace(match[1]),
				Style: Style{
					Width:        200,
					Height:       44,
					Background:   "linear-gradient(135deg, #0ea5e9, #6366f1)",
					BorderRadius: 8,
					Color:        "#ffffff",
					FontSize:     14,
				},
			})
		}
	}

	if len(newChildren) > 0 {
		m.Root.Children = newChildren
	}

	return nil
}

// GenerateAxiomCode compiles the UI into pure declarative AxiomLang source code
func (m *AXLUIModel) GenerateAxiomCode() string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var sb strings.Builder
	sb.WriteString("// Generated by AxiomLang Visual Studio & GlassDark GUI Engine\n")
	sb.WriteString("require cap[GUI.Render]\n")
	sb.WriteString("require cap[Console.Out]\n\n")
	sb.WriteString(fmt.Sprintf("intent \"Рендеринг интерфейса: %s\" {\n", m.Title))

	for _, child := range m.Root.Children {
		sb.WriteString(fmt.Sprintf("    let elem_%s = { id: %q, tag: %q, text: %q };\n",
			child.ID, child.ID, child.Tag, child.Text))
		sb.WriteString(fmt.Sprintf("    gui.render_element(elem_%s);\n", child.ID))
	}
	sb.WriteString("}\n")
	return sb.String()
}

// RenderToBMP renders the UI model directly to a native 32-bit BMP framebuffer image without OS/browser
func (m *AXLUIModel) RenderToBMP(outputPath string) error {
	canvas := NewNativeCanvas(1920, 1080)
	RenderSceneToCanvas(canvas, m)
	return canvas.SaveAsBMP(outputPath)
}
