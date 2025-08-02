.PHONY: build install clean test version

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
BUILD_DATE := $(shell date -u '+%Y-%m-%d %H:%M:%S UTC')

LDFLAGS := -X 'github.com/karurosux/saas-go-kit/cmd/sgk/commands.Version=$(VERSION)' \
           -X 'github.com/karurosux/saas-go-kit/cmd/sgk/commands.Commit=$(COMMIT)' \
           -X 'github.com/karurosux/saas-go-kit/cmd/sgk/commands.BuildDate=$(BUILD_DATE)'

# Default target
all: build

# Build the CLI binary
build:
	@echo "🔨 Building sgk CLI..."
	@echo "   Version: $(VERSION)"
	@echo "   Commit:  $(COMMIT)"
	@go build -ldflags "$(LDFLAGS)" -o sgk

# Install the CLI to GOPATH/bin
install:
	@echo "📦 Installing sgk..."
	@echo "   Version: $(VERSION)"
	@echo "   Commit:  $(COMMIT)"
	@go install -ldflags "$(LDFLAGS)"

# Clean build artifacts
clean:
	@echo "🧹 Cleaning..."
	@rm -f sgk saas-kit saas-kit-embedded saas-kit-test

# Run tests
test:
	@echo "🧪 Running tests..."
	@go test ./...

# Display version information
version:
	@echo "SGK Version Information:"
	@echo "  Version:    $(VERSION)"
	@echo "  Commit:     $(COMMIT)"
	@echo "  Build Date: $(BUILD_DATE)"

.DEFAULT_GOAL := all