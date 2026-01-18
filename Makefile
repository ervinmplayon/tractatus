.PHONY: build run clean test help

# Build the binary
build:
	@echo "Building tractatus..."
	@go build -o tractatus ./cmd/main.go
	@echo "✓ Build complete: ./tractatus"

# Run with default settings (requires config.json)
run:
	@go run ./cmd/main.go --account production

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -f tractatus
	@echo "✓ Clean complete"

# Run tests (when you add them)
test:
	@go test -v ./...

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy
	@echo "✓ Dependencies ready"

# Format code
fmt:
	@echo "Formatting code..."
	@go fmt ./...
	@echo "✓ Format complete"

# Lint code (requires golangci-lint)
lint:
	@echo "Linting code..."
	@golangci-lint run
	@echo "✓ Lint complete"

# Install the binary to $GOPATH/bin
install:
	@echo "Installing tractatus..."
	@go install ./cmd/main.go
	@echo "✓ Installed to $(go env GOPATH)/bin/tractatus"

# Show help
help:
	@echo "Available targets:"
	@echo "  build    - Build the binary"
	@echo "  run      - Run with default settings"
	@echo "  clean    - Remove build artifacts"
	@echo "  test     - Run tests"
	@echo "  deps     - Download dependencies"
	@echo "  fmt      - Format code"
	@echo "  lint     - Lint code (requires golangci-lint)"
	@echo "  install  - Install to GOPATH/bin"
	@echo "  help     - Show this help message"