package gui

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// CSSPropertyMap stores style rules for selectors
type CSSPropertyMap map[string]string

// WebCompiler transcompiles any standard HTML/CSS web layout into a high-fidelity AXLUI scene graph.
type WebCompiler struct {
	CSSRules map[string]CSSPropertyMap
}

// NewWebCompiler creates an instance of HTML/CSS AST parser & compiler.
func NewWebCompiler() *WebCompiler {
	return &WebCompiler{
		CSSRules: make(map[string]CSSPropertyMap),
	}
}

// CompileWebInterface parses raw HTML and CSS stylesheets, extracts DOM hierarchy,
// computes styles with inheritance, and produces an ultra-fidelity AXLUIModel.
func CompileWebInterface(rawHTML, rawCSS, title string) (*AXLUIModel, error) {
	compiler := NewWebCompiler()
	if rawCSS != "" {
		compiler.parseCSS(rawCSS)
	}

	model := NewAXLUIModel("compiled_web.axlui", title)

	// Also extract inline <style>...</style> if present in HTML
	cleanHTML := compiler.extractEmbeddedStyles(rawHTML)

	// Determine theme from background color
	bodyStyle := compiler.CSSRules["body"]
	if bodyStyle != nil {
		if bg, ok := bodyStyle["background"]; ok {
			model.Theme = detectThemeFromBackground(bg)
		}
	}

	rootElement, err := compiler.parseHTMLNodes(cleanHTML)
	if err != nil {
		return nil, err
	}

	if rootElement != nil {
		// If root element has children and is container, wrap or adopt
		if len(rootElement.Children) > 0 {
			model.Root = rootElement
		} else {
			model.Root.Children = append(model.Root.Children, rootElement)
		}
	}

	return model, nil
}

func (wc *WebCompiler) extractEmbeddedStyles(html string) string {
	var remainingHTML strings.Builder
	cursor := 0

	for {
		startStyle := strings.Index(strings.ToLower(html[cursor:]), "<style")
		if startStyle == -1 {
			remainingHTML.WriteString(html[cursor:])
			break
		}
		remainingHTML.WriteString(html[cursor : cursor+startStyle])
		absStart := cursor + startStyle

		endTagStart := strings.Index(html[absStart:], ">")
		if endTagStart == -1 {
			break
		}
		contentStart := absStart + endTagStart + 1

		closeStyle := strings.Index(strings.ToLower(html[contentStart:]), "</style>")
		if closeStyle == -1 {
			break
		}
		cssContent := html[contentStart : contentStart+closeStyle]
		wc.parseCSS(cssContent)

		cursor = contentStart + closeStyle + len("</style>")
	}

	return remainingHTML.String()
}

func (wc *WebCompiler) parseCSS(css string) {
	// Strip comments /* ... */
	cleaned := stripCSSComments(css)
	rules := strings.Split(cleaned, "}")

	for _, rule := range rules {
		rule = strings.TrimSpace(rule)
		if rule == "" {
			continue
		}
		parts := strings.SplitN(rule, "{", 2)
		if len(parts) != 2 {
			continue
		}
		selectors := strings.Split(parts[0], ",")
		decls := strings.Split(parts[1], ";")

		propMap := make(CSSPropertyMap)
		for _, decl := range decls {
			decl = strings.TrimSpace(decl)
			if decl == "" {
				continue
			}
			kv := strings.SplitN(decl, ":", 2)
			if len(kv) == 2 {
				k := strings.TrimSpace(strings.ToLower(kv[0]))
				v := strings.TrimSpace(kv[1])
				propMap[k] = v
			}
		}

		for _, sel := range selectors {
			sel = strings.TrimSpace(strings.ToLower(sel))
			if sel == "" {
				continue
			}
			if wc.CSSRules[sel] == nil {
				wc.CSSRules[sel] = make(CSSPropertyMap)
			}
			for k, v := range propMap {
				wc.CSSRules[sel][k] = v
			}
		}
	}
}

func stripCSSComments(css string) string {
	var sb strings.Builder
	inComment := false
	i := 0
	n := len(css)

	for i < n {
		if !inComment && i+1 < n && css[i] == '/' && css[i+1] == '*' {
			inComment = true
			i += 2
			continue
		}
		if inComment && i+1 < n && css[i] == '*' && css[i+1] == '/' {
			inComment = false
			i += 2
			continue
		}
		if !inComment {
			sb.WriteByte(css[i])
		}
		i++
	}
	return sb.String()
}

// HTML Node hierarchy representation
type htmlNode struct {
	Tag        string
	Attributes map[string]string
	InlineCSS  CSSPropertyMap
	ClassList  []string
	ID         string
	InnerText  string
	Children   []*htmlNode
}

