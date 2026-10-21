package checker

import (
	"AxiomLang/pkg/ast"
	"fmt"
	"strings"
)

type Capability struct {
	Resource string
	Target   string
}

type ScopeChecker struct {
	grantedCaps map[string]bool // key: "Resource" or "Resource:target"
	structs     map[string]*ast.StructStatement
	enums       map[string]*ast.EnumStatement
	types       map[string]string // variable -> declared/inferred type
	errors      []string
}

func New() *ScopeChecker {
	return &ScopeChecker{
		grantedCaps: make(map[string]bool),
		structs:     make(map[string]*ast.StructStatement),
		enums:       make(map[string]*ast.EnumStatement),
		types:       make(map[string]string),
		errors:      []string{},
	}
}

func (sc *ScopeChecker) Check(program *ast.Program) []string {
	sc.errors = []string{}
	sc.grantedCaps = make(map[string]bool)
	sc.structs = make(map[string]*ast.StructStatement)
	sc.enums = make(map[string]*ast.EnumStatement)
	sc.types = make(map[string]string)

	// Phase 0: Register declared Enums and Structs
	for _, en := range program.Enums {
		sc.enums[en.Name.Value] = en
	}
	for _, st := range program.Structs {
		sc.structs[st.Name.Value] = st
	}

	// Phase 1: Register declared capabilities
	for _, capStmt := range program.Capabilities {
		key := capStmt.Resource
		if capStmt.Arg != nil {
			if strLit, ok := capStmt.Arg.(*ast.StringLiteral); ok {
				key = fmt.Sprintf("%s:%s", capStmt.Resource, strLit.Value)
			}
		}
		sc.grantedCaps[key] = true

		// Allow Console wildcard only if explicitly "Console" or "Console.Out"
		if capStmt.Resource == "Console.Out" || capStmt.Resource == "Console" {
			sc.grantedCaps["Console.Out"] = true
			sc.grantedCaps["Console"] = true
		}
		// Notice: Do NOT automatically grant parent wildcard for sensitive namespaces like FileSystem!
		// "FileSystem.Read" MUST NOT grant "FileSystem.Write" or blanket "FileSystem"!
		if capStmt.Resource == "FileSystem" {
			sc.grantedCaps["FileSystem.Read"] = true
			sc.grantedCaps["FileSystem.Write"] = true
		}
		if capStmt.Resource == "Actor" {
			sc.grantedCaps["Actor.Spawn"] = true
			sc.grantedCaps["Actor.Send"] = true
		}
		if capStmt.Resource == "Net" {
			sc.grantedCaps["Net.Listen"] = true
			sc.grantedCaps["Net.Connect"] = true
			sc.grantedCaps["Net.Send"] = true
			sc.grantedCaps["Net.Receive"] = true
		}
		if capStmt.Resource == "Hardware" || capStmt.Resource == "Hardware.BareMetal" {
			sc.grantedCaps["Hardware.BareMetal"] = true
			sc.grantedCaps["Hardware.MMIO"] = true
			sc.grantedCaps["Hardware.PortIO"] = true
		}
	}

	// Phase 2: Check functions, intents, and statements for capability violations
	for _, fn := range program.Functions {
		sc.checkBlock(fn.Body)
	}

	for _, intent := range program.Intents {
		sc.checkBlock(intent.Body)
	}

	for _, stmt := range program.Statements {
		sc.checkStatement(stmt)
	}

	return sc.errors
}

func (sc *ScopeChecker) checkBlock(block *ast.BlockStatement) {
	if block == nil {
		return
	}
	for _, stmt := range block.Statements {
		sc.checkStatement(stmt)
	}
}

func (sc *ScopeChecker) checkStatement(stmt ast.Statement) {
	if stmt == nil {
		return
	}
	switch s := stmt.(type) {
	case *ast.ExpressionStatement:
		sc.checkExpression(s.Expression)
	case *ast.LetStatement:
		valType := sc.inferType(s.Value)
		sc.types[s.Name.Value] = valType
		sc.checkExpression(s.Value)
	case *ast.MatchStatement:
		sc.checkExpression(s.Value)
		for _, c := range s.Cases {
			sc.checkExpression(c.Pattern)
			sc.checkBlock(c.Body)
		}
	case *ast.ReturnStatement:
		if s.ReturnValue != nil {
			sc.checkExpression(s.ReturnValue)
		}
	case *ast.IfStatement:
		sc.checkExpression(s.Condition)
		sc.checkBlock(s.Consequence)
		if s.Alternative != nil {
			sc.checkBlock(s.Alternative)
		}
	case *ast.ForInStatement:
		sc.checkExpression(s.Collection)
		sc.checkBlock(s.Body)
	case *ast.BlockStatement:
		sc.checkBlock(s)
	case *ast.SpawnStatement:
		if !sc.hasCap("Actor.Spawn") && !sc.hasCap("Actor") {
			sc.errors = append(sc.errors, fmt.Sprintf("E_ZERO_AMBIENT_VIOLATION: Spawning actor '%s' requires ungranted capability 'Actor.Spawn'", s.ActorID.Value))
		}
		if s.Payload != nil {
			sc.checkExpression(s.Payload)
		}
	case *ast.SendStatement:
		if !sc.hasCap("Actor.Send") && !sc.hasCap("Actor.Spawn") && !sc.hasCap("Actor") {
			sc.errors = append(sc.errors, fmt.Sprintf("E_ZERO_AMBIENT_VIOLATION: Sending to actor '%s' requires ungranted capability 'Actor.Send'", s.Target.Value))
		}
		sc.checkExpression(s.Message)
	}
}

