# Makefile for Pure Freestanding AxiomLang Toolchain (Zero Go Runtime)

.PHONY: all build-asm build-axpm test clean run-examples

BIN_DIR = bin
ASM_COMPILER = $(BIN_DIR)/axlc_asm
AXPM = $(BIN_DIR)/axpm
INSTALL_PREFIX ?= /usr/local/bin

all: build-asm build-axpm test

build-asm:
	@echo "==> Building Pure Freestanding x86_64 Assembly Compiler $(ASM_COMPILER)..."
	@mkdir -p $(BIN_DIR) build/
	as -g -o build/main_asm.o src_asm/main.s
	as -g -o build/compiler_asm.o src_asm/compiler.s
	as -g -o build/mca_runtime.o src_asm/mca_runtime.s
	as -g -o build/dwarf.o src_asm/debug/dwarf.s
	ld -o $(ASM_COMPILER) build/main_asm.o build/compiler_asm.o build/mca_runtime.o build/dwarf.o
	@echo "==> $(ASM_COMPILER) built successfully (Pure Machine Code)."

build-axpm:
	@echo "==> Building Pure ASM Package Manager $(AXPM)..."
	@mkdir -p $(BIN_DIR) build/
	as -g -o build/axpm.o src_asm/axpm/main.s
	ld -o $(AXPM) build/axpm.o
	@echo "==> $(AXPM) built successfully."

install: build-asm build-axpm
	@echo "==> Installing $(ASM_COMPILER) and $(AXPM) to $(INSTALL_PREFIX)..."
	sudo cp $(ASM_COMPILER) $(INSTALL_PREFIX)/axlc_asm
	sudo cp $(ASM_COMPILER) $(INSTALL_PREFIX)/axlc
	sudo cp $(AXPM) $(INSTALL_PREFIX)/axpm
	sudo chmod +x $(INSTALL_PREFIX)/axlc $(INSTALL_PREFIX)/axlc_asm $(INSTALL_PREFIX)/axpm

test: build-asm build-axpm
	@echo "==> Running Pure Assembly Verification on Test Suite..."
	@for t in tests/pure_asm/*.axl; do \
		echo "--> Testing $$t..."; \
		./$(ASM_COMPILER) $$t -run || exit 1; \
	done
	@echo "==> Running AXPM Package Suite..."
	./$(AXPM) test
	@echo "==> [ALL TESTS PASSED SUCCESSFULLY]"

clean:
	@echo "==> Cleaning build artifacts..."
	rm -rf $(BIN_DIR)/* build/

run-examples: build
	@echo "==> Executing example suite..."
	./$(BIN_NAME) -run examples/01_hello_world.axl
	./$(BIN_NAME) -run examples/04_data_pipeline.axl
	./$(BIN_NAME) -run examples/05_for_in_loop.axl
	./$(BIN_NAME) -run examples/06_functions.axl
	./$(BIN_NAME) -run examples/07_structs.axl
	./$(BIN_NAME) -run examples/08_actor_erlang_syntax.axl
