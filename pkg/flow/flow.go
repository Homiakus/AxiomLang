package flow

// ShapeKind defines the visual role of a flowchart block according to ISO/IEC 5807 / standard flowchart conventions.
type ShapeKind string

const (
	ShapeTerminator ShapeKind = "terminator" // Oval / Rounded: Function Start, End, Return
	ShapeProcess    ShapeKind = "process"    // Rectangle: Assignment, calculations, mutations
	ShapeDecision   ShapeKind = "decision"   // Diamond: If condition, match branching
	ShapeIO         ShapeKind = "io"         // Slanted Parallelogram: print, println, I/O operations
	ShapeSubroutine ShapeKind = "subroutine" // Predefined Process: Function calls
	ShapeLoop       ShapeKind = "loop"       // Hexagon / Diamond: For-in loops
)

// Port defines connection points on a node.
type Port string

const (
	PortTop    Port = "top"
	PortBottom Port = "bottom"
	PortLeft   Port = "left"
	PortRight  Port = "right"
)

// FlowNode represents a single visual block in the flowchart.
type FlowNode struct {
	ID        string    `json:"id"`
	Kind      ShapeKind `json:"kind"`
	Label     string    `json:"label"`
	Detail    string    `json:"detail,omitempty"`
	Line      int       `json:"line,omitempty"`
	Col       int       `json:"col,omitempty"`
	X         int       `json:"x"`
	Y         int       `json:"y"`
	Width     int       `json:"width"`
	Height    int       `json:"height"`
	Layer     int       `json:"layer"`
	IsActive  bool      `json:"is_active,omitempty"` // For live execution trace
}

// FlowEdge represents a directed connection (arrow) between blocks.
type FlowEdge struct {
	ID        string `json:"id"`
	FromID    string `json:"from_id"`
	ToID      string `json:"to_id"`
	FromPort  Port   `json:"from_port"`
	ToPort    Port   `json:"to_port"`
	Label     string `json:"label,omitempty"` // e.g., "Yes", "No", "Next"
	Color     string `json:"color,omitempty"` // e.g. green for true, red for false
	IsBackArc bool   `json:"is_back_arc"`     // Loop feedback arc
}

// Flowchart represents a complete control flow graph for a function or program.
type Flowchart struct {
	Title       string      `json:"title"`
	Nodes       []*FlowNode `json:"nodes"`
	Edges       []*FlowEdge `json:"edges"`
	Width       int         `json:"width"`
	Height      int         `json:"height"`
	nodeMap     map[string]*FlowNode
}

// NewFlowchart initializes an empty flowchart.
func NewFlowchart(title string) *Flowchart {
	return &Flowchart{
		Title:   title,
		Nodes:   make([]*FlowNode, 0),
		Edges:   make([]*FlowEdge, 0),
		nodeMap: make(map[string]*FlowNode),
	}
}

// AddNode registers a node in the chart.
func (fc *Flowchart) AddNode(node *FlowNode) {
	fc.Nodes = append(fc.Nodes, node)
	fc.nodeMap[node.ID] = node
}

// AddEdge registers a directed edge in the chart.
func (fc *Flowchart) AddEdge(edge *FlowEdge) {
	fc.Edges = append(fc.Edges, edge)
}

// GetNode retrieves a node by ID.
func (fc *Flowchart) GetNode(id string) *FlowNode {
	return fc.nodeMap[id]
}
