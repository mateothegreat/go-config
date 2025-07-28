# Default targets to run when no arguments are provided to `make`.
all: generate verify run

# Install go-validate if not present.
install-tools:
	@echo "🛠️  Installing go-validate..."
	go install github.com/mateothegreat/go-config/cmd/go-validate@latest
	@echo "✅ Tools installed"

# Preview validation code without writing files.
preview:
	@echo "👀 Previewing validation code..."
	go-validate generate --dry-run --verbose

# Generate validation code.
generate:
	@echo "🔄 Generating validation code..."
	go-validate generate --verbose
	@echo "✅ Validation code generated"

# Generate for specific structs only.
generate-specific:
	@echo "🎯 Generating validation for specific structs..."
	@echo "Structs: $(STRUCTS)"
	go-validate generate --structs $(STRUCTS) --verbose

# Generate to separate directory with multiple files.
generate-multi:
	@echo "📁 Generating validation to separate directory..."
	mkdir -p validation
	go-validate generate --output-dir validation --multi --verbose
	@echo "✅ Multi-file validation generated in ./validation/"

# Generate with custom package.
generate-custom-package:
	@echo "📦 Generating validation with custom package..."
	mkdir -p generated
	go-validate generate --output-dir generated --package validation --verbose

# Clean generated files.
clean:
	@echo "🧹 Cleaning generated files..."
	rm -f *_generated.go
	rm -rf validation/
	rm -rf generated/
	@echo "✅ Cleanup complete"

# Verify generated code compiles.
verify: generate
	@echo "🔍 Verifying generated code..."
	go build -o /dev/null .
	go vet .
	@echo "✅ Generated code verified"

# Run the example.
run: generate
	@echo "🚀 Running CLI usage example..."
	go run .

# Development workflow with file watching (requires fswatch).
watch:
	@echo "👀 Watching for changes..."
	@command -v fswatch >/dev/null 2>&1 || (echo "fswatch not found. Install with: brew install fswatch" && exit 1)
	fswatch -o *.go | xargs -n1 -I{} make generate

# Test the generated validation.			
test: generate
	@echo "🧪 Testing generated validation..."
	go test -v .

# Benchmark the generated validation.
bench: generate
	@echo "📊 Benchmarking generated validation..."
	go test -bench=. -benchmem .

# Show differences between preview and actual generation.
diff-preview:
	@echo "🔍 Comparing preview vs actual generation..."
	@go-validate generate --dry-run > /tmp/preview.go
	@go-validate generate --output-dir /tmp --verbose >/dev/null 2>&1
	@diff -u /tmp/preview.go /tmp/*_generated.go || true
	@rm -f /tmp/preview.go /tmp/*_generated.go

# Validate that generated code is up to date.
validate-current:
	@echo "🔍 Validating current generated code..."
	@go-validate generate --dry-run > /tmp/current.go
	@if diff -q *_generated.go /tmp/current.go >/dev/null 2>&1; then \
		echo "✅ Generated code is up to date"; \
	else \
		echo "❌ Generated code is out of date. Run 'make generate'"; \
		exit 1; \
	fi
	@rm -f /tmp/current.go

# Full CI pipeline.
ci: clean generate verify test
	@echo "✅ CI pipeline completed successfully"

# Help target.
help:
	@echo "CLI Usage Example Makefile"
	@echo "=========================="
	@echo ""
	@echo "Main targets:"
	@echo "  preview              - Preview generated code without writing files"
	@echo "  generate             - Generate validation code"
	@echo "  clean                - Remove generated files"
	@echo "  verify               - Generate and verify code compiles"
	@echo "  run                  - Generate and run example"
	@echo ""
	@echo "Advanced targets:"
	@echo "  generate-specific    - Generate for specific structs (set STRUCTS=...)"
	@echo "  generate-multi       - Generate separate files per struct"
	@echo "  generate-custom-package - Generate with custom package name"
	@echo "  watch                - Watch for changes and regenerate"
	@echo "  test                 - Test generated validation"
	@echo "  bench                - Benchmark generated validation"
	@echo "  diff-preview         - Compare preview vs actual generation"
	@echo "  validate-current     - Check if generated code is up to date"
	@echo "  ci                   - Full CI pipeline"
	@echo "  install-tools        - Install required tools"
	@echo ""
	@echo "Examples:"
	@echo "  make generate-specific STRUCTS=ServerConfig,APIConfig"
	@echo "  make preview"
	@echo "  make clean generate run"