func (sc *ScopeChecker) checkExpression(expr ast.Expression) {
	if expr == nil {
		return
	}

	switch e := expr.(type) {
	case *ast.CallExpression:
		sc.verifyCallAuthority(e)
		for _, arg := range e.Arguments {
			sc.checkExpression(arg)
		}
	case *ast.BinaryExpression:
		sc.checkExpression(e.Left)
		sc.checkExpression(e.Right)
	case *ast.PrefixExpression:
		sc.checkExpression(e.Right)
	case *ast.ArrayLiteral:
		for _, el := range e.Elements {
			sc.checkExpression(el)
		}
	case *ast.MapLiteral:
		for k, v := range e.Pairs {
			sc.checkExpression(k)
			sc.checkExpression(v)
		}
	case *ast.MemberAccessExpression:
		sc.checkExpression(e.Object)
	}
}

func (sc *ScopeChecker) verifyCallAuthority(call *ast.CallExpression) {
	if call == nil || call.Function == nil {
		return
	}
	callStr := call.Function.String()

	// Console operations
	if callStr == "print" || callStr == "println" || callStr == "log" {
		if !sc.hasCap("Console.Out") && !sc.hasCap("Console") {
			sc.errors = append(sc.errors, fmt.Sprintf("E_ZERO_AMBIENT_VIOLATION: Call to '%s' requires ungranted capability 'Console.Out'", callStr))
		}
	}

	// FileSystem operations
	if strings.HasPrefix(callStr, "fs.") {
		action := strings.TrimPrefix(callStr, "fs.")
		if action == "write" || action == "append" {
			if len(call.Arguments) > 0 {
				if pathArg, ok := call.Arguments[0].(*ast.StringLiteral); ok {
					specificKey := fmt.Sprintf("FileSystem.Write:%s", pathArg.Value)
					if !sc.grantedCaps[specificKey] && !sc.hasCap("FileSystem.Write") && !sc.hasCap("FileSystem") {
						sc.errors = append(sc.errors, fmt.Sprintf("E_ZERO_AMBIENT_VIOLATION: fs.%s for target '%s' requires ungranted capability 'FileSystem.Write(%q)'", action, pathArg.Value, pathArg.Value))
					}
					return
				}
			}
		} else if action == "read" {
			if len(call.Arguments) > 0 {
				if pathArg, ok := call.Arguments[0].(*ast.StringLiteral); ok {
					specificKey := fmt.Sprintf("FileSystem.Read:%s", pathArg.Value)
					if !sc.grantedCaps[specificKey] && !sc.hasCap("FileSystem.Read") && !sc.hasCap("FileSystem") {
						sc.errors = append(sc.errors, fmt.Sprintf("E_ZERO_AMBIENT_VIOLATION: fs.read for target '%s' requires ungranted capability 'FileSystem.Read(%q)'", pathArg.Value, pathArg.Value))
					}
					return
				}
			}
			if !sc.hasCap("FileSystem.Read") && !sc.hasCap("FileSystem") {
				sc.errors = append(sc.errors, "E_ZERO_AMBIENT_VIOLATION: File system access requires ungranted capability 'FileSystem.Read'")
			}
		}
	}

	// GUI operations
	if strings.HasPrefix(callStr, "gui.") {
		if !sc.hasCap("GUI.Render") && !sc.hasCap("GUI") && !sc.hasCap("Hardware.BareMetal") {
			sc.errors = append(sc.errors, fmt.Sprintf("E_ZERO_AMBIENT_VIOLATION: Call to '%s' requires ungranted capability 'GUI.Render'", callStr))
		}
	}

	// Network operations
	if strings.HasPrefix(callStr, "net.") {
		action := strings.TrimPrefix(callStr, "net.")
		switch action {
		case "listen":
			if len(call.Arguments) > 0 {
				if addrArg, ok := call.Arguments[0].(*ast.StringLiteral); ok {
					specificKey := fmt.Sprintf("Net.Listen:%s", addrArg.Value)
					if !sc.grantedCaps[specificKey] && !sc.hasCap("Net.Listen") && !sc.hasCap("Net") {
						sc.errors = append(sc.errors, fmt.Sprintf("E_ZERO_AMBIENT_VIOLATION: net.listen on '%s' requires ungranted capability 'Net.Listen(%q)'", addrArg.Value, addrArg.Value))
					}
					return
				}
			}
			if !sc.hasCap("Net.Listen") && !sc.hasCap("Net") {
				sc.errors = append(sc.errors, "E_ZERO_AMBIENT_VIOLATION: Network listening requires ungranted capability 'Net.Listen'")
			}
		case "connect", "send", "dial":
			if !sc.hasCap("Net.Connect") && !sc.hasCap("Net.Send") && !sc.hasCap("Net") {
				sc.errors = append(sc.errors, fmt.Sprintf("E_ZERO_AMBIENT_VIOLATION: Network operation '%s' requires ungranted capability 'Net.Connect'", callStr))
			}
		}
	}

	// Hardware Direct MMIO / Port I/O
	if callStr == "draw_pixel" || callStr == "vga_char" || strings.HasPrefix(callStr, "hw.") {
		if !sc.hasCap("Hardware.BareMetal") && !sc.hasCap("Hardware.MMIO") && !sc.hasCap("GUI.Render") && !sc.hasCap("GUI") {
			sc.errors = append(sc.errors, fmt.Sprintf("E_ZERO_AMBIENT_VIOLATION: Direct hardware instruction '%s' requires ungranted capability 'Hardware.BareMetal'", callStr))
		}
	}

	// UART / Modbus operations
	if strings.HasPrefix(callStr, "uart.") {
		if !sc.hasCap("Hardware.UART") && !sc.hasCap("Hardware.BareMetal") && !sc.hasCap("Hardware") {
			sc.errors = append(sc.errors, fmt.Sprintf("E_ZERO_AMBIENT_VIOLATION: UART serial communication '%s' requires ungranted capability 'Hardware.UART'", callStr))
		}
	}
	if strings.HasPrefix(callStr, "modbus.") {
		if !sc.hasCap("Hardware.Modbus") && !sc.hasCap("Net") && !sc.hasCap("Hardware") {
			sc.errors = append(sc.errors, fmt.Sprintf("E_ZERO_AMBIENT_VIOLATION: Industrial Modbus protocol '%s' requires ungranted capability 'Hardware.Modbus'", callStr))
		}
	}
}

