package llvm

import (
	"AxiomLang/pkg/ast"
	"bytes"
	"fmt"
	"strings"
)

// TargetTriple defines the target architecture for freestanding execution
type TargetTriple string

const (
	TargetX86_64BareMetal  TargetTriple = "x86_64-unknown-none-elf"
	TargetAArch64BareMetal TargetTriple = "aarch64-unknown-none"
	TargetRISCV64BareMetal TargetTriple = "riscv64gc-unknown-none-elf"
	TargetRISCV32BareMetal TargetTriple = "riscv32imc-unknown-none-elf"
	TargetARMV7MBareMetal  TargetTriple = "thumbv7em-none-eabihf"
	TargetCortexM0BareMetal TargetTriple = "thumbv6m-none-eabi"
)

// LLVMGenerator translates Axiom AST directly into freestanding LLVM IR (no OS runtime required)
type LLVMGenerator struct {
	target      TargetTriple
	dataLayout  string
	varCounter  int
	strLiterals []string
}

func NewLLVMGenerator(target TargetTriple) *LLVMGenerator {
	dl := "e-m:e-p270:32:32-p271:32:32-p272:64:64-i64:64-f80:128-n8:16:32:64-S128"
	if target == TargetAArch64BareMetal {
		dl = "e-m:e-i8:8:32-i16:16:32-i64:64-i128:128-n32:64-S128"
	} else if target == TargetARMV7MBareMetal || target == TargetCortexM0BareMetal {
		dl = "e-m:e-p:32:32-Fi8-i64:64-v128:64:128-a:0:32-n32-S64"
	} else if target == TargetRISCV32BareMetal {
		dl = "e-m:e-p:32:32-i64:64-n32-S128"
	} else if target == TargetRISCV64BareMetal {
		dl = "e-m:e-p:64:64-i64:64-i128:128-n32:64-S128"
	}
	return &LLVMGenerator{
		target:     target,
		dataLayout: dl,
	}
}

// GenerateLLVMIR emits valid textual LLVM IR assembly
func (g *LLVMGenerator) GenerateLLVMIR(program *ast.Program) string {
	var buf bytes.Buffer

	// Header & Target specifications
	buf.WriteString(fmt.Sprintf("; ModuleID = 'AxiomBareMetalKernel'\nsource_filename = \"kernel.axl\"\ntarget datalayout = \"%s\"\ntarget triple = \"%s\"\n\n", g.dataLayout, g.target))

	// Freestanding intrinsic declarations
	buf.WriteString("; Freestanding Kernel Intrinsics\n")
	buf.WriteString("declare void @llvm.x86.sse.pause() nounwind\n")
	buf.WriteString("declare void @llvm.donothing() nounwind readnone\n\n")

	// Emit bare-metal hardware drawing routines:
	buf.WriteString("; Hardware Video Memory Mapped I/O Framebuffer Routines (Zero OS / Freestanding)\n")
	buf.WriteString("; Writes a 32-bit pixel directly to physical linear framebuffer address: addr + (y * pitch) + (x * 4)\n")
	buf.WriteString("define void @baremetal_fb_draw_pixel(i64 %fb_base, i64 %pitch, i64 %x, i64 %y, i32 %color) nounwind {\n")
	buf.WriteString("entry:\n")
	buf.WriteString("  %row_off = mul i64 %y, %pitch\n")
	buf.WriteString("  %col_off = shl i64 %x, 2\n")
	buf.WriteString("  %tot_off = add i64 %row_off, %col_off\n")
	buf.WriteString("  %pix_addr = add i64 %fb_base, %tot_off\n")
	buf.WriteString("  %ptr = inttoptr i64 %pix_addr to i32*\n")
	buf.WriteString("  store i32 %color, i32* %ptr, align 4\n")
	buf.WriteString("  ret void\n")
	buf.WriteString("}\n\n")

	buf.WriteString("; Writes ASCII character to standard VGA Text Buffer 0xB8000 (Ring 0)\n")
	buf.WriteString("define void @baremetal_vga_write_char(i64 %col, i64 %row, i8 %ch, i8 %color_attr) nounwind {\n")
	buf.WriteString("entry:\n")
	buf.WriteString("  %row_bytes = mul i64 %row, 160\n")
	buf.WriteString("  %col_bytes = shl i64 %col, 1\n")
	buf.WriteString("  %off = add i64 %row_bytes, %col_bytes\n")
	buf.WriteString("  %vga_addr = add i64 753664, %off ; 0xB8000 = 753664\n")
	buf.WriteString("  %ch_ptr = inttoptr i64 %vga_addr to i8*\n")
	buf.WriteString("  store i8 %ch, i8* %ch_ptr, align 1\n")
	buf.WriteString("  %attr_addr = add i64 %vga_addr, 1\n")
	buf.WriteString("  %attr_ptr = inttoptr i64 %attr_addr to i8*\n")
	buf.WriteString("  store i8 %color_attr, i8* %attr_ptr, align 1\n")
	buf.WriteString("  ret void\n")
	buf.WriteString("}\n\n")

	// Emit structs as LLVM types
	for _, st := range program.Structs {
		var fieldTypes []string
		for _, f := range st.Fields {
			fieldTypes = append(fieldTypes, g.mapTypeToLLVM(f.Type))
		}
		buf.WriteString(fmt.Sprintf("%%%s = type { %s }\n", st.Name.Value, strings.Join(fieldTypes, ", ")))
	}
	buf.WriteString("\n")

	// Emit user functions
	for _, fn := range program.Functions {
		g.emitFunction(&buf, fn)
		buf.WriteString("\n")
	}

	// Emit Kernel Entry Point _start
	buf.WriteString("; Kernel Entry Point for Bootloader (Multiboot / Limine / UEFI)\ndefine void @_start() noreturn nounwind section \".text.boot\" {\nentry:\n")
	
	// Execute top-level intents as bare-metal initialization phases
	for _, intent := range program.Intents {
		buf.WriteString(fmt.Sprintf("  ; Phase Intent: %s\n", intent.Description))
		for _, stmt := range intent.Body.Statements {
			g.emitStatement(&buf, stmt)
		}
	}

	buf.WriteString("  ; Bare-metal halt loop: cli; hlt\n")
	buf.WriteString("  call void asm sideeffect \"cli; hlt\", \"~{memory}\"() nounwind\n")
	buf.WriteString("  unreachable\n")
	buf.WriteString("}\n")

	return buf.String()
}

