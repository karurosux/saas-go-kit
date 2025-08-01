.PHONY: build install clean test

# Default target
all: build

# Build the CLI binary
build:
	@echo "🔨 Building sgk CLI..."
	@go build -o sgk

# Install the CLI to GOPATH/bin
install:
	@echo "📦 Installing sgk..."
	@go install

# Clean build artifacts
clean:
	@echo "🧹 Cleaning..."
	@rm -f sgk saas-kit saas-kit-embedded saas-kit-test

# Run tests
test:
	@echo "🧪 Running tests..."
	@go test ./...

.DEFAULT_GOAL := all