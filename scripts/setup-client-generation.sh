#!/bin/bash

# SaaS Go Kit - Client Generation Setup Script
# This script helps set up client generation in your project

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(pwd)"

echo "🚀 SaaS Go Kit - Setting up Client Generation"
echo "Project directory: $PROJECT_DIR"
echo ""

# Check if we're in a saas-go-kit project
if [ ! -f "go.mod" ]; then
    echo "❌ Error: go.mod not found. Please run this script from your Go project root."
    exit 1
fi

# Create directories
echo "📁 Creating directories..."
mkdir -p generated
mkdir -p scripts

# Copy configuration example
echo "📋 Setting up configuration..."
if [ ! -f "saas-kit-modules.json" ]; then
    cp "$SCRIPT_DIR/../saas-kit-modules.json.example" "saas-kit-modules.json"
    echo "✅ Created saas-kit-modules.json (please customize for your modules)"
else
    echo "⚠️  saas-kit-modules.json already exists, skipping"
fi

# Create local Makefile targets
echo "🔧 Setting up Makefile targets..."
if [ ! -f "Makefile" ]; then
    cat > Makefile << 'EOF'
.PHONY: generate-clients clean-generated

# Generate TypeScript clients
generate-clients:
	@echo "Generating TypeScript clients..."
	@go run github.com/karurosux/saas-go-kit/cmd/generate-routes -config ./saas-kit-modules.json -output ./generated/routes.json
	@go run github.com/karurosux/saas-go-kit/cmd/generate-clients -routes ./generated/routes.json -o ./generated/clients

# Clean generated files
clean-generated:
	@rm -rf ./generated
EOF
    echo "✅ Created Makefile with client generation targets"
else
    echo "⚠️  Makefile already exists. Add these targets manually:"
    echo ""
    echo "generate-clients:"
    echo "	@go run github.com/karurosux/saas-go-kit/cmd/generate-routes -config ./saas-kit-modules.json -output ./generated/routes.json"
    echo "	@go run github.com/karurosux/saas-go-kit/cmd/generate-clients -routes ./generated/routes.json -o ./generated/clients"
    echo ""
fi

# Create package.json for frontend dependencies (if not exists)
echo "📦 Setting up frontend dependencies..."
if [ ! -f "package.json" ]; then
    cat > package.json << 'EOF'
{
  "name": "saas-app-frontend",
  "version": "1.0.0",
  "dependencies": {
    "axios": "^1.6.0"
  },
  "devDependencies": {
    "typescript": "^5.0.0"
  }
}
EOF
    echo "✅ Created package.json with axios dependency"
else
    echo "⚠️  package.json already exists. Make sure to install 'axios' dependency"
fi

# Create .gitignore entries
echo "📝 Updating .gitignore..."
if [ ! -f ".gitignore" ]; then
    touch .gitignore
fi

# Add generated files to gitignore (optional)
if ! grep -q "generated/" .gitignore 2>/dev/null; then
    echo "" >> .gitignore
    echo "# Generated TypeScript clients (uncomment to ignore)" >> .gitignore
    echo "# generated/" >> .gitignore
    echo "✅ Added generated/ to .gitignore (commented out by default)"
fi

echo ""
echo "🎉 Client generation setup complete!"
echo ""
echo "Next steps:"
echo "1. Customize saas-kit-modules.json for your modules"
echo "2. Run: make generate-clients"
echo "3. Import clients in your frontend: import { createClients } from './generated/clients'"
echo ""
echo "For more information, see: https://github.com/karurosux/saas-go-kit/blob/main/CLIENT_GENERATION.md"