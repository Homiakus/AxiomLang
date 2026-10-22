# AxiomLang

**AxiomLang** is a strictly typed systems programming language engineered for verifiable memory safety and bounded, deterministic execution time.

The compiler toolchain (`axlc_asm`) and package manager (`axpm`) are distributed as lightweight standalone binaries without external dynamic library dependencies, communicating directly with the Linux kernel via raw system calls.

---

## Core Capabilities

- **Predictable Performance**: Statically compiled x86_64 machine code featuring sub-microsecond cold starts and binary footprints starting at a few kilobytes.
- **Mandate Capability Arena (MCA) Memory Model**:
  - Direct region allocation via `sys_mmap` with $O(1)$ 8-byte aligned bump allocation.
  - Scope-level access control enforced through bitmask capabilities (`CAP_MEM_READ`, `CAP_MEM_WRITE`, `CAP_MEM_GROW`).
  - $O(1)$ scope rewinds and full arena deallocation upon scope exit, completely eliminating memory fragmentation.
- **Integrated Package Manager (`axpm`)**: Workspace initialization, local dependency resolution, project builds, and test runs driven by a single unified command.
- **Standard DWARF Debugging (GDB / LLDB)**: Automatic generation of `.debug_line` tables providing exact mapping from emitted machine instructions back to `.axl` source lines.
- **Static Invariant Verification**: Built-in compile-time bounds checking and division-by-zero validation.

---

## Prerequisites and Building

Building the toolchain from source requires only standard `as` and `ld` utilities (from GNU `binutils`):

```bash
make
```

The build process produces two binaries in `bin/`:
- `bin/axlc_asm` — the core `.axl` compiler;
- `bin/axpm` — the project and package manager.

---

## Quickstart Guide

### 1. Initialize a Project
Create a new project workspace:
```bash
./bin/axpm init my_service
```

This generates the `axiom.pkg` manifest and a starter entry point at `src/main.axl`:
```axiom
require [Memory, IO]

intent "Production microservice"
cap memory [Scope: Arena, Max: 64KB]

fn main() {
    println("[✓] Service running under MCA memory model.")
}
```

### 2. Build and Run
```bash
# Compile the project
./bin/axpm build

# Execute the resulting binary
./bin/axpm run
```

### 3. Run Invariant Tests
```bash
./bin/axpm test
```

---

## Direct Usage of `axlc_asm`

### Fast In-Memory Execution
Compile and execute an Axiom script directly:
```bash
./bin/axlc_asm examples/01_hello_world.axl -run
```

### Compile to a Standalone Binary
Produce an optimized static binary:
```bash
./bin/axlc_asm examples/01_hello_world.axl -o app
./app
```

### Interactive Debugging with GDB
DWARF debug records allow stepping through lines directly in GDB:
```bash
gdb ./app
(gdb) break _start
(gdb) run
(gdb) stepi
```

---

## Verification Test Suite

A complete verification test suite is maintained under `tests/pure_asm/`:

| Test File | Verification Target |
| :--- | :--- |
| [`01_hello_world.axl`](tests/pure_asm/01_hello_world.axl) | Streaming I/O (`sys_write`) and exit code verification (`sys_exit`). |
| [`02_mca_arena.axl`](tests/pure_asm/02_mca_arena.axl) | MCA page mapping, boundary enforcement, and $O(1)$ scope reclamation. |
| [`03_arithmetic.axl`](tests/pure_asm/03_arithmetic.axl) | Integer arithmetic, bitwise operations, and constant folding. |
| [`04_actor_mandate.axl`](tests/pure_asm/04_actor_mandate.axl) | Actor context isolation and capability-checked message passing. |
| [`05_dwarf_debug.axl`](tests/pure_asm/05_dwarf_debug.axl) | DWARF symbol generation and debugger breakpoint alignment. |

Run the full suite:
```bash
make test
```

---

## Project Structure

```text
AxiomLang/
├── bin/                 # Compiled binaries (axlc_asm, axpm)
├── src_asm/             # Freestanding assembly toolchain sources
│   ├── main.s           # Compiler entry point and CLI option dispatcher
│   ├── compiler.s       # Lexer, parser, and machine instruction emitter
│   ├── mca_runtime.s    # Mandate Capability Arena runtime (sys_mmap)
│   ├── debug/dwarf.s    # DWARF line number emitter (.debug_line)
│   ├── axpm/            # Native package manager sources
│   └── include/         # System call numbers, tokens, and MCA structures
├── tests/pure_asm/      # Formal verification test specifications
├── examples/            # Reference programs, graphics, and algorithms
├── editors/vscode/      # Visual Studio Code syntax highlighting
├── Makefile             # Build automation
├── README.md            # Documentation (Russian)
└── README_EN.md         # Documentation (English)
```

---

## License

MIT License. Open for use in commercial, embedded, and research applications.
