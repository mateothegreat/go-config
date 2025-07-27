# Makefile for go-generate validation example

.PHONY: generate build run test clean help

# Default target
all: generate build

# Generate validation code using go generate
generate:
	@echo "🔄 Generating validation code..."
	go generate ./...
	@echo "✅ Code generation complete"

# Build the application
build: generate
	@echo "🔨 Building application..."
	go build -o bin/app .
	@echo "✅ Build complete"

# Run the application
run: generate
	@echo "🚀 Running application..."
	go run .

# Run tests
test: generate
	@echo "🧪 Running tests..."
	go test -v ./...

# Run benchmarks
bench: generate
	@echo "📊 Running benchmarks..."
	go test -bench=. -benchmem ./...

# Preview generated code without writing files
preview:
	@echo "👀 Previewing generated code..."
	go-validate generate --dry-run --verbose

# Clean generated files and build artifacts
clean:
	@echo "🧹 Cleaning up..."
	rm -f *_generated.go
	rm -rf bin/
	@echo "✅ Cleanup complete"

# Verify code generation is up to date (useful for CI)
verify: generate
	@echo "🔍 Verifying generated code is up to date..."
	git diff --exit-code -- '*_generated.go' || (echo "❌ Generated code is out of date. Run 'make generate'" && exit 1)
	@echo "✅ Generated code is up to date"

# Install go-validate if not present
install-tools:
	@echo "🛠️  Installing go-validate..."
	go install github.com/mateothegreat/go-config/cmd/go-validate@latest
	@echo "✅ Tools installed"

# Development workflow: watch for changes and regenerate
watch:
	@echo "👀 Watching for changes..."
	@command -v fswatch >/dev/null 2>&1 || (echo "fswatch not found. Install with: brew install fswatch" && exit 1)
	fswatch -o . --exclude='.*_generated\.go$$' | xargs -n1 -I{} make generate

# Help target
help:
	@echo "Available targets:"
	@echo "  generate    - Generate validation code using go generate"
	@echo "  build       - Build the application (includes generate)"
	@echo "  run         - Run the application (includes generate)"
	@echo "  test        - Run tests (includes generate)"
	@echo "  bench       - Run benchmarks (includes generate)"
	@echo "  preview     - Preview generated code without writing files"
	@echo "  clean       - Remove generated files and build artifacts"
	@echo "  verify      - Verify generated code is up to date"
	@echo "  install-tools - Install required tools"
	@echo "  watch       - Watch for changes and regenerate (requires fswatch)"
	@echo "  help        - Show this help message"