func (g *LLVMGenerator) emitFunction(buf *bytes.Buffer, fn *ast.FunctionStatement) {
	retType := g.mapTypeToLLVM(fn.ReturnType)
	var params []string
	for _, p := range fn.Parameters {
		params = append(params, fmt.Sprintf("%s %%%s", g.mapTypeToLLVM(p.Type), p.Name.Value))
	}

	buf.WriteString(fmt.Sprintf("define %s @%s(%s) nounwind {\nentry:\n", retType, fn.Name.Value, strings.Join(params, ", ")))
	for _, stmt := range fn.Body.Statements {
		g.emitStatement(buf, stmt)
	}
	if retType == "void" && !strings.Contains(buf.String(), "ret void") {
		buf.WriteString("  ret void\n")
	}
	buf.WriteString("}\n")
}

func (g *LLVMGenerator) emitStatement(buf *bytes.Buffer, stmt ast.Statement) {
	switch s := stmt.(type) {
	case *ast.LetStatement:
		val := g.emitExpression(buf, s.Value)
		buf.WriteString(fmt.Sprintf("  %%%s = alloca i64, align 8\n", s.Name.Value))
		buf.WriteString(fmt.Sprintf("  store i64 %s, i64* %%%s, align 8\n", val, s.Name.Value))
	case *ast.ReturnStatement:
		if s.ReturnValue != nil {
			val := g.emitExpression(buf, s.ReturnValue)
			buf.WriteString(fmt.Sprintf("  ret i64 %s\n", val))
		} else {
			buf.WriteString("  ret void\n")
		}
	case *ast.ExpressionStatement:
		_ = g.emitExpression(buf, s.Expression)
	}
}