func (sc *ScopeChecker) hasCap(res string) bool {
	return sc.grantedCaps[res]
}

// inferType performs static type inference across AST expression nodes
func (sc *ScopeChecker) inferType(expr ast.Expression) string {
	if expr == nil {
		return "void"
	}
	switch e := expr.(type) {
	case *ast.IntegerLiteral:
		return "int"
	case *ast.StringLiteral:
		return "string"
	case *ast.BooleanLiteral:
		return "bool"
	case *ast.ArrayLiteral:
		if len(e.Elements) > 0 {
			return "[" + sc.inferType(e.Elements[0]) + "]"
		}
		return "[]"
	case *ast.MapLiteral:
		return "map"
	case *ast.Identifier:
		if t, ok := sc.types[e.Value]; ok {
			return t
		}
		return "unknown"
	case *ast.BinaryExpression:
		lt := sc.inferType(e.Left)
		rt := sc.inferType(e.Right)
		if e.Operator == "==" || e.Operator == "!=" || e.Operator == "<" || e.Operator == ">" || e.Operator == "<=" || e.Operator == ">=" {
			return "bool"
		}
		if lt == "int" && rt == "int" {
			return "int"
		}
		if lt == "string" || rt == "string" {
			return "string"
		}
		return lt
	default:
		return "any"
	}
}

// AutoGrantInferred analyzes missing capabilities and provisions them into the program's AST
func (sc *ScopeChecker) AutoGrantInferred(program *ast.Program) {
	for _, errStr := range sc.errors {
		// Look for "'Resource.Name'" in "requires ungranted capability 'Resource.Name'"
		idx := strings.Index(errStr, "capability '")
		if idx != -1 {
			sub := errStr[idx+12:]
			endQuote := strings.Index(sub, "'")
			if endQuote != -1 {
				capName := sub[:endQuote]
				alreadyHas := false
				for _, c := range program.Capabilities {
					if c.Resource == capName {
						alreadyHas = true
						break
					}
				}
				if !alreadyHas {
					program.Capabilities = append(program.Capabilities, &ast.CapabilityStatement{
						Resource: capName,
					})
				}
			}
		}
	}
}
