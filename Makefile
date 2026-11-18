# Makefile for 100% Pure Freestanding AxiomLang Toolchain (Zero Go Runtime)

.PHONY: all build-asm build-axpm build-lsp build-browser build-stage1 test clean run-examples

BIN_DIR = bin
ASM_COMPILER = $(BIN_DIR)/axlc_asm
AXPM = $(BIN_DIR)/axpm
LSP_SERVER = $(BIN_DIR)/axiom_lsp
INSTALL_PREFIX ?= /usr/local/bin

all: build-asm build-axpm build-lsp build-stage1 test

build-asm:
	@echo "==> Building Pure Freestanding x86_64 Assembly Compiler $(ASM_COMPILER)..."
	@mkdir -p $(BIN_DIR) build/
	as -g -I. -o build/main_asm.o src_asm/main.s
	as -g -I. -o build/compiler_asm.o src_asm/compiler.s
	as -g -I. -o build/mca_runtime.o src_asm/mca_runtime.s
	as -g -I. -o build/hpt_optimizer.o src_asm/hpt_opt/hpt_optimizer.s
	as -g -I. -o build/lean4_emitter.o src_asm/lean4/lean4_emitter.s
	as -g -I. -o build/runtime_heal.o src_asm/runtime_heal/runtime_heal.s
	as -g -I. -o build/actor_sheaf.o src_asm/actor_sheaf/actor_sheaf.s
	as -g -I. -o build/dwarf.o src_asm/debug/dwarf.s
	as -g -I. -o build/lexer_core.o src_asm/core/lexer.s
	as -g -I. -o build/parser_ast.o src_asm/parser/ast.s
	as -g -I. -o build/checker_asm.o src_asm/checker/checker.s
	as -g -I. -o build/stdlib_actor_io.o src_asm/stdlib/actor_io.s
	ld -o $(ASM_COMPILER) build/main_asm.o build/compiler_asm.o build/mca_runtime.o build/hpt_optimizer.o build/lean4_emitter.o build/runtime_heal.o build/actor_sheaf.o build/dwarf.o build/lexer_core.o build/parser_ast.o build/checker_asm.o build/stdlib_actor_io.o
	@cp -f $(ASM_COMPILER) $(BIN_DIR)/axlc
	@echo "==> $(ASM_COMPILER) and $(BIN_DIR)/axlc built successfully (Pure Machine Code)."

build-axpm:
	@echo "==> Building Pure ASM Package Manager $(AXPM)..."
	@mkdir -p $(BIN_DIR) build/
	as -g -o build/axpm.o src_asm/axpm/main.s
	ld -o $(AXPM) build/axpm.o
	@echo "==> $(AXPM) built successfully."

build-lsp:
	@echo "==> Building Pure ASM LSP Server $(LSP_SERVER)..."
	@mkdir -p $(BIN_DIR) build/
	as -g -o build/axiom_lsp_asm.o src_asm/lsp/main.s
	ld -o $(LSP_SERVER) build/axiom_lsp_asm.o
	@echo "==> $(LSP_SERVER) built successfully."

build-stage1: build-asm
	@echo "==> Bootstrapping Stage 1 AxiomLang Frontend $(BIN_DIR)/axlc_stage1..."
	./$(ASM_COMPILER) src_axl/main.axl -o $(BIN_DIR)/axlc_stage1 > /dev/null
	@echo "==> $(BIN_DIR)/axlc_stage1 built successfully."

build-browser:
	@echo "==> Building Pure ASM AxiomBrowser $(BIN_DIR)/axiom_browser..."
	@mkdir -p $(BIN_DIR) build/
	as -g -I. -o build/browser_main.o src_asm/browser_main.s
	as -g -I. -o build/x11_engine.o src_asm/gui_x11/x11_engine.s
	as -g -I. -o build/net_client.o src_asm/core/net_client.s
	ld -o $(BIN_DIR)/axiom_browser build/browser_main.o build/x11_engine.o build/net_client.o
	@echo "==> $(BIN_DIR)/axiom_browser built successfully."

