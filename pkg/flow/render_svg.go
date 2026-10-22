package flow

import (
	"fmt"
	"strings"
)

// ToMermaid converts a Flowchart to GitHub Flavored Mermaid diagram markdown.
func (fc *Flowchart) ToMermaid() string {
	var sb strings.Builder
	sb.WriteString("flowchart TD\n")

	for _, n := range fc.Nodes {
		label := strings.ReplaceAll(n.Label, "\"", "'")
		switch n.Kind {
		case ShapeTerminator:
			sb.WriteString(fmt.Sprintf("    %s([\"%s\"])\n", n.ID, label))
		case ShapeDecision:
			sb.WriteString(fmt.Sprintf("    %s{\"%s\"}\n", n.ID, label))
		case ShapeIO:
			sb.WriteString(fmt.Sprintf("    %s[\\\"%s\"/]\n", n.ID, label))
		case ShapeSubroutine:
			sb.WriteString(fmt.Sprintf("    %s[[\"%s\"]]\n", n.ID, label))
		case ShapeLoop:
			sb.WriteString(fmt.Sprintf("    %s{{\"%s\"}}\n", n.ID, label))
		default: // ShapeProcess
			sb.WriteString(fmt.Sprintf("    %s[\"%s\"]\n", n.ID, label))
		}
	}

	for _, e := range fc.Edges {
		if e.Label != "" {
			sb.WriteString(fmt.Sprintf("    %s -->|\"%s\"| %s\n", e.FromID, e.Label, e.ToID))
		} else {
			sb.WriteString(fmt.Sprintf("    %s --> %s\n", e.FromID, e.ToID))
		}
	}

	return sb.String()
}

