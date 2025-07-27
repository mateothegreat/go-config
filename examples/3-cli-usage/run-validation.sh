#!/bin/bash
# Manual validation generation script

set -e

echo "🔧 Manual CLI Validation Generation"
echo "=================================="
echo

# Check if go-validate is installed
if ! command -v go-validate &> /dev/null; then
    echo "❌ go-validate not found. Installing..."
    go install github.com/mateothegreat/go-config/cmd/go-validate@latest
    echo "✅ go-validate installed"
    echo
fi

# Step 1: Preview what will be generated
echo "👀 Step 1: Previewing validation code..."
echo "Command: go-validate generate --dry-run --verbose"
echo "---"
go-validate generate --dry-run --verbose
echo
echo "---"
echo

# Step 2: Generate validation code
echo "🔄 Step 2: Generating validation code..."
echo "Command: go-validate generate --verbose"
echo "---"
go-validate generate --verbose
echo "---"
echo

# Step 3: Verify the generated code compiles
echo "🔍 Step 3: Verifying generated code..."
echo "Command: go build -o /dev/null ."
go build -o /dev/null .
echo "✅ Generated code compiles successfully!"
echo

# Step 4: Show what files were generated
echo "📁 Step 4: Generated files:"
ls -la *_generated.go 2>/dev/null || echo "No generated files found (check if structs have validation tags)"
echo

# Step 5: Run the example
echo "🚀 Step 5: Running the example..."
echo "Command: go run ."
echo "---"
go run .
echo "---"
echo

echo "✅ Manual CLI validation demo completed successfully!"
echo
echo "💡 Try these additional commands:"
echo "   go-validate generate --structs ServerConfig --dry-run"
echo "   go-validate generate --output-dir ./validation --multi"
echo "   go-validate generate --help"