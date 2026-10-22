package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"AxiomLang/pkg/checker"
	"AxiomLang/pkg/flow"
	"AxiomLang/pkg/lexer"
	"AxiomLang/pkg/parser"
)

type LSPRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      interface{}     `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type LSPResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Result  interface{} `json:"result,omitempty"`
	Error   interface{} `json:"error,omitempty"`
}

type PublishDiagnosticsParams struct {
	URI         string       `json:"uri"`
	Diagnostics []Diagnostic `json:"diagnostics"`
}

type Diagnostic struct {
	Range    Range  `json:"range"`
	Severity int    `json:"severity"` // 1: Error, 2: Warning
	Message  string `json:"message"`
	Source   string `json:"source"`
}

type Range struct {
	Start Position `json:"start"`
	End   Position `json:"end"`
}

type Position struct {
	Line      int `json:"line"`
	Character int `json:"character"`
}

type ServerCapabilities struct {
	TextDocumentSync int `json:"textDocumentSync"` // 1 = Full
}

type InitializeResult struct {
	Capabilities ServerCapabilities `json:"capabilities"`
}

func main() {
	reader := bufio.NewReader(os.Stdin)

	for {
		// Read headers: Content-Length: <n>\r\n\r\n
		contentLength := 0
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				if err == io.EOF {
					return
				}
				fmt.Fprintf(os.Stderr, "[LSP ERROR] Failed to read header: %v\n", err)
				return
			}
			line = strings.TrimSpace(line)
			if line == "" {
				break
			}
			if strings.HasPrefix(line, "Content-Length:") {
				parts := strings.Split(line, ":")
				if len(parts) == 2 {
					cl, _ := strconv.Atoi(strings.TrimSpace(parts[1]))
					contentLength = cl
				}
			}
		}

		if contentLength == 0 {
			continue
		}

		body := make([]byte, contentLength)
		_, err := io.ReadFull(reader, body)
		if err != nil {
			return
		}

		var req LSPRequest
		if err := json.Unmarshal(body, &req); err != nil {
			continue
		}

		handleRequest(&req)
	}
}

func handleRequest(req *LSPRequest) {
	switch req.Method {
	case "initialize":
		res := LSPResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result: InitializeResult{
				Capabilities: ServerCapabilities{
					TextDocumentSync: 1,
				},
			},
		}
		sendLSPMessage(res)

	case "textDocument/didOpen", "textDocument/didChange":
		var didOpen struct {
			TextDocument struct {
				URI  string `json:"uri"`
				Text string `json:"text"`
			} `json:"textDocument"`
		}
		if err := json.Unmarshal(req.Params, &didOpen); err == nil {
			diags := validateAxiomSource(didOpen.TextDocument.Text)
			notification := struct {
				JSONRPC string                   `json:"jsonrpc"`
				Method  string                   `json:"method"`
				Params  PublishDiagnosticsParams `json:"params"`
			}{
				JSONRPC: "2.0",
				Method:  "textDocument/publishDiagnostics",
				Params: PublishDiagnosticsParams{
					URI:         didOpen.TextDocument.URI,
					Diagnostics: diags,
				},
			}
			sendLSPNotification(notification)
		}

	case "shutdown":
		res := LSPResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result:  nil,
		}
		sendLSPMessage(res)

	case "axiom/flowchart":
		var params struct {
			Text string `json:"text"`
		}
		if err := json.Unmarshal(req.Params, &params); err == nil {
			result := generateFlowchartsJSON(params.Text)
			res := LSPResponse{
				JSONRPC: "2.0",
				ID:      req.ID,
				Result:  result,
			}
			sendLSPMessage(res)
		} else {
			res := LSPResponse{
				JSONRPC: "2.0",
				ID:      req.ID,
				Error: map[string]interface{}{
					"code":    -32602,
					"message": "Invalid params",
				},
			}
			sendLSPMessage(res)
		}

	case "exit":
		os.Exit(0)
	}
}

func validateAxiomSource(src string) []Diagnostic {
	var diagnostics []Diagnostic

	l := lexer.New(src)
	p := parser.New(l)
	prog := p.ParseProgram()

	for _, pErr := range p.Errors() {
		diagnostics = append(diagnostics, Diagnostic{
			Range: Range{
				Start: Position{Line: 0, Character: 0},
				End:   Position{Line: 0, Character: 80},
			},
			Severity: 1,
			Message:  pErr,
			Source:   "axiom-parser",
		})
	}

	sc := checker.New()
	capErrs := sc.Check(prog)
	for _, cErr := range capErrs {
		diagnostics = append(diagnostics, Diagnostic{
			Range: Range{
				Start: Position{Line: 0, Character: 0},
				End:   Position{Line: 0, Character: 80},
			},
			Severity: 1,
			Message:  cErr,
			Source:   "axiom-zero-ambient",
		})
	}

	return diagnostics
}

type FlowchartResponse struct {
	Charts []FlowchartData `json:"charts"`
}

type FlowchartData struct {
	Title   string            `json:"title"`
	SVG     string            `json:"svg"`
	Mermaid string            `json:"mermaid"`
	Nodes   []*flow.FlowNode  `json:"nodes"`
	Edges   []*flow.FlowEdge  `json:"edges"`
	Width   int               `json:"width"`
	Height  int               `json:"height"`
}

func generateFlowchartsJSON(src string) FlowchartResponse {
	l := lexer.New(src)
	p := parser.New(l)
	prog := p.ParseProgram()

	builder := flow.NewBuilder()
	charts := builder.BuildProgram(prog)

	var res FlowchartResponse
	for _, c := range charts {
		res.Charts = append(res.Charts, FlowchartData{
			Title:   c.Title,
			SVG:     c.ToSVG(),
			Mermaid: c.ToMermaid(),
			Nodes:   c.Nodes,
			Edges:   c.Edges,
			Width:   c.Width,
			Height:  c.Height,
		})
	}
	return res
}

func sendLSPMessage(msg LSPResponse) {
	data, _ := json.Marshal(msg)
	fmt.Printf("Content-Length: %d\r\n\r\n%s", len(data), string(data))
}

func sendLSPNotification(msg interface{}) {
	data, _ := json.Marshal(msg)
	fmt.Printf("Content-Length: %d\r\n\r\n%s", len(data), string(data))
}
