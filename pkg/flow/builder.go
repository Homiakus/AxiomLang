package flow

import (
	"fmt"
	"strings"

	"AxiomLang/pkg/ast"
)

// Builder extracts Flowcharts from an AST Program or FunctionStatement without any syntax annotations.
type Builder struct {
	nodeCounter int
	edgeCounter int
}

// NewBuilder creates a new CFG Flowchart builder.
func NewBuilder() *Builder {
	return &Builder{}
}

func (b *Builder) nextNodeID(prefix string) string {
	b.nodeCounter++
	return fmt.Sprintf("%s_%d", prefix, b.nodeCounter)
}

func (b *Builder) nextEdgeID() string {
	b.edgeCounter++
	return fmt.Sprintf("edge_%d", b.edgeCounter)
}

// BuildProgram creates flowcharts for all functions in the program.
func (b *Builder) BuildProgram(prog *ast.Program) []*Flowchart {
	var charts []*Flowchart
	if prog == nil {
		return charts
	}

	for _, fn := range prog.Functions {
		chart := b.BuildFunction(fn)
		if chart != nil {
			charts = append(charts, chart)
		}
	}

	// Also build top-level statements if any (script mode)
	if len(prog.Statements) > 0 {
		topChart := NewFlowchart("Main Script")
		startNode := &FlowNode{
			ID:     b.nextNodeID("start"),
			Kind:   ShapeTerminator,
			Label:  "Start",
			Width:  140,
			Height: 44,
		}
		topChart.AddNode(startNode)

		lastNodeID := b.buildStatements(topChart, prog.Statements, startNode.ID)

		endNode := &FlowNode{
			ID:     b.nextNodeID("end"),
			Kind:   ShapeTerminator,
			Label:  "End",
			Width:  140,
			Height: 44,
		}
		topChart.AddNode(endNode)
		if lastNodeID != "" {
			topChart.AddEdge(&FlowEdge{
				ID:       b.nextEdgeID(),
				FromID:   lastNodeID,
				ToID:     endNode.ID,
				FromPort: PortBottom,
				ToPort:   PortTop,
			})
		}
		LayoutFlowchart(topChart)
		charts = append(charts, topChart)
	}

	return charts
}

// BuildFunction generates a flowchart for a given FunctionStatement.
func (b *Builder) BuildFunction(fn *ast.FunctionStatement) *Flowchart {
	if fn == nil {
		return nil
	}

	name := "anonymous"
	if fn.Name != nil {
		name = fn.Name.Value
	}

	chart := NewFlowchart(fmt.Sprintf("Function: %s", name))

	var paramStrs []string
	for _, p := range fn.Parameters {
		paramStrs = append(paramStrs, fmt.Sprintf("%s: %s", p.Name.Value, p.Type))
	}
	startLabel := fmt.Sprintf("Start: %s(%s)", name, strings.Join(paramStrs, ", "))
	if len(startLabel) > 36 {
		startLabel = fmt.Sprintf("Start: %s(...)", name)
	}

	startNode := &FlowNode{
		ID:     b.nextNodeID("start"),
		Kind:   ShapeTerminator,
		Label:  startLabel,
		Line:   fn.Token.Position.Line,
		Col:    fn.Token.Position.Column,
		Width:  180,
		Height: 46,
	}
	chart.AddNode(startNode)

	lastNodeID := startNode.ID
	if fn.Body != nil && len(fn.Body.Statements) > 0 {
		lastNodeID = b.buildStatements(chart, fn.Body.Statements, startNode.ID)
	}

	endNode := &FlowNode{
		ID:     b.nextNodeID("end"),
		Kind:   ShapeTerminator,
		Label:  fmt.Sprintf("End: %s", name),
		Width:  160,
		Height: 46,
	}
	chart.AddNode(endNode)

	if lastNodeID != "" && chart.GetNode(lastNodeID) != nil && chart.GetNode(lastNodeID).Kind != ShapeTerminator {
		chart.AddEdge(&FlowEdge{
			ID:       b.nextEdgeID(),
			FromID:   lastNodeID,
			ToID:     endNode.ID,
			FromPort: PortBottom,
			ToPort:   PortTop,
		})
	}

	LayoutFlowchart(chart)
	return chart
}

