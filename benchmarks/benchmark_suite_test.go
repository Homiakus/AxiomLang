package benchmarks

import (
	"context"
	"strings"
	"sync/atomic"
	"testing"

	"AxiomLang/pkg/ai"
	"AxiomLang/pkg/checker"
	"AxiomLang/pkg/codegen"
	"AxiomLang/pkg/lexer"
	"AxiomLang/pkg/llvm"
	"AxiomLang/pkg/parser"
	"AxiomLang/stdlib/actor"
	"AxiomLang/stdlib/data"
	"AxiomLang/stdlib/gui"
	"AxiomLang/stdlib/nn"
)

// ----------------------------------------------------------------------------
// 1. FRONTEND / COMPILER BENCHMARKS (Lexer, Parser, AI Healing, Checker, Codegen)
// ----------------------------------------------------------------------------

const sampleSmallSource = `
требовать Capability[StdIO]

намерение Compute {
    пусть изм sum = 0
    для i в [1, 2, 3, 4, 5, 6, 7, 8, 9, 10] {
        если i > 5 {
            sum = sum + i * 2
        } иначе {
            sum = sum + 1
        }
    }
    вернуть sum
}
`

const sampleFaultySource = `
требует capa[stdout]
намерение AutoFixMe {
    пусть val = 100
    если val != 0 {
        вернуть 42
    }
}
`

// Benchmark 1.1: SIMD Lexer
func BenchmarkCompiler_Lexer_Small(b *testing.B) {
	b.SetBytes(int64(len(sampleSmallSource)))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l := lexer.New(sampleSmallSource)
		for {
			tok := l.NextToken()
			if tok.Type == lexer.EOF {
				break
			}
		}
	}
}

// Benchmark 1.2: Parser AST generation
func BenchmarkCompiler_Parser_Small(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l := lexer.New(sampleSmallSource)
		p := parser.New(l)
		_ = p.ParseProgram()
	}
}

// Benchmark 1.3: Micro-Neural AI Auto-Healing Pass
func BenchmarkCompiler_AIHealer(b *testing.B) {
	healer := ai.NewMicroNeuralHealer()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = healer.HealSource(sampleFaultySource)
	}
}

// Benchmark 1.4: Strict Zero-Ambient Capability Scope Checker
func BenchmarkCompiler_CapabilityChecker(b *testing.B) {
	l := lexer.New(sampleSmallSource)
	p := parser.New(l)
	prog := p.ParseProgram()
	checkerInstance := checker.New()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = checkerInstance.Check(prog)
	}
}

// Benchmark 1.5: Pure x86_64 ASM Generation
func BenchmarkCompiler_Codegen_PureASM(b *testing.B) {
	l := lexer.New(sampleSmallSource)
	p := parser.New(l)
	prog := p.ParseProgram()
	gen := codegen.NewAsmGenerator(false, false)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = gen.GenerateX86_64Asm(prog)
	}
}

// Benchmark 1.6: LLVM IR Generation
func BenchmarkCompiler_Codegen_LLVM(b *testing.B) {
	l := lexer.New(sampleSmallSource)
	p := parser.New(l)
	prog := p.ParseProgram()
	llvmGen := llvm.NewLLVMGenerator(llvm.TargetX86_64BareMetal)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = llvmGen.GenerateLLVMIR(prog)
	}
}

// ----------------------------------------------------------------------------
// 2. RUNTIME COMPUTE & DATA BENCHMARKS (SIMD vs Scalar, NN, Actors)
// ----------------------------------------------------------------------------

// Benchmark 2.1: AVX2 Data Vector Reduction (1 Million Int64s)
func BenchmarkRuntime_Data_AVX2Sum_1M(b *testing.B) {
	n := 1_000_000
	df := data.NewDataFrame()
	vals := make([]int64, n)
	for i := 0; i < n; i++ {
		vals[i] = int64(i)
	}
	df.AddInt64Column("vals", vals)
	b.SetBytes(int64(n * 8))
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = df.SumInt64("vals")
	}
}

// Benchmark 2.2: Scalar Data Sum (1 Million Int64s - Baseline)
func BenchmarkRuntime_Data_ScalarSum_1M(b *testing.B) {
	n := 1_000_000
	vals := make([]int64, n)
	for i := 0; i < n; i++ {
		vals[i] = int64(i)
	}
	b.SetBytes(int64(n * 8))
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var s int64
		for _, v := range vals {
			s += v
		}
		_ = s
	}
}

// Benchmark 2.3: FMA3 AVX Neural Network Forward Pass (Layer 128 -> 64)
func BenchmarkRuntime_NN_ForwardPass_128x64(b *testing.B) {
	in := 128
	out := 64
	layer := nn.NewLinearLayer(in, out)
	input := nn.NewTensor(in)
	for i := 0; i < in; i++ {
		input.Data[i] = float32(i) * 0.01
	}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = layer.Forward(input)
	}
}

// Benchmark 2.4: Actor Message Passing (Concurrency & Mailbox)
func BenchmarkRuntime_Actor_MessagePassing(b *testing.B) {
	system := actor.NewSystem()
	var counter int64
	_, _ = system.Spawn("worker", 10000, func(ctx context.Context, msg actor.Message) error {
		atomic.AddInt64(&counter, 1)
		return nil
	})
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = system.Send("worker", actor.Message{Type: "TICK"})
	}
}

// ----------------------------------------------------------------------------
// 3. GRAPHICS & FRAMEBUFFER BENCHMARKS (1920x1080 FullHD)
// ----------------------------------------------------------------------------

// Benchmark 3.1: SIMD AVX2 Canvas Clear (1080p, 8MB buffer)
func BenchmarkRuntime_GUI_AVX2Clear_1080p(b *testing.B) {
	canvas := gui.NewNativeCanvas(1920, 1080)
	color := gui.Color32(0xFF1E1E2E)
	b.SetBytes(int64(1920 * 1080 * 4))
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		canvas.Clear(color)
	}
}

// Benchmark 3.2: SIMD Span Alpha Blend (Row 1920px)
func BenchmarkRuntime_GUI_AlphaBlend_1920pxRow(b *testing.B) {
	canvas := gui.NewNativeCanvas(1920, 1080)
	color := gui.Color32(0x80FF00FF)
	b.SetBytes(int64(1920 * 4))
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		canvas.DrawRect(0, 500, 1920, 1, color)
	}
}

// Benchmark 3.3: Native Scene Graph Layout & Render (Glassmorphic Window)
func BenchmarkRuntime_GUI_FullWindowRender(b *testing.B) {
	model := gui.NewAXLUIModel("bench.axlui", "BenchmarkUI")
	model.Theme = "cyber-dark"
	canvas := gui.NewNativeCanvas(1920, 1080)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		gui.RenderSceneToCanvas(canvas, model)
	}
}

// ----------------------------------------------------------------------------
// 4. MEMORY / ALLOCATION STRESS TEST
// ----------------------------------------------------------------------------

// Benchmark 4.1: High churn AST construction (Reveals GC pressure vs BareMetal)
func BenchmarkStress_ASTAllocations(b *testing.B) {
	longCode := strings.Repeat(sampleSmallSource, 10)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		l := lexer.New(longCode)
		p := parser.New(l)
		_ = p.ParseProgram()
	}
}
