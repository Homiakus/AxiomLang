# Makefile for AxiomLang (axlc) Toolchain

.PHONY: all build build-asm install test clean run-examples

BIN_DIR = bin
BIN_NAME = $(BIN_DIR)/axlc
ASM_COMPILER = $(BIN_DIR)/axlc_asm
SRC = cmd/axlc/main.go
INSTALL_PREFIX ?= /usr/local/bin

all: build build-asm test

build:
	@echo "==> Building $(BIN_NAME)..."
	@mkdir -p $(BIN_DIR)
	go build -ldflags="-s -w" -o $(BIN_NAME) $(SRC)

build-asm:
	@echo "==> Building Pure Freestanding x86_64 Assembly Compiler $(ASM_COMPILER)..."
	@mkdir -p $(BIN_DIR) build/
	as -o build/main_asm.o src_asm/main.s
	as -o build/compiler_asm.o src_asm/compiler.s
	ld -o $(ASM_COMPILER) build/main_asm.o build/compiler_asm.o
	@echo "==> $(ASM_COMPILER) built successfully."

install: build build-asm
	@echo "==> Installing $(BIN_NAME) and $(ASM_COMPILER) to $(INSTALL_PREFIX)..."
	sudo cp $(BIN_NAME) $(INSTALL_PREFIX)/axlc
	sudo chmod +x $(INSTALL_PREFIX)/axlc
	sudo cp $(ASM_COMPILER) $(INSTALL_PREFIX)/axlc_asm
	sudo chmod +x $(INSTALL_PREFIX)/axlc_asm

test:
	@echo "==> Running all AxiomLang unit test suites..."
	go test -v ./pkg/... ./stdlib/...
	@echo "==> Running Pure Assembly Compiler Verification..."
	./$(ASM_COMPILER) examples/01_hello_world.axl -run

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