func (b *Builder) buildStatements(chart *Flowchart, stmts []ast.Statement, prevID string) string {
	currID := prevID

	for _, stmt := range stmts {
		switch s := stmt.(type) {
		case *ast.LetStatement:
			label := fmt.Sprintf("%s = %s", s.Name.Value, s.Value.String())
			node := &FlowNode{
				ID:     b.nextNodeID("let"),
				Kind:   ShapeProcess,
				Label:  truncateLabel(label, 32),
				Detail: label,
				Line:   s.Token.Position.Line,
				Col:    s.Token.Position.Column,
				Width:  160,
				Height: 50,
			}
			chart.AddNode(node)
			if currID != "" {
				chart.AddEdge(&FlowEdge{
					ID:       b.nextEdgeID(),
					FromID:   currID,
					ToID:     node.ID,
					FromPort: PortBottom,
					ToPort:   PortTop,
				})
			}
			currID = node.ID

		case *ast.ReturnStatement:
			retVal := ""
			if s.ReturnValue != nil {
				retVal = " " + s.ReturnValue.String()
			}
			label := fmt.Sprintf("Return%s", retVal)
			node := &FlowNode{
				ID:     b.nextNodeID("return"),
				Kind:   ShapeTerminator,
				Label:  truncateLabel(label, 28),
				Detail: label,
				Line:   s.Token.Position.Line,
				Col:    s.Token.Position.Column,
				Width:  150,
				Height: 44,
			}
			chart.AddNode(node)
			if currID != "" {
				chart.AddEdge(&FlowEdge{
					ID:       b.nextEdgeID(),
					FromID:   currID,
					ToID:     node.ID,
					FromPort: PortBottom,
					ToPort:   PortTop,
				})
			}
			currID = node.ID

		case *ast.IfStatement:
			condLabel := "if " + s.Condition.String()
			diamondNode := &FlowNode{
				ID:     b.nextNodeID("cond"),
				Kind:   ShapeDecision,
				Label:  truncateLabel(condLabel, 26),
				Detail: condLabel,
				Line:   s.Token.Position.Line,
				Col:    s.Token.Position.Column,
				Width:  160,
				Height: 64,
			}
			chart.AddNode(diamondNode)
			if currID != "" {
				chart.AddEdge(&FlowEdge{
					ID:       b.nextEdgeID(),
					FromID:   currID,
					ToID:     diamondNode.ID,
					FromPort: PortBottom,
					ToPort:   PortTop,
				})
			}

			// Branch True (consequence)
			var trueLastID string
			if s.Consequence != nil && len(s.Consequence.Statements) > 0 {
				firstTrueID := b.buildStatements(chart, s.Consequence.Statements[:1], "")
				chart.AddEdge(&FlowEdge{
					ID:       b.nextEdgeID(),
					FromID:   diamondNode.ID,
					ToID:     firstTrueID,
					FromPort: PortLeft,
					ToPort:   PortTop,
					Label:    "True",
					Color:    "#10b981",
				})
				if len(s.Consequence.Statements) > 1 {
					trueLastID = b.buildStatements(chart, s.Consequence.Statements[1:], firstTrueID)
				} else {
					trueLastID = firstTrueID
				}
			}

			// Branch False (alternative)
			var falseLastID string
			if s.Alternative != nil && len(s.Alternative.Statements) > 0 {
				firstFalseID := b.buildStatements(chart, s.Alternative.Statements[:1], "")
				chart.AddEdge(&FlowEdge{
					ID:       b.nextEdgeID(),
					FromID:   diamondNode.ID,
					ToID:     firstFalseID,
					FromPort: PortRight,
					ToPort:   PortTop,
					Label:    "False",
					Color:    "#ef4444",
				})
				if len(s.Alternative.Statements) > 1 {
					falseLastID = b.buildStatements(chart, s.Alternative.Statements[1:], firstFalseID)
				} else {
					falseLastID = firstFalseID
				}
			}

			// Join node if both branches rejoin
			joinNode := &FlowNode{
				ID:     b.nextNodeID("join"),
				Kind:   ShapeProcess,
				Label:  "Merge",
				Width:  80,
				Height: 30,
			}
			chart.AddNode(joinNode)

			if trueLastID != "" && chart.GetNode(trueLastID).Kind != ShapeTerminator {
				chart.AddEdge(&FlowEdge{
					ID:       b.nextEdgeID(),
					FromID:   trueLastID,
					ToID:     joinNode.ID,
					FromPort: PortBottom,
					ToPort:   PortTop,
				})
			} else if s.Consequence == nil || len(s.Consequence.Statements) == 0 {
				chart.AddEdge(&FlowEdge{
					ID:       b.nextEdgeID(),
					FromID:   diamondNode.ID,
					ToID:     joinNode.ID,
					FromPort: PortLeft,
					ToPort:   PortTop,
					Label:    "True",
				})
			}

			if falseLastID != "" && chart.GetNode(falseLastID).Kind != ShapeTerminator {
				chart.AddEdge(&FlowEdge{
					ID:       b.nextEdgeID(),
					FromID:   falseLastID,
					ToID:     joinNode.ID,
					FromPort: PortBottom,
					ToPort:   PortTop,
				})
			} else {
				// No else branch -> direct edge from diamond to join
				chart.AddEdge(&FlowEdge{
					ID:       b.nextEdgeID(),
					FromID:   diamondNode.ID,
					ToID:     joinNode.ID,
					FromPort: PortRight,
					ToPort:   PortTop,
					Label:    "False",
					Color:    "#ef4444",
				})
			}

			currID = joinNode.ID

		case *ast.ForInStatement:
			loopLabel := fmt.Sprintf("for %s in %s", s.Item.Value, s.Collection.String())
			loopNode := &FlowNode{
				ID:     b.nextNodeID("loop"),
				Kind:   ShapeLoop,
				Label:  truncateLabel(loopLabel, 26),
				Detail: loopLabel,
				Line:   s.Token.Position.Line,
				Col:    s.Token.Position.Column,
				Width:  160,
				Height: 60,
			}
			chart.AddNode(loopNode)
			if currID != "" {
				chart.AddEdge(&FlowEdge{
					ID:       b.nextEdgeID(),
					FromID:   currID,
					ToID:     loopNode.ID,
					FromPort: PortBottom,
					ToPort:   PortTop,
				})
			}

			// Loop body
			if s.Body != nil && len(s.Body.Statements) > 0 {
				firstBodyID := b.buildStatements(chart, s.Body.Statements[:1], "")
				chart.AddEdge(&FlowEdge{
					ID:       b.nextEdgeID(),
					FromID:   loopNode.ID,
					ToID:     firstBodyID,
					FromPort: PortBottom,
					ToPort:   PortTop,
					Label:    "Each Item",
				})
				var lastBodyID string
				if len(s.Body.Statements) > 1 {
					lastBodyID = b.buildStatements(chart, s.Body.Statements[1:], firstBodyID)
				} else {
					lastBodyID = firstBodyID
				}
				// Back-edge from body back to loop header
				if lastBodyID != "" {
					chart.AddEdge(&FlowEdge{
						ID:        b.nextEdgeID(),
						FromID:    lastBodyID,
						ToID:      loopNode.ID,
						FromPort:  PortLeft,
						ToPort:    PortLeft,
						Label:     "Loop",
						IsBackArc: true,
					})
				}
			}

			currID = loopNode.ID

		case *ast.ExpressionStatement:
			// Check if expression is an I/O operation (e.g. print/println) or Subroutine Call
			exprStr := s.Expression.String()
			kind := ShapeProcess
			if isIOOperation(s.Expression) {
				kind = ShapeIO
			} else if isCallOperation(s.Expression) {
				kind = ShapeSubroutine
			}

			node := &FlowNode{
				ID:     b.nextNodeID("expr"),
				Kind:   kind,
				Label:  truncateLabel(exprStr, 30),
				Detail: exprStr,
				Line:   s.Token.Position.Line,
				Col:    s.Token.Position.Column,
				Width:  160,
				Height: 50,
			}
			chart.AddNode(node)
			if currID != "" {
				chart.AddEdge(&FlowEdge{
					ID:       b.nextEdgeID(),
					FromID:   currID,
					ToID:     node.ID,
					FromPort: PortBottom,
					ToPort:   PortTop,
				})
			}
			currID = node.ID

		default:
			// Generic statement fallback
			stmtStr := stmt.String()
			node := &FlowNode{
				ID:     b.nextNodeID("stmt"),
				Kind:   ShapeProcess,
				Label:  truncateLabel(stmtStr, 28),
				Detail: stmtStr,
				Line:   0,
				Col:    0,
				Width:  150,
				Height: 46,
			}
			chart.AddNode(node)
			if currID != "" {
				chart.AddEdge(&FlowEdge{
					ID:       b.nextEdgeID(),
					FromID:   currID,
					ToID:     node.ID,
					FromPort: PortBottom,
					ToPort:   PortTop,
				})
			}
			currID = node.ID
		}
	}

	return currID
}

func isIOOperation(expr ast.Expression) bool {
	if call, ok := expr.(*ast.CallExpression); ok {
		fnName := call.Function.String()
		if fnName == "println" || fnName == "print" || strings.Contains(fnName, "write") || strings.Contains(fnName, "read") || strings.HasPrefix(fnName, "io.") {
			return true
		}
	}
	return false
}

func isCallOperation(expr ast.Expression) bool {
	_, ok := expr.(*ast.CallExpression)
	return ok
}

func truncateLabel(s string, maxLen int) string {
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.TrimSpace(s)
	if len(s) > maxLen {
		return s[:maxLen-3] + "..."
	}
	return s
}
