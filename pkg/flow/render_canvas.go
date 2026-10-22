package flow

import (
	"math"

	"AxiomLang/stdlib/gui"
)

// RenderToCanvas renders the flowchart using Axiom's stdlib/gui NativeCanvas (120 FPS / BMP export / X11).
func (fc *Flowchart) RenderToCanvas() *gui.NativeCanvas {
	w := fc.Width
	if w < 1000 {
		w = 1000
	}
	h := fc.Height
	if h < 800 {
		h = 800
	}

	canvas := gui.NewNativeCanvas(w, h)
	// Dark background
	canvas.Clear(gui.RGB(10, 14, 23))

	// Draw Title
	canvas.DrawString(30, 25, fc.Title, gui.RGB(248, 250, 252), 2)

	// Draw Connections / Arrows
	for _, e := range fc.Edges {
		fromNode := fc.GetNode(e.FromID)
		toNode := fc.GetNode(e.ToID)
		if fromNode == nil || toNode == nil {
			continue
		}

		x1, y1 := getNodePortCoord(fromNode, e.FromPort)
		x2, y2 := getNodePortCoord(toNode, e.ToPort)

		edgeCol := gui.RGB(96, 165, 250) // Light blue
		if e.Color == "#10b981" || e.Label == "True" || e.Label == "Yes" {
			edgeCol = gui.RGB(52, 211, 153) // Green
		} else if e.Color == "#ef4444" || e.Label == "False" || e.Label == "No" {
			edgeCol = gui.RGB(248, 113, 113) // Red
		}

		if e.IsBackArc {
			// Loop arc around the left
			offset := 50
			drawThickLine(canvas, x1, y1, x1-offset, y1, edgeCol, 2)
			drawThickLine(canvas, x1-offset, y1, x1-offset, y2, edgeCol, 2)
			drawThickLine(canvas, x1-offset, y2, x2, y2, edgeCol, 2)
			drawArrowHead(canvas, x2-10, y2, x2, y2, edgeCol)
		} else if x1 == x2 {
			drawThickLine(canvas, x1, y1, x2, y2, edgeCol, 2)
			drawArrowHead(canvas, x2, y2-8, x2, y2, edgeCol)
		} else {
			midY := (y1 + y2) / 2
			drawThickLine(canvas, x1, y1, x1, midY, edgeCol, 2)
			drawThickLine(canvas, x1, midY, x2, midY, edgeCol, 2)
			drawThickLine(canvas, x2, midY, x2, y2, edgeCol, 2)
			drawArrowHead(canvas, x2, y2-8, x2, y2, edgeCol)
		}

		if e.Label != "" {
			lx := (x1+x2)/2 + 6
			ly := (y1+y2)/2 - 8
			if e.IsBackArc {
				lx = x1 - 50 + 8
				ly = (y1 + y2) / 2
			}
			canvas.DrawString(lx, ly, e.Label, edgeCol, 1)
		}
	}

	// Draw Nodes
	for _, n := range fc.Nodes {
		switch n.Kind {
		case ShapeTerminator:
			// Rounded Oval Terminator
			radius := n.Height / 2
			canvas.DrawRoundedRect(n.X, n.Y, n.Width, n.Height, radius, gui.RGB(23, 37, 84))
			canvas.DrawStrokeRoundedRect(n.X, n.Y, n.Width, n.Height, radius, 2, gui.RGB(59, 130, 246))

		case ShapeDecision:
			// Diamond (Decision)
			midX := n.X + n.Width/2
			midY := n.Y + n.Height/2
			drawDiamond(canvas, midX, midY, n.Width, n.Height, gui.RGB(80, 7, 36), gui.RGB(236, 72, 153))

		case ShapeIO:
			// Slanted parallelogram
			drawParallelogram(canvas, n.X, n.Y, n.Width, n.Height, 16, gui.RGB(19, 78, 74), gui.RGB(20, 184, 166))

		case ShapeSubroutine:
			// Rectangle with side bars
			canvas.DrawRoundedRect(n.X, n.Y, n.Width, n.Height, 6, gui.RGB(30, 41, 59))
			canvas.DrawStrokeRoundedRect(n.X, n.Y, n.Width, n.Height, 6, 2, gui.RGB(168, 85, 247))
			drawThickLine(canvas, n.X+10, n.Y, n.X+10, n.Y+n.Height, gui.RGB(168, 85, 247), 2)
			drawThickLine(canvas, n.X+n.Width-10, n.Y, n.X+n.Width-10, n.Y+n.Height, gui.RGB(168, 85, 247), 2)

		default: // ShapeProcess
			canvas.DrawRoundedRect(n.X, n.Y, n.Width, n.Height, 6, gui.RGB(30, 41, 59))
			canvas.DrawStrokeRoundedRect(n.X, n.Y, n.Width, n.Height, 6, 2, gui.RGB(71, 85, 105))
		}

		// Draw node label centered (8x16 font at scale 1)
		textW := len(n.Label) * 8
		textX := n.X + (n.Width-textW)/2
		if textX < n.X+5 {
			textX = n.X + 5
		}
		textY := n.Y + (n.Height-16)/2
		canvas.DrawString(textX, textY, n.Label, gui.RGB(241, 245, 249), 1)
	}

	return canvas
}