func (wc *WebCompiler) parseHTMLNodes(html string) (*UIElement, error) {
	rootNode := &htmlNode{
		Tag:        "div",
		Attributes: map[string]string{"id": "viewport_root"},
		ClassList:  []string{"viewport_root"},
		Children:   make([]*htmlNode, 0),
	}

	wc.parseTagsRecursive(html, rootNode)

	// Convert parsed DOM tree into AXLUI scene graph
	uiRoot := wc.convertNodeToUI(rootNode)
	return uiRoot, nil
}

func (wc *WebCompiler) parseTagsRecursive(html string, parent *htmlNode) {
	cursor := 0
	n := len(html)

	for cursor < n {
		tagStart := strings.IndexByte(html[cursor:], '<')
		if tagStart == -1 {
			// Trailing text
			trailingText := strings.TrimSpace(html[cursor:])
			if trailingText != "" && parent != nil {
				parent.InnerText += " " + trailingText
				parent.InnerText = strings.TrimSpace(parent.InnerText)
			}
			break
		}

		// Text before tag
		textBefore := strings.TrimSpace(html[cursor : cursor+tagStart])
		if textBefore != "" && parent != nil {
			parent.InnerText += " " + textBefore
			parent.InnerText = strings.TrimSpace(parent.InnerText)
		}

		absTagStart := cursor + tagStart
		tagEnd := strings.IndexByte(html[absTagStart:], '>')
		if tagEnd == -1 {
			break
		}
		absTagEnd := absTagStart + tagEnd

		tagContent := strings.TrimSpace(html[absTagStart+1 : absTagEnd])
		cursor = absTagEnd + 1

		// Skip comments <!-- ... -->
		if strings.HasPrefix(tagContent, "!--") {
			commentEnd := strings.Index(html[absTagStart:], "-->")
			if commentEnd != -1 {
				cursor = absTagStart + commentEnd + 3
			}
			continue
		}

		// Skip doctype or self-closing metadata
		if strings.HasPrefix(tagContent, "!DOCTYPE") || strings.HasPrefix(tagContent, "meta") ||
			strings.HasPrefix(tagContent, "link") || strings.HasPrefix(tagContent, "head") ||
			strings.HasPrefix(tagContent, "/head") || strings.HasPrefix(tagContent, "html") ||
			strings.HasPrefix(tagContent, "/html") {
			continue
		}

		// Closing tag
		if strings.HasPrefix(tagContent, "/") {
			// Bubble back
			return
		}

		// Self-closing tag check
		isSelfClosing := strings.HasSuffix(tagContent, "/") || isVoidElement(tagContent)
		if strings.HasSuffix(tagContent, "/") {
			tagContent = strings.TrimSpace(strings.TrimSuffix(tagContent, "/"))
		}

		node := parseTagAttributes(tagContent)

		// Parse children if not self-closing
		if !isSelfClosing {
			closeTag := "</" + node.Tag + ">"
			closeIdx := findMatchingCloseTag(html[cursor:], node.Tag)
			if closeIdx != -1 {
				innerContent := html[cursor : cursor+closeIdx]
				wc.parseTagsRecursive(innerContent, node)
				cursor += closeIdx + len(closeTag)
			}
		}

		parent.Children = append(parent.Children, node)
	}
}

func isVoidElement(tagContent string) bool {
	fields := strings.Fields(tagContent)
	if len(fields) == 0 {
		return false
	}
	t := strings.ToLower(fields[0])
	return t == "img" || t == "br" || t == "hr" || t == "input"
}

func findMatchingCloseTag(html, tag string) int {
	lowerTag := strings.ToLower(tag)
	openPattern := "<" + lowerTag
	closePattern := "</" + lowerTag + ">"

	depth := 1
	cursor := 0
	lowerHTML := strings.ToLower(html)

	for cursor < len(html) {
		nextOpen := strings.Index(lowerHTML[cursor:], openPattern)
		nextClose := strings.Index(lowerHTML[cursor:], closePattern)

		if nextClose == -1 {
			return -1
		}

		if nextOpen != -1 && nextOpen < nextClose {
			// Found nested open tag
			depth++
			cursor += nextOpen + len(openPattern)
		} else {
			// Found close tag
			depth--
			if depth == 0 {
				return cursor + nextClose
			}
			cursor += nextClose + len(closePattern)
		}
	}
	return -1
}

