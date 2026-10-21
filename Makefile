# Makefile for AxiomLang (axlc) Toolchain

.PHONY: all build install test clean run-examples

BIN_DIR = bin
BIN_NAME = $(BIN_DIR)/axlc
SRC = cmd/axlc/main.go
INSTALL_PREFIX ?= /usr/local/bin

all: build test

build:
	@echo "==> Building $(BIN_NAME)..."
	@mkdir -p $(BIN_DIR)
	go build -ldflags="-s -w" -o $(BIN_NAME) $(SRC)

install: build
	@echo "==> Installing $(BIN_NAME) to $(INSTALL_PREFIX)..."
	sudo cp $(BIN_NAME) $(INSTALL_PREFIX)/axlc
	sudo chmod +x $(INSTALL_PREFIX)/axlc

test:
	@echo "==> Running all AxiomLang unit test suites..."
	go test -v ./pkg/... ./stdlib/...

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
