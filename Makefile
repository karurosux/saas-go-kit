.PHONY: help generate-routes generate-clients clean-generated

# Default target
help:
	@echo "SaaS Go Kit - Client Generation Tools"
	@echo ""
	@echo "Available targets:"
	@echo "  generate-routes     Generate routes metadata from modules"
	@echo "  generate-clients    Generate TypeScript clients from routes"
	@echo "  generate-all        Generate both routes and clients"
	@echo "  clean-generated     Clean generated files"
	@echo "  help               Show this help message"

# Generate routes metadata
generate-routes:
	@echo "Generating routes metadata..."
	@mkdir -p ./generated
	@go run ./cmd/generate-routes -config ./saas-kit-modules.json -output ./generated/routes.json

# Generate TypeScript clients
generate-clients: generate-routes
	@echo "Generating TypeScript clients..."
	@go run ./cmd/generate-clients -routes ./generated/routes.json -o ./generated/clients

# Generate both routes and clients
generate-all: generate-clients
	@echo "✅ Client generation complete!"
	@echo "Routes metadata: ./generated/routes.json"
	@echo "TypeScript clients: ./generated/clients/"

# Clean generated files
clean-generated:
	@echo "Cleaning generated files..."
	@rm -rf ./generated

# Install required dependencies for the tools
install-deps:
	@echo "Installing dependencies for client generation tools..."
	@cd cmd/generate-routes && go mod tidy
	@cd cmd/generate-clients && go mod tidy

# Example: Generate clients for a specific project
# This would be used in projects that use saas-go-kit
example-generate:
	@echo "Example: Generating clients for a project using saas-go-kit"
	@echo "1. Copy saas-kit-modules.json.example to your project"
	@echo "2. Modify the module_path fields to point to your local modules"
	@echo "3. Run: make generate-all"