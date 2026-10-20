package checker_test

import (
	"testing"

	"AxiomLang/pkg/checker"
	"AxiomLang/pkg/lexer"
	"AxiomLang/pkg/parser"
)

func TestScopeChecker_ZeroAmbientViolation(t *testing.T) {
	// Program attempts to print without Console.Out capability
	inputUnauthorized := `
intent "Неавторизованный вывод" {
    print("Взлом консоли без мандата");
}
`
	l := lexer.New(inputUnauthorized)
	p := parser.New(l)
	prog := p.ParseProgram()

	sc := checker.New()
	errs := sc.Check(prog)

	if len(errs) == 0 {
		t.Fatalf("expected Zero-Ambient violation error, got 0 errors")
	}

	// Program with valid capability
	inputAuthorized := `
require cap[Console.Out]

intent "Авторизованный вывод" {
    print("Привет из песочницы");
}
`
	l2 := lexer.New(inputAuthorized)
	p2 := parser.New(l2)
	prog2 := p2.ParseProgram()

	sc2 := checker.New()
	errs2 := sc2.Check(prog2)

	if len(errs2) != 0 {
		t.Fatalf("expected 0 errors for authorized capability, got: %v", errs2)
	}
}

func TestScopeChecker_FunctionAndControlFlowAuditing(t *testing.T) {
	// 1. Violation inside a top-level function must be caught
	fnViolation := `
fn stealth_logger() {
    print("Stealth logging without capability");
}
`
	l := lexer.New(fnViolation)
	p := parser.New(l)
	prog := p.ParseProgram()

	sc := checker.New()
	errs := sc.Check(prog)
	if len(errs) == 0 {
		t.Fatalf("expected violation inside top-level function to be caught")
	}

	// 2. Violation inside an if-statement must be caught
	ifViolation := `
intent "Conditional exploit" {
    if 1 == 1 {
        print("Violation in if branch");
    }
}
`
	l = lexer.New(ifViolation)
	p = parser.New(l)
	prog = p.ParseProgram()

	errs = sc.Check(prog)
	if len(errs) == 0 {
		t.Fatalf("expected violation inside if block to be caught")
	}

	// 3. FileSystem.Read must NOT grant FileSystem.Write
	readOnlyProg := `
require cap[FileSystem.Read]
intent "Write with read cap" {
    fs.write("/etc/shadow", "data");
}
`
	l = lexer.New(readOnlyProg)
	p = parser.New(l)
	prog = p.ParseProgram()

	errs = sc.Check(prog)
	if len(errs) == 0 {
		t.Fatalf("expected FileSystem.Write to be denied when only FileSystem.Read is granted")
	}
}

func TestScopeChecker_FineGrainedCapabilities(t *testing.T) {
	sc := checker.New()

	// 1. Specific file path read isolation
	specificPathProg := `
require cap[FileSystem.Read("/var/log/app.log")]

intent "Secure Read" {
    fs.read("/etc/shadow");
}
`
	l := lexer.New(specificPathProg)
	p := parser.New(l)
	prog := p.ParseProgram()
	errs := sc.Check(prog)
	if len(errs) == 0 {
		t.Fatalf("expected error when reading /etc/shadow with only /var/log/app.log cap granted")
	}

	// 2. Network without capability
	netProg := `
intent "Unauthorized Net" {
    net.listen("0.0.0.0:8080");
}
`
	l2 := lexer.New(netProg)
	p2 := parser.New(l2)
	prog2 := p2.ParseProgram()
	errs2 := sc.Check(prog2)
	if len(errs2) == 0 {
		t.Fatalf("expected error when listening to network without Net.Listen cap")
	}

	// 3. Network with valid capability
	netAuthProg := `
require cap[Net.Listen("0.0.0.0:8080")]

intent "Authorized Net" {
    net.listen("0.0.0.0:8080");
}
`
	l3 := lexer.New(netAuthProg)
	p3 := parser.New(l3)
	prog3 := p3.ParseProgram()
	errs3 := sc.Check(prog3)
	if len(errs3) != 0 {
		t.Fatalf("expected 0 errors for authorized net listen, got: %v", errs3)
	}
}