func drawThickLine(c *gui.NativeCanvas, x0, y0, x1, y1 int, col gui.Color32, thickness int) {
	dx := math.Abs(float64(x1 - x0))
	dy := math.Abs(float64(y1 - y0))
	sx := 1
	if x0 > x1 {
		sx = -1
	}
	sy := 1
	if y0 > y1 {
		sy = -1
	}
	err := dx - dy

	for {
		for tx := -thickness / 2; tx <= thickness/2; tx++ {
			for ty := -thickness / 2; ty <= thickness/2; ty++ {
				c.SetPixel(x0+tx, y0+ty, col)
			}
		}
		if x0 == x1 && y0 == y1 {
			break
		}
		e2 := 2 * err
		if e2 > -dy {
			err -= dy
			x0 += sx
		}
		if e2 < dx {
			err += dx
			y0 += sy
		}
	}
}

func drawArrowHead(c *gui.NativeCanvas, fromX, fromY, toX, toY int, col gui.Color32) {
	// Draw triangular arrow head pointing to (toX, toY)
	dx := float64(toX - fromX)
	dy := float64(toY - fromY)
	length := math.Hypot(dx, dy)
	if length == 0 {
		return
	}
	ux := dx / length
	uy := dy / length

	arrowSize := 8.0
	p1X := float64(toX) - arrowSize*ux + (arrowSize/2)*uy
	p1Y := float64(toY) - arrowSize*uy - (arrowSize/2)*ux

	p2X := float64(toX) - arrowSize*ux - (arrowSize/2)*uy
	p2Y := float64(toY) - arrowSize*uy + (arrowSize/2)*ux

	drawThickLine(c, toX, toY, int(p1X), int(p1Y), col, 2)
	drawThickLine(c, toX, toY, int(p2X), int(p2Y), col, 2)
}

func drawDiamond(c *gui.NativeCanvas, midX, midY, width, height int, fillCol, strokeCol gui.Color32) {
	hw := width / 2
	hh := height / 2

	// Fill diamond horizontally
	for y := -hh; y <= hh; y++ {
		span := int(float64(hw) * (1.0 - math.Abs(float64(y))/float64(hh)))
		for x := -span; x <= span; x++ {
			c.SetPixel(midX+x, midY+y, fillCol)
		}
	}

	// Outline
	drawThickLine(c, midX, midY-hh, midX+hw, midY, strokeCol, 2)
	drawThickLine(c, midX+hw, midY, midX, midY+hh, strokeCol, 2)
	drawThickLine(c, midX, midY+hh, midX-hw, midY, strokeCol, 2)
	drawThickLine(c, midX-hw, midY, midX, midY-hh, strokeCol, 2)
}

func drawParallelogram(c *gui.NativeCanvas, x, y, width, height, skew int, fillCol, strokeCol gui.Color32) {
	for currY := 0; currY < height; currY++ {
		offset := int(float64(skew) * (1.0 - float64(currY)/float64(height)))
		for currX := 0; currX < width; currX++ {
			c.SetPixel(x+offset+currX, y+currY, fillCol)
		}
	}

	drawThickLine(c, x+skew, y, x+width, y, strokeCol, 2)
	drawThickLine(c, x+width, y, x+width-skew, y+height, strokeCol, 2)
	drawThickLine(c, x+width-skew, y+height, x, y+height, strokeCol, 2)
	drawThickLine(c, x, y+height, x+skew, y, strokeCol, 2)
}
