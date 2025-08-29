# Makefile for Slop Shop

.PHONY: build run clean test examples help

# Build the program
build:
	go build -o slop-shop

# Run with a basic prompt (requires -prompt argument)
run:
	@if [ -z "$(prompt)" ]; then \
		echo "Usage: make run prompt='Your prompt here'"; \
		echo "Example: make run prompt='Analyze this codebase'"; \
		exit 1; \
	fi
	./slop-shop -prompt "$(prompt)"

# Start interactive REPL mode
repl:
	./slop-shop -repl

# Apply changes using tools mode with custom prompt
diff:
	@if [ -z "$(prompt)" ]; then \
		echo "Usage: make diff prompt='Your prompt here'"; \
		echo "Example: make diff prompt='Add error handling to main function'"; \
		exit 1; \
	fi
	./slop-shop -tools -prompt "$(prompt)"

# Run with tools enabled
tools:
	@if [ -z "$(prompt)" ]; then \
		echo "Usage: make tools prompt='Your prompt here'"; \
		echo "Example: make tools prompt='Check directory and test Go'"; \
		exit 1; \
	fi
	./slop-shop -tools -prompt "$(prompt)"

# Run with default prompt for testing
test:
	./slop-shop -prompt "What is this repository about?"

# Clean build artifacts
clean:
	rm -f slop-shop
	go clean

# Show examples
examples:
	./examples.sh

# Install dependencies (if needed)
deps:
	go mod tidy

# Show help
help:
	@echo "Available targets:"
	@echo "  build     - Build the program"
	@echo "  run       - Run with custom prompt (use: make run prompt='Your prompt')"
	@echo "  repl      - Start interactive REPL mode"
	@echo "  diff      - Apply changes using tools (use: make diff prompt='Your prompt')"
	@echo "  tools     - Run with tools enabled (use: make tools prompt='Your prompt')"
	@echo "  test      - Run with test prompt"
	@echo "  clean     - Remove build artifacts"
	@echo "  examples  - Show usage examples"
	@echo "  deps      - Install dependencies"
	@echo "  help      - Show this help"
	@echo ""
	@echo "Examples:"
	@echo "  make build"
	@echo "  make run prompt='Analyze this code for security issues'"
	@echo "  make repl"
	@echo "  make diff prompt='Add error handling to main function'"
	@echo "  make tools prompt='Check directory and test Go'"
	@echo "  make test"
	@echo "  make examples"