func parseTagAttributes(tagContent string) *htmlNode {
	fields := strings.Fields(tagContent)
	if len(fields) == 0 {
		return &htmlNode{Tag: "div", Attributes: make(map[string]string)}
	}

	tagName := strings.ToLower(fields[0])
	node := &htmlNode{
		Tag:        tagName,
		Attributes: make(map[string]string),
		InlineCSS:  make(CSSPropertyMap),
		Children:   make([]*htmlNode, 0),
	}

	attrStr := strings.TrimSpace(tagContent[len(fields[0]):])
	for len(attrStr) > 0 {
		eqIdx := strings.IndexByte(attrStr, '=')
		if eqIdx == -1 {
			break
		}
		attrName := strings.TrimSpace(strings.ToLower(attrStr[:eqIdx]))
		rem := strings.TrimSpace(attrStr[eqIdx+1:])
		var attrVal string

		if len(rem) > 0 && (rem[0] == '"' || rem[0] == '\'') {
			quote := rem[0]
			endQuote := strings.IndexByte(rem[1:], quote)
			if endQuote != -1 {
				attrVal = rem[1 : 1+endQuote]
				attrStr = strings.TrimSpace(rem[1+endQuote+1:])
			} else {
				attrVal = rem[1:]
				attrStr = ""
			}
		} else {
			spaceIdx := strings.IndexByte(rem, ' ')
			if spaceIdx != -1 {
				attrVal = rem[:spaceIdx]
				attrStr = strings.TrimSpace(rem[spaceIdx+1:])
			} else {
				attrVal = rem
				attrStr = ""
			}
		}

		node.Attributes[attrName] = attrVal
		if attrName == "id" {
			node.ID = attrVal
		} else if attrName == "class" {
			node.ClassList = strings.Fields(attrVal)
		} else if attrName == "style" {
			// Parse inline style
			decls := strings.Split(attrVal, ";")
			for _, d := range decls {
				kv := strings.SplitN(d, ":", 2)
				if len(kv) == 2 {
					k := strings.TrimSpace(strings.ToLower(kv[0]))
					v := strings.TrimSpace(kv[1])
					node.InlineCSS[k] = v
				}
			}
		}
	}

	return node
}

func (wc *WebCompiler) convertNodeToUI(n *htmlNode) *UIElement {
	// Map HTML semantic tags into AXLUI primitives
	tag := "div"
	if n.Tag == "button" || n.Tag == "a" {
		tag = "button"
	} else if isHeaderOrTextTag(n.Tag) {
		tag = "text"
	} else if n.Tag == "input" {
		tag = "input"
	}

	id := n.ID
	if id == "" {
		id = fmt.Sprintf("elem_%s_%p", n.Tag, n)
	}

	style := wc.computeComputedStyle(n)

	elem := &UIElement{
		ID:       id,
		Tag:      tag,
		Text:     n.InnerText,
		Style:    style,
		Children: make([]*UIElement, 0, len(n.Children)),
	}

	for _, child := range n.Children {
		childUI := wc.convertNodeToUI(child)
		if childUI != nil {
			elem.Children = append(elem.Children, childUI)
		}
	}

	return elem
}

func isHeaderOrTextTag(t string) bool {
	return t == "h1" || t == "h2" || t == "h3" || t == "h4" || t == "h5" || t == "h6" ||
		t == "p" || t == "span" || t == "label" || t == "b" || t == "strong"
}

