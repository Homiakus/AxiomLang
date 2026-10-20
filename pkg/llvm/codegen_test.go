package llvm_test

import (
	"strings"
	"testing"

	"AxiomLang/pkg/lexer"
	"AxiomLang/pkg/llvm"
	"AxiomLang/pkg/parser"
)

func TestLLVMGenerator_BareMetalKernel(t *testing.T) {
	input := `
require cap[Hardware.Port]

fn kernel_compute(x: int, y: int) -> int {
    let result = x * y + 100;
    return result;
}

intent "Инициализация ядерных структур и памяти" {
    let a = kernel_compute(5, 20);
}
`
	l := lexer.New(input)
	p := parser.New(l)
	prog := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	gen := llvm.NewLLVMGenerator(llvm.TargetX86_64BareMetal)
	ir := gen.GenerateLLVMIR(prog)

	if !strings.Contains(ir, "target triple = \"x86_64-unknown-none-elf\"") {
		t.Fatalf("missing target triple in LLVM IR")
	}

	if !strings.Contains(ir, "define void @_start()") {
		t.Fatalf("missing bare-metal entry point @_start in LLVM IR")
	}

	if !strings.Contains(ir, "define i64 @kernel_compute") {
		t.Fatalf("missing function declaration in LLVM IR")
	}

	if !strings.Contains(ir, "cli; hlt") {
		t.Fatalf("missing bare-metal halt loop assembly")
	}

	// Verify bare-metal framebuffer & VGA hardware video routines
	if !strings.Contains(ir, "define void @baremetal_fb_draw_pixel") {
		t.Fatalf("missing bare-metal hardware framebuffer routine in LLVM IR")
	}
	if !strings.Contains(ir, "define void @baremetal_vga_write_char") {
		t.Fatalf("missing bare-metal VGA text routine in LLVM IR")
	}
}