install: build-asm build-axpm build-lsp
	@echo "==> Installing binaries to $(INSTALL_PREFIX)..."
	sudo cp $(ASM_COMPILER) $(INSTALL_PREFIX)/axlc_asm
	sudo cp $(ASM_COMPILER) $(INSTALL_PREFIX)/axlc
	sudo cp $(AXPM) $(INSTALL_PREFIX)/axpm
	sudo cp $(LSP_SERVER) $(INSTALL_PREFIX)/axiom_lsp
	sudo chmod +x $(INSTALL_PREFIX)/axlc $(INSTALL_PREFIX)/axlc_asm $(INSTALL_PREFIX)/axpm $(INSTALL_PREFIX)/axiom_lsp

test: build-asm build-axpm build-lsp
	@echo "==> Running Pure Assembly Verification on Test Suite..."
	@for t in tests/pure_asm/*.axl; do \
		echo "--> Testing $$t..."; \
		./$(ASM_COMPILER) $$t -run || exit 1; \
	done
	@echo "--> Testing arithmetic, -o flag, and negative checks..."
	@printf 'fn main() {\n    let x = 21 * 2\n    println(x)\n}\n' > /tmp/axl_arith_check.axl
	@./$(ASM_COMPILER) /tmp/axl_arith_check.axl -o /tmp/axl_arith_bin > /dev/null
	@test "$$(/tmp/axl_arith_bin)" = "42" || (echo "Arithmetic check failed"; exit 1)
	@rm -f /tmp/axl_arith_check.axl /tmp/axl_arith_bin
	@! ./$(ASM_COMPILER) /tmp/nonexistent_axiom_file.axl 2>/dev/null
	@printf 'fn main() {\n    let bad = 10 / 0\n}\n' > /tmp/axl_div0.axl
	@! ./$(ASM_COMPILER) /tmp/axl_div0.axl 2>/dev/null
	@printf 'fn main() {\n    println(undeclared_var)\n}\n' > /tmp/axl_undef.axl
	@! ./$(ASM_COMPILER) /tmp/axl_undef.axl 2>/dev/null
	@printf 'require []\nfn main() {\n    println("denied")\n}\n' > /tmp/axl_cap.axl
	@! ./$(ASM_COMPILER) /tmp/axl_cap.axl 2>/dev/null
	@printf 'fn main() {\n    let s = "abc"\n    let x = s + 5\n}\n' > /tmp/axl_type.axl
	@! ./$(ASM_COMPILER) /tmp/axl_type.axl 2>/dev/null
	@printf 'require [IO]\nfn main() {\n    let a = [1, 2, 3]\n}\n' > /tmp/axl_memcap.axl
	@! ./$(ASM_COMPILER) /tmp/axl_memcap.axl 2>/dev/null
	@printf 'fn add(a, b) { return a + b }\nfn main() { println(add(1)) }\n' > /tmp/axl_arity.axl
	@! ./$(ASM_COMPILER) /tmp/axl_arity.axl 2>/dev/null
	@printf 'fn inc(x: int) -> int { return x + 1 }\nfn main() { println(inc("bad")) }\n' > /tmp/axl_argtype.axl
	@! ./$(ASM_COMPILER) /tmp/axl_argtype.axl 2>/dev/null
	@printf 'require [Memory, IO]\nstruct User { name: string, age: int }\nfn main() {\n    let u = { name: "Admin", age: 30 }\n    println(u.name)\n    println(u.age)\n    for s in ["A", "B"] { print(s) }\n}\n' > /tmp/axl_str_field.axl
	@./$(ASM_COMPILER) /tmp/axl_str_field.axl -o /tmp/axl_str_field_bin > /dev/null
	@test "$$(/tmp/axl_str_field_bin)" = "$$(printf 'Admin\n30\nAB')" || (echo "Struct/for-in string check failed"; exit 1)
	@printf 'require [Memory, IO]\nfn main() {\n    let a = [1, 2]\n    println(a[5])\n}\n' > /tmp/axl_oob.axl
	@./$(ASM_COMPILER) /tmp/axl_oob.axl -o /tmp/axl_oob_bin > /dev/null
	@! /tmp/axl_oob_bin 2>/tmp/axl_oob_err.txt
	@grep -q "MCA-FATAL" /tmp/axl_oob_err.txt || (echo "MCA-FATAL diagnostic missing"; exit 1)
	@printf 'require [Memory, IO]\nstruct Point { x: int, y: int }\nfn main() {\n    let p = { x: 1, y: 2 }\n    println(p.z)\n}\n' > /tmp/axl_badfld.axl
	@! ./$(ASM_COMPILER) /tmp/axl_badfld.axl 2>/dev/null
	@printf 'require [Memory, IO]\ncap memory [Scope: Arena, Max: 1KB]\nfn main() {\n    let f = read_file("Makefile")\n    println(f)\n}\n' > /tmp/axl_capmax.axl
	@./$(ASM_COMPILER) /tmp/axl_capmax.axl -o /tmp/axl_capmax_bin > /dev/null
	@! /tmp/axl_capmax_bin 2>/tmp/axl_capmax_err.txt
	@grep -q "MCA-FATAL" /tmp/axl_capmax_err.txt || (echo "cap Max:1KB enforcement missing"; exit 1)
	@rm -f /tmp/axl_div0.axl /tmp/axl_undef.axl /tmp/axl_cap.axl /tmp/axl_type.axl /tmp/axl_memcap.axl /tmp/axl_arity.axl /tmp/axl_argtype.axl /tmp/axl_str_field.axl /tmp/axl_str_field_bin /tmp/axl_oob.axl /tmp/axl_oob_bin /tmp/axl_oob_err.txt /tmp/axl_badfld.axl /tmp/axl_capmax.axl /tmp/axl_capmax_bin /tmp/axl_capmax_err.txt
	@echo "==> Running Beta Integration Suite..."
	@./$(ASM_COMPILER) tests/integration/01_functions_recursion.axl -o /tmp/axl_int1 > /dev/null
	@test "$$(/tmp/axl_int1)" = "$$(printf '55\n42')" || (echo "Integration 01 failed"; exit 1)
	@./$(ASM_COMPILER) tests/integration/02_control_flow.axl -o /tmp/axl_int2 > /dev/null
	@test "$$(/tmp/axl_int2)" = "$$(printf '15\n15\nFLOW_OK')" || (echo "Integration 02 failed"; exit 1)
	@./$(ASM_COMPILER) tests/integration/03_comparisons_logic.axl -o /tmp/axl_int3 > /dev/null
	@test "$$(/tmp/axl_int3)" = "$$(printf '1\n10\nLOGIC_OK')" || (echo "Integration 03 failed"; exit 1)
	@./$(ASM_COMPILER) tests/integration/04_arrays_and_opt.axl -o /tmp/axl_int4 > /dev/null
	@test "$$(/tmp/axl_int4)" = "$$(printf '4\n30\n100\n14')" || (echo "Integration 04 failed"; exit 1)
	@./$(ASM_COMPILER) tests/integration/05_structs_and_pipeline.axl -o /tmp/axl_int5 > /dev/null
	@test "$$(/tmp/axl_int5)" = "$$(printf '42\n17')" || (echo "Integration 05 failed"; exit 1)
	@./$(ASM_COMPILER) tests/integration/06_strings_for_mut.axl -o /tmp/axl_int6 > /dev/null
	@test "$$(/tmp/axl_int6)" = "$$(printf 'AxiomStr\n8\n40\n24')" || (echo "Integration 06 failed"; exit 1)
	@./$(ASM_COMPILER) tests/integration/07_str_ops_and_exit.axl -o /tmp/axl_int7 > /dev/null
	@test "$$(/tmp/axl_int7)" = "$$(printf 'AxiomLang\nSTR_EQ_OK')" || (echo "Integration 07 failed"; exit 1)
	@./$(ASM_COMPILER) tests/integration/08_cli_and_file_io.axl -o /tmp/axl_int8 > /dev/null
	@test "$$(/tmp/axl_int8 hello_cli)" = "$$(printf '2\nhello_cli\n17\nAxiomIO:hello_cli')" || (echo "Integration 08 failed"; exit 1)
	@./$(ASM_COMPILER) tests/integration/09_strings_slice_and_fn_ret.axl -o /tmp/axl_int9 > /dev/null
	@test "$$(/tmp/axl_int9)" = "$$(printf 'Hello, Axiom\nLang\n66\n58')" || (echo "Integration 09 failed"; exit 1)
	@./$(ASM_COMPILER) tests/integration/10_threads_and_channels.axl -o /tmp/axl_int10 > /dev/null
	@test "$$(/tmp/axl_int10)" = "$$(printf '100\n200\n101\n201\n100000\n84')" || (echo "Integration 10 failed"; exit 1)
	@./$(ASM_COMPILER) tests/integration/11_industrial_beta_stdlib.axl -o /tmp/axl_int11 > /dev/null
	@test "$$(/tmp/axl_int11)" = "$$(printf 'ARENA_OK\n0,-42,12345\n350\n0\n42\ndivision by zero\nECHO:PING_BETA\nTIME_OK')" || (echo "Integration 11 failed"; exit 1)
	@rm -f /tmp/axl_int1 /tmp/axl_int2 /tmp/axl_int3 /tmp/axl_int4 /tmp/axl_int5 /tmp/axl_int6 /tmp/axl_int7 /tmp/axl_int8 /tmp/axl_int9 /tmp/axl_int10 /tmp/axl_int11 /tmp/axl_io_test.txt
	@echo "==> Running AXPM Package Suite..."
	./$(AXPM) test
	@echo "==> Running Stage 1 Self-Hosting Bootstrap Check..."
	./$(ASM_COMPILER) src_axl/main.axl -o /tmp/axl_stage1_test > /dev/null
	@test "$$(/tmp/axl_stage1_test examples/01_hello_world.axl)" = "Stage1 Lexer [examples/01_hello_world.axl] tokens: 23" || (echo "Stage1 bootstrap check failed"; exit 1)
	@/tmp/axl_stage1_test examples/01_hello_world.axl -S /tmp/axl_s1.s > /dev/null
	@as /tmp/axl_s1.s -o /tmp/axl_s1.o && ld /tmp/axl_s1.o -o /tmp/axl_s1_bin
	@test "$$(/tmp/axl_s1_bin)" = "$$(./$(ASM_COMPILER) examples/01_hello_world.axl -o /tmp/axl_s0_bin > /dev/null && /tmp/axl_s0_bin)" || (echo "Stage1 codegen output mismatch"; exit 1)
	@echo "==> Running MCA Fiber Cooperative Concurrency Check..."
	@printf 'require [Memory, IO]\nfn worker_a() {\n    println("A1")\n    fiber_yield()\n    println("A2")\n}\nfn worker_b() {\n    println("B1")\n    fiber_yield()\n    println("B2")\n}\nfn main() {\n    spawn worker_a()\n    spawn worker_b()\n    fiber_run_all()\n}\n' > /tmp/axl_fiber.axl
	@./$(ASM_COMPILER) /tmp/axl_fiber.axl -o /tmp/axl_fiber_bin > /dev/null
	@test "$$(/tmp/axl_fiber_bin)" = "$$(printf 'A1\nB1\nA2\nB2')" || (echo "MCA Fiber test failed"; exit 1)
	@rm -f /tmp/axl_stage1_test /tmp/axl_s1.s /tmp/axl_s1.o /tmp/axl_s1_bin /tmp/axl_s0_bin /tmp/axl_fiber.axl /tmp/axl_fiber_bin
	@echo "==> [ALL PURE ASM TESTS PASSED SUCCESSFULLY]"

clean:
	@echo "==> Cleaning build artifacts..."
	rm -rf $(BIN_DIR)/* build/

run-examples: build-asm
	@echo "==> Executing example suite..."
	./$(ASM_COMPILER) examples/01_hello_world.axl -run
	./$(ASM_COMPILER) examples/02_gui_app.axl -run
	./$(ASM_COMPILER) examples/03_actor_concurrency.axl -run
	./$(ASM_COMPILER) examples/04_data_pipeline.axl -run
	./$(ASM_COMPILER) examples/05_for_in_loop.axl -run
	./$(ASM_COMPILER) examples/06_functions.axl -run
	./$(ASM_COMPILER) examples/07_structs.axl -run
	./$(ASM_COMPILER) examples/09_axlui_gui_studio.axl -run
	./$(ASM_COMPILER) examples/17_range_loop_demo.axl -run
	./$(ASM_COMPILER) examples/18_autocap_toplevel.axl -run
	./$(ASM_COMPILER) examples/19_pipeline_demo.axl -run