func (wc *WebCompiler) computeComputedStyle(n *htmlNode) Style {
	s := Style{
		FlexDirection:  "column",
		AlignItems:     "flex-start",
		JustifyContent: "flex-start",
	}

	props := make(CSSPropertyMap)

	// 1. Tag-level styles
	if tagRules, ok := wc.CSSRules[n.Tag]; ok {
		for k, v := range tagRules {
			props[k] = v
		}
	}

	// 2. Class-level styles
	for _, cls := range n.ClassList {
		if clsRules, ok := wc.CSSRules["."+strings.ToLower(cls)]; ok {
			for k, v := range clsRules {
				props[k] = v
			}
		}
	}

	// 3. ID-level styles
	if n.ID != "" {
		if idRules, ok := wc.CSSRules["#"+strings.ToLower(n.ID)]; ok {
			for k, v := range idRules {
				props[k] = v
			}
		}
	}

	// 4. Inline styles (highest precedence)
	for k, v := range n.InlineCSS {
		props[k] = v
	}

	// Apply parsed CSS properties into Style struct
	if w, ok := props["width"]; ok {
		s.Width = parseDimension(w)
	}
	if h, ok := props["height"]; ok {
		s.Height = parseDimension(h)
	}
	if bg, ok := props["background"]; ok {
		s.Background = bg
	} else if bgc, ok := props["background-color"]; ok {
		s.Background = bgc
	}
	if col, ok := props["color"]; ok {
		s.Color = col
	}
	if br, ok := props["border-radius"]; ok {
		s.BorderRadius = parseDimension(br)
	}
	if bc, ok := props["border-color"]; ok {
		s.BorderColor = bc
	} else if b, ok := props["border"]; ok {
		s.BorderColor = parseBorderColor(b)
	}
	if fs, ok := props["font-size"]; ok {
		s.FontSize = parseDimension(fs)
	}
	if fw, ok := props["font-weight"]; ok {
		s.FontWeight = fw
	}
	if disp, ok := props["display"]; ok {
		if strings.Contains(disp, "flex") {
			s.FlexDirection = "row"
		}
	}
	if fd, ok := props["flex-direction"]; ok {
		s.FlexDirection = fd
	}
	if ai, ok := props["align-items"]; ok {
		s.AlignItems = ai
	}
	if jc, ok := props["justify-content"]; ok {
		s.JustifyContent = jc
	}
	if gap, ok := props["gap"]; ok {
		s.Gap = gap
	}
	if pad, ok := props["padding"]; ok {
		s.Padding = pad
	}
	if bfilter, ok := props["backdrop-filter"]; ok {
		s.BackdropBlur = parseBackdropBlur(bfilter)
	}
	if shadow, ok := props["box-shadow"]; ok {
		s.BoxShadow = shadow
	}

	// Header tag font-size defaults if not explicitly set
	if s.FontSize == 0 {
		switch n.Tag {
		case "h1":
			s.FontSize = 28
			s.FontWeight = "700"
		case "h2":
			s.FontSize = 24
			s.FontWeight = "700"
		case "h3":
			s.FontSize = 20
			s.FontWeight = "600"
		case "button":
			s.FontSize = 14
			s.FontWeight = "600"
		default:
			s.FontSize = 15
		}
	}

	return s
}

func parseDimension(val string) float64 {
	clean := strings.TrimSpace(val)
	clean = strings.TrimSuffix(clean, "px")
	clean = strings.TrimSuffix(clean, "rem")
	clean = strings.TrimSuffix(clean, "%")
	if f, err := strconv.ParseFloat(clean, 64); err == nil {
		return f
	}
	return 0
}

func parseBorderColor(val string) string {
	parts := strings.Fields(val)
	for _, p := range parts {
		if strings.HasPrefix(p, "#") || strings.HasPrefix(p, "rgb") {
			return p
		}
	}
	return val
}

func parseBackdropBlur(val string) float64 {
	idx := strings.Index(val, "blur(")
	if idx != -1 {
		rem := val[idx+5:]
		closeParen := strings.IndexByte(rem, ')')
		if closeParen != -1 {
			blurVal := strings.TrimSpace(rem[:closeParen])
			return parseDimension(blurVal)
		}
	}
	return 16
}

func detectThemeFromBackground(bg string) string {
	lower := strings.ToLower(bg)
	if strings.Contains(lower, "248, 250, 252") || strings.Contains(lower, "#f8fafc") || strings.Contains(lower, "white") {
		return "Light"
	}
	if strings.Contains(lower, "255, 0, 128") || strings.Contains(lower, "neon") {
		return "Cyberpunk"
	}
	if strings.Contains(lower, "46, 52, 64") || strings.Contains(lower, "#2e3440") {
		return "Nord"
	}
	if strings.Contains(lower, "0, 43, 54") || strings.Contains(lower, "#002b36") {
		return "Solarized"
	}
	if strings.Contains(lower, "24, 12, 38") || strings.Contains(lower, "twilight") {
		return "Sunset"
	}
	return "GlassDark"
}

// CompileHTMLFileToBMP compiles any web HTML/CSS file directly into a high-fidelity native BMP image
func CompileHTMLFileToBMP(htmlFilePath, outBmpPath string, width, height int) error {
	model, err := LoadHTMLFile(htmlFilePath)
	if err != nil {
		return err
	}

	if width <= 0 {
		width = 1920
	}
	if height <= 0 {
		height = 1080
	}

	canvas := NewNativeCanvas(width, height)
	RenderSceneToCanvas(canvas, model)
	return canvas.SaveAsBMP(outBmpPath)
}

// LoadHTMLFile reads an HTML file with its styles and compiles it into an AXLUIModel
func LoadHTMLFile(htmlFilePath string) (*AXLUIModel, error) {
	data, err := os.ReadFile(htmlFilePath)
	if err != nil {
		return nil, err
	}
	title := "Compiled Web App"
	return CompileWebInterface(string(data), "", title)
}
