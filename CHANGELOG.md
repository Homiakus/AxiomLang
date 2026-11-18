# AxiomLang Release Notes

## v1.0.0 (Linux x86_64 Strict Dialect)

### Core Compiler (`bin/axlc_asm` / `bin/axlc`)
- Pure freestanding `x86_64` assembly toolchain (zero libc, zero Go runtime).
- Dynamic `SYS_MMAP` source and output buffers (`SYS_FSTAT`-sized input, 16 MB output).
- Recursive `import "path/to/file.axl"` module expansion with duplicate/cycle protection.
- Static checker: constant div-by-zero, undeclared variables/functions, function arity & parameter type checking (`int`, `str`), struct field validation, `IO` / `Memory` capability enforcement, lexical shadowing, 4096-entry symbol tables.
- DWARF `.file` / `.loc` line debug emission and `hpt_optimizer` peephole `pushq %rax; popq %rax` elimination.

### MCA Memory Model & Runtime
- `SYS_MMAP`-backed MCA Arena in generated binaries with 8-byte alignment, `CAP_MEM_WRITE` enforcement (`mca_runtime.s`), `axiom_rt_save` / `axiom_rt_restore` / `axiom_rt_reset` primitives.
- Runtime enforcement of `cap memory [Scope: Arena, Max: <N>KB|MB]` and bounds-checked arrays with `[MCA-FATAL]` stderr diagnostics.

### Stage 1 Self-Hosting Bootstrap (`bin/axlc_stage1`, `src_axl/`)
- Multi-file AxiomLang lexer (`src_axl/lexer.axl`), x86_64 assembly codegen (`src_axl/codegen.axl`), and CLI (`src_axl/main.axl`) compiled by Stage 0 `axlc_asm` and verified for output parity.

### Tooling
- `bin/axpm`: project scaffolding (`init`), manifest `entry` resolution (`build`, `run`, `test`).
- `bin/axiom_lsp`: JSON-RPC server with dynamic request `"id"` extraction and `Content-Length` framing.
