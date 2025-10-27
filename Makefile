# Makefile for go-sockets development

.PHONY: all build test clean fmt vet lint help

# Default target
all: build

# Build all packages
build:
	go build ./...

# Run all tests
test:
	go test ./...

# Run tests with coverage
test-cover:
	go test -cover ./...

# Run tests with race detection
test-race:
	go test -race ./...

# Clean build artifacts
clean:
	go clean ./...
	go mod tidy

# Format code
fmt:
	go fmt ./...

# Vet code
vet:
	go vet ./...

# Lint code (requires golangci-lint)
lint:
	golangci-lint run

# Run benchmarks
bench:
	go test -bench=. ./...

# Help
help:
	@echo "Available targets:"
	@echo "  all        - Build the project (default)"
	@echo "  build      - Build all packages"
	@echo "  test       - Run all tests"
	@echo "  test-cover - Run tests with coverage"
	@echo "  test-race  - Run tests with race detection"
	@echo "  clean      - Clean build artifacts and tidy modules"
	@echo "  fmt        - Format code"
	@echo "  vet        - Vet code"
	@echo "  lint       - Lint code (requires golangci-lint)"
	@echo "  bench      - Run benchmarks"
	@echo "  help       - Show this help"