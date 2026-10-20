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

func TestLLVMGenerator_EmbeddedMCUTargets(t *testing.T) {
	input := `
require cap[Hardware.Port]

fn gpio_write(pin: int, val: int) -> int {
    mmio_write(1073807360, val);
    return mmio_read(1073807360);
}

intent "MCU Peripheral and Interrupt Loop" {
    mcu.disable_irq();
    gpio_write(13, 1);
    mcu.enable_irq();
}
`
	l := lexer.New(input)
	p := parser.New(l)
	prog := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	// Test Cortex-M / Thumb
	genARM := llvm.NewLLVMGenerator(llvm.TargetARMV7MBareMetal)
	irARM := genARM.GenerateLLVMIR(prog)
	if !strings.Contains(irARM, "thumbv7em-none-eabihf") {
		t.Fatalf("missing Cortex-M thumbv7em target triple")
	}
	if !strings.Contains(irARM, "store volatile i32") || !strings.Contains(irARM, "load volatile i32") {
		t.Fatalf("missing MMIO volatile operations in generated LLVM IR")
	}
	if !strings.Contains(irARM, "cpsie i") || !strings.Contains(irARM, "cpsid i") {
		t.Fatalf("missing IRQ enable/disable intrinsics")
	}

	// Test RISC-V 32 IMC
	genRV32 := llvm.NewLLVMGenerator(llvm.TargetRISCV32BareMetal)
	irRV32 := genRV32.GenerateLLVMIR(prog)
	if !strings.Contains(irRV32, "riscv32imc-unknown-none-elf") {
		t.Fatalf("missing RISC-V 32 target triple")
	}
}