// ToSVG converts a Flowchart into a vector SVG diagram.
func (fc *Flowchart) ToSVG() string {
	var sb strings.Builder
	w := fc.Width
	if w < 600 {
		w = 600
	}
	h := fc.Height
	if h < 400 {
		h = 400
	}

	sb.WriteString(fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 %d %d" width="%d" height="%d">`+"\n", w, h, w, h))
	sb.WriteString(`<defs>
  <marker id="arrow" viewBox="0 0 10 10" refX="8" refY="5" markerWidth="6" markerHeight="6" orient="auto-start-reverse">
    <path d="M 0 1 L 10 5 L 0 9 z" fill="#60a5fa" />
  </marker>
  <marker id="arrow-green" viewBox="0 0 10 10" refX="8" refY="5" markerWidth="6" markerHeight="6" orient="auto-start-reverse">
    <path d="M 0 1 L 10 5 L 0 9 z" fill="#34d399" />
  </marker>
  <marker id="arrow-red" viewBox="0 0 10 10" refX="8" refY="5" markerWidth="6" markerHeight="6" orient="auto-start-reverse">
    <path d="M 0 1 L 10 5 L 0 9 z" fill="#f87171" />
  </marker>
  <linearGradient id="grad-proc" x1="0%%" y1="0%%" x2="100%%" y2="100%%">
    <stop offset="0%%" stop-color="#1e293b"/>
    <stop offset="100%%" stop-color="#0f172a"/>
  </linearGradient>
  <linearGradient id="grad-term" x1="0%%" y1="0%%" x2="100%%" y2="100%%">
    <stop offset="0%%" stop-color="#1e3a8a"/>
    <stop offset="100%%" stop-color="#172554"/>
  </linearGradient>
  <linearGradient id="grad-dec" x1="0%%" y1="0%%" x2="100%%" y2="100%%">
    <stop offset="0%%" stop-color="#831843"/>
    <stop offset="100%%" stop-color="#500724"/>
  </linearGradient>
  <linearGradient id="grad-io" x1="0%%" y1="0%%" x2="100%%" y2="100%%">
    <stop offset="0%%" stop-color="#134e4a"/>
    <stop offset="100%%" stop-color="#042f2e"/>
  </linearGradient>
</defs>
<rect width="100%" height="100%" fill="#090d16"/>
`)

	// Title
	sb.WriteString(fmt.Sprintf(`<text x="30" y="35" fill="#f8fafc" font-family="system-ui, sans-serif" font-size="18" font-weight="bold">%s</text>`+"\n", escapeXML(fc.Title)))

	// Draw Edges first so nodes render on top
	for _, e := range fc.Edges {
		fromNode := fc.GetNode(e.FromID)
		toNode := fc.GetNode(e.ToID)
		if fromNode == nil || toNode == nil {
			continue
		}

		x1, y1 := getNodePortCoord(fromNode, e.FromPort)
		x2, y2 := getNodePortCoord(toNode, e.ToPort)

		strokeColor := "#60a5fa"
		markerId := "arrow"
		if e.Color == "#10b981" || strings.ToLower(e.Label) == "true" || strings.ToLower(e.Label) == "yes" {
			strokeColor = "#34d399"
			markerId = "arrow-green"
		} else if e.Color == "#ef4444" || strings.ToLower(e.Label) == "false" || strings.ToLower(e.Label) == "no" {
			strokeColor = "#f87171"
			markerId = "arrow-red"
		}

		if e.IsBackArc {
			// Loop curve to the left
			offset := 50
			sb.WriteString(fmt.Sprintf(`<path d="M %d %d C %d %d, %d %d, %d %d" fill="none" stroke="%s" stroke-width="2" stroke-dasharray="4,4" marker-end="url(#%s)"/>`+"\n",
				x1, y1, x1-offset, y1, x2-offset, y2, x2, y2, strokeColor, markerId))
		} else if x1 == x2 {
			sb.WriteString(fmt.Sprintf(`<line x1="%d" y1="%d" x2="%d" y2="%d" stroke="%s" stroke-width="2" marker-end="url(#%s)"/>`+"\n",
				x1, y1, x2, y2, strokeColor, markerId))
		} else {
			midY := (y1 + y2) / 2
			sb.WriteString(fmt.Sprintf(`<path d="M %d %d L %d %d L %d %d L %d %d" fill="none" stroke="%s" stroke-width="2" marker-end="url(#%s)"/>`+"\n",
				x1, y1, x1, midY, x2, midY, x2, y2, strokeColor, markerId))
		}

		if e.Label != "" {
			labelX := (x1 + x2) / 2
			labelY := (y1+y2)/2 - 6
			if e.IsBackArc {
				labelX = x1 - 35
			}
			sb.WriteString(fmt.Sprintf(`<text x="%d" y="%d" fill="%s" font-family="system-ui, sans-serif" font-size="11" font-weight="bold" text-anchor="middle">%s</text>`+"\n",
				labelX, labelY, strokeColor, escapeXML(e.Label)))
		}
	}

	// Draw Nodes
	for _, n := range fc.Nodes {
		switch n.Kind {
		case ShapeTerminator:
			sb.WriteString(fmt.Sprintf(`<rect x="%d" y="%d" width="%d" height="%d" rx="%d" ry="%d" fill="url(#grad-term)" stroke="#3b82f6" stroke-width="2"/>`+"\n",
				n.X, n.Y, n.Width, n.Height, n.Height/2, n.Height/2))
		case ShapeDecision:
			// Diamond
			midX := n.X + n.Width/2
			midY := n.Y + n.Height/2
			pts := fmt.Sprintf("%d,%d %d,%d %d,%d %d,%d",
				midX, n.Y, n.X+n.Width, midY, midX, n.Y+n.Height, n.X, midY)
			sb.WriteString(fmt.Sprintf(`<polygon points="%s" fill="url(#grad-dec)" stroke="#ec4899" stroke-width="2"/>`+"\n", pts))
		case ShapeIO:
			// Slanted parallelogram
			skew := 16
			pts := fmt.Sprintf("%d,%d %d,%d %d,%d %d,%d",
				n.X+skew, n.Y, n.X+n.Width, n.Y, n.X+n.Width-skew, n.Y+n.Height, n.X, n.Y+n.Height)
			sb.WriteString(fmt.Sprintf(`<polygon points="%s" fill="url(#grad-io)" stroke="#14b8a6" stroke-width="2"/>`+"\n", pts))
		case ShapeSubroutine:
			// Double vertical borders
			sb.WriteString(fmt.Sprintf(`<rect x="%d" y="%d" width="%d" height="%d" rx="6" fill="url(#grad-proc)" stroke="#a855f7" stroke-width="2"/>`+"\n",
				n.X, n.Y, n.Width, n.Height))
			sb.WriteString(fmt.Sprintf(`<line x1="%d" y1="%d" x2="%d" y2="%d" stroke="#a855f7" stroke-width="1.5"/>`+"\n",
				n.X+10, n.Y, n.X+10, n.Y+n.Height))
			sb.WriteString(fmt.Sprintf(`<line x1="%d" y1="%d" x2="%d" y2="%d" stroke="#a855f7" stroke-width="1.5"/>`+"\n",
				n.X+n.Width-10, n.Y, n.X+n.Width-10, n.Y+n.Height))
		default: // ShapeProcess & others
			sb.WriteString(fmt.Sprintf(`<rect x="%d" y="%d" width="%d" height="%d" rx="6" fill="url(#grad-proc)" stroke="#475569" stroke-width="2"/>`+"\n",
				n.X, n.Y, n.Width, n.Height))
		}

		// Node text
		textY := n.Y + n.Height/2 + 4
		sb.WriteString(fmt.Sprintf(`<text x="%d" y="%d" fill="#f1f5f9" font-family="monospace, sans-serif" font-size="12" text-anchor="middle">%s</text>`+"\n",
			n.X+n.Width/2, textY, escapeXML(n.Label)))
	}

	sb.WriteString("</svg>\n")
	return sb.String()
}

func getNodePortCoord(n *FlowNode, p Port) (int, int) {
	switch p {
	case PortTop:
		return n.X + n.Width/2, n.Y
	case PortBottom:
		return n.X + n.Width/2, n.Y + n.Height
	case PortLeft:
		return n.X, n.Y + n.Height/2
	case PortRight:
		return n.X + n.Width, n.Y + n.Height/2
	default:
		return n.X + n.Width/2, n.Y + n.Height
	}
}

func escapeXML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	return s
}