func (g *LLVMGenerator) emitExpression(buf *bytes.Buffer, expr ast.Expression) string {
	if expr == nil {
		return "0"
	}

	switch e := expr.(type) {
	case *ast.IntegerLiteral:
		return fmt.Sprintf("%d", e.Value)
	case *ast.Identifier:
		g.varCounter++
		tmp := fmt.Sprintf("%%t%d", g.varCounter)
		buf.WriteString(fmt.Sprintf("  %s = load i64, i64* %%%s, align 8\n", tmp, e.Value))
		return tmp
	case *ast.BinaryExpression:
		left := g.emitExpression(buf, e.Left)
		right := g.emitExpression(buf, e.Right)
		g.varCounter++
		tmp := fmt.Sprintf("%%t%d", g.varCounter)
		op := "add"
		switch e.Operator {
		case "+":
			op = "add"
		case "-":
			op = "sub"
		case "*":
			op = "mul"
		case "/":
			op = "sdiv"
		}
		buf.WriteString(fmt.Sprintf("  %s = %s i64 %s, %s\n", tmp, op, left, right))
		return tmp
	case *ast.CallExpression:
		fnName := e.Function.String()
		if fnName == "print" || fnName == "outb" || fnName == "mmio_write" {
			// Bare-metal hardware port write or memory mapped I/O
			if fnName == "mmio_write" && len(e.Arguments) >= 2 {
				addr := g.emitExpression(buf, e.Arguments[0])
				val := g.emitExpression(buf, e.Arguments[1])
				g.varCounter++
				ptr := fmt.Sprintf("%%ptr%d", g.varCounter)
				buf.WriteString(fmt.Sprintf("  %s = inttoptr i64 %s to i32*\n", ptr, addr))
				buf.WriteString(fmt.Sprintf("  store volatile i32 %s, i32* %s, align 4\n", val, ptr))
				return "0"
			}
			buf.WriteString(fmt.Sprintf("  ; Hardware Bare-Metal Action: %s\n", fnName))
			return "0"
		}
		if fnName == "mmio_read" && len(e.Arguments) >= 1 {
			addr := g.emitExpression(buf, e.Arguments[0])
			g.varCounter++
			ptr := fmt.Sprintf("%%ptr%d", g.varCounter)
			buf.WriteString(fmt.Sprintf("  %s = inttoptr i64 %s to i32*\n", ptr, addr))
			g.varCounter++
			tmp := fmt.Sprintf("%%t%d", g.varCounter)
			buf.WriteString(fmt.Sprintf("  %s = load volatile i32, i32* %s, align 4\n", tmp, ptr))
			return tmp
		}
		if fnName == "enable_interrupts" || fnName == "mcu.enable_irq" {
			buf.WriteString("  call void asm sideeffect \"cpsie i\", \"~{memory}\"() nounwind\n")
			return "0"
		}
		if fnName == "disable_interrupts" || fnName == "mcu.disable_irq" {
			buf.WriteString("  call void asm sideeffect \"cpsid i\", \"~{memory}\"() nounwind\n")
			return "0"
		}
		if fnName == "gui.draw_pixel" || fnName == "draw_pixel" {
			var args []string
			for _, arg := range e.Arguments {
				argVal := g.emitExpression(buf, arg)
				args = append(args, argVal)
			}
			if len(args) >= 5 {
				buf.WriteString(fmt.Sprintf("  call void @baremetal_fb_draw_pixel(i64 %s, i64 %s, i64 %s, i64 %s, i32 %s)\n",
					args[0], args[1], args[2], args[3], args[4]))
			}
			return "0"
		}
		if fnName == "gui.vga_char" || fnName == "vga_char" {
			var args []string
			for _, arg := range e.Arguments {
				argVal := g.emitExpression(buf, arg)
				args = append(args, argVal)
			}
			if len(args) >= 4 {
				buf.WriteString(fmt.Sprintf("  call void @baremetal_vga_write_char(i64 %s, i64 %s, i8 %s, i8 %s)\n",
					args[0], args[1], args[2], args[3]))
			}
			return "0"
		}
		var args []string
		for _, arg := range e.Arguments {
			argVal := g.emitExpression(buf, arg)
			args = append(args, fmt.Sprintf("i64 %s", argVal))
		}
		g.varCounter++
		tmp := fmt.Sprintf("%%t%d", g.varCounter)
		buf.WriteString(fmt.Sprintf("  %s = call i64 @%s(%s)\n", tmp, fnName, strings.Join(args, ", ")))
		return tmp
	default:
		return "0"
	}
}

func (g *LLVMGenerator) mapTypeToLLVM(t string) string {
	switch t {
	case "int", "int64", "uint64":
		return "i64"
	case "int32", "uint32":
		return "i32"
	case "int8", "byte", "bool":
		return "i8"
	case "void", "":
		return "void"
	default:
		return "i64"
	}
}
