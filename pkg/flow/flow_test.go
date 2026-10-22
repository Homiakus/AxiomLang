package flow

import (
	"strings"
	"testing"

	"AxiomLang/pkg/lexer"
	"AxiomLang/pkg/parser"
)

func TestFlowchartGeneration(t *testing.T) {
	code := `
fn calculate_score(val: int) -> int {
    let score = val * 2;
    if score > 100 {
        println("High score!");
        return score;
    } else {
        println("Try again");
    }
    return 0;
}
`
	l := lexer.New(code)
	p := parser.New(l)
	prog := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("Parse errors: %v", p.Errors())
	}

	builder := NewBuilder()
	charts := builder.BuildProgram(prog)
	if len(charts) == 0 {
		t.Fatalf("Expected at least 1 flowchart, got 0")
	}

	chart := charts[0]
	if chart.Title != "Function: calculate_score" {
		t.Errorf("Unexpected title: %s", chart.Title)
	}

	// Verify nodes
	var hasDecision, hasIO, hasTerminator bool
	for _, n := range chart.Nodes {
		if n.Kind == ShapeDecision {
			hasDecision = true
		}
		if n.Kind == ShapeIO {
			hasIO = true
		}
		if n.Kind == ShapeTerminator {
			hasTerminator = true
		}
	}

	if !hasDecision {
		t.Errorf("Expected Decision (if) node in chart")
	}
	if !hasIO {
		t.Errorf("Expected IO (println) node in chart")
	}
	if !hasTerminator {
		t.Errorf("Expected Terminator (fn start/return) node in chart")
	}

	// Verify Mermaid conversion
	mermaid := chart.ToMermaid()
	if !strings.Contains(mermaid, "flowchart TD") {
		t.Errorf("Mermaid export missing header")
	}

	// Verify SVG conversion
	svg := chart.ToSVG()
	if !strings.Contains(svg, "<svg") || !strings.Contains(svg, "</svg>") {
		t.Errorf("SVG export invalid")
	}

	// Verify Canvas rendering
	canvas := chart.RenderToCanvas()
	if canvas == nil || canvas.Width == 0 || canvas.Height == 0 {
		t.Errorf("Canvas rendering failed")
	}
}
