# CLI Usage Example - Manual Approach

This example demonstrates the traditional approach of using the `go-validate` CLI tool manually without go:generate directives.

## 📋 What You'll Learn

- Manual CLI usage patterns
- Command-line flags and options
- Dry-run and preview functionality
- Output control and debugging
- Integration with build scripts

## 🏗️ Files

- `config.go` - Configuration struct (no go:generate directives)
- `main.go` - Example usage with manual validation
- `run-validation.sh` - Shell script for manual generation
- `Makefile` - Build automation with manual steps
- `config.yaml` - Sample configuration

## 🚀 Quick Start

1. **Manual generation:**
   ```bash
   go-validate generate
   ```

2. **Preview generated code:**
   ```bash
   go-validate generate --dry-run --verbose
   ```

3. **Run the example:**
   ```bash
   go run .
   ```

## 📖 Manual CLI Workflow

### 1. Basic Generation

```bash
# Generate validation for all structs in current directory
go-validate generate

# Generate with verbose output
go-validate generate --verbose

# Generate for specific structs
go-validate generate --structs ServerConfig,DatabaseConfig
```

### 2. Preview and Debugging

```bash
# Preview generated code without writing files
go-validate generate --dry-run

# Verbose preview with detailed output
go-validate generate --dry-run --verbose

# Preview for specific structs
go-validate generate --structs ServerConfig --dry-run --verbose
```

### 3. Output Control

```bash
# Generate to custom directory
go-validate generate --output-dir ./generated

# Generate multiple files (one per struct)
go-validate generate --multi

# Generate to specific package
go-validate generate --package validation
```

### 4. Input Control

```bash
# Scan specific directory
go-validate generate --input-dir ./internal/config

# Combine input and output directories
go-validate generate --input-dir ./config --output-dir ./validation
```

## 🔧 Shell Script Integration

### Manual Generation Script

```bash
#!/bin/bash
# run-validation.sh

set -e

echo "🔄 Generating validation code..."

# Clean previous generated files
rm -f *_generated.go

# Generate validation code with verbose output
go-validate generate --verbose

echo "✅ Validation code generated successfully!"

# Verify the generated code compiles
echo "🔍 Verifying generated code..."
go build -o /dev/null .

echo "✅ Generated code verified!"
```

### Advanced Generation Script

```bash
#!/bin/bash
# advanced-generate.sh

set -e

CONFIG_DIR="./config"
OUTPUT_DIR="./validation"
STRUCTS="ServerConfig,DatabaseConfig,APIConfig"

echo "🔄 Advanced validation generation..."

# Create output directory
mkdir -p "$OUTPUT_DIR"

# Preview first
echo "👀 Previewing generated code..."
go-validate generate \
    --input-dir "$CONFIG_DIR" \
    --output-dir "$OUTPUT_DIR" \
    --structs "$STRUCTS" \
    --dry-run \
    --verbose

# Confirm generation
read -p "Generate code? (y/N): " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    go-validate generate \
        --input-dir "$CONFIG_DIR" \
        --output-dir "$OUTPUT_DIR" \
        --structs "$STRUCTS" \
        --verbose
    echo "✅ Code generated successfully!"
else
    echo "❌ Generation cancelled"
    exit 1
fi
```

## 🔧 Makefile Integration

```makefile
.PHONY: validate-preview validate-generate validate-clean help

# Preview validation code
validate-preview:
	@echo "👀 Previewing validation code..."
	go-validate generate --dry-run --verbose

# Generate validation code
validate-generate:
	@echo "🔄 Generating validation code..."
	go-validate generate --verbose
	@echo "✅ Validation code generated"

# Generate for specific structs
validate-specific:
	@echo "🔄 Generating validation for specific structs..."
	go-validate generate --structs $(STRUCTS) --verbose

# Generate to separate directory
validate-separate:
	@echo "🔄 Generating validation to separate directory..."
	mkdir -p validation
	go-validate generate --output-dir validation --multi --verbose

# Clean generated files
validate-clean:
	@echo "🧹 Cleaning validation files..."
	rm -f *_generated.go
	rm -rf validation/

# Verify generated code
validate-verify: validate-generate
	@echo "🔍 Verifying generated code..."
	go build -o /dev/null .
	@echo "✅ Generated code verified"

# Development workflow
dev: validate-clean validate-generate
	@echo "🚀 Running application..."
	go run .

# Help
help:
	@echo "Available targets:"
	@echo "  validate-preview   - Preview generated code"
	@echo "  validate-generate  - Generate validation code"
	@echo "  validate-specific  - Generate for specific structs (set STRUCTS=...)"
	@echo "  validate-separate  - Generate to separate directory"
	@echo "  validate-clean     - Clean generated files"
	@echo "  validate-verify    - Generate and verify code"
	@echo "  dev                - Clean, generate, and run"
	@echo "  help               - Show this help"
```

## 🎯 CLI Flag Reference

### Core Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--input-dir` | string | `.` | Directory to scan for structs |
| `--output-dir` | string | `.` | Directory for generated files |
| `--structs` | []string | all | Comma-separated struct names |
| `--package` | string | auto | Package name for generated code |

### Output Control

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--dry-run` | bool | false | Preview without writing files |
| `--verbose` | bool | false | Enable verbose output |
| `--multi` | bool | false | Generate separate files per struct |

### Examples by Flag Combination

```bash
# Basic usage
go-validate generate

# Verbose with preview
go-validate generate --dry-run --verbose

# Specific structs to custom directory
go-validate generate \
    --structs "Config,Settings" \
    --output-dir ./validation \
    --verbose

# Multi-file generation
go-validate generate \
    --multi \
    --output-dir ./generated \
    --verbose

# Cross-package generation
go-validate generate \
    --input-dir ./internal/config \
    --output-dir ./pkg/validation \
    --package validation
```

## 🔍 Debugging Generated Code

### Preview Generated Output

```bash
# Basic preview
go-validate generate --dry-run

# Detailed preview with metadata
go-validate generate --dry-run --verbose
```

### Verify Generated Code

```bash
# Generate and immediately verify
go-validate generate && go build -o /dev/null .

# Check for syntax errors
go-validate generate && go vet .

# Run tests on generated code
go-validate generate && go test .
```

### Troubleshooting Common Issues

```bash
# Check if structs are detected
go-validate generate --dry-run --verbose | grep "Found struct"

# Verify package detection
go-validate generate --dry-run --verbose | grep "Package"

# Check validation rules parsing
go-validate generate --dry-run --verbose | grep "Field"
```

## 🔄 CI/CD Integration

### GitHub Actions

```yaml
name: Validate Generated Code

on: [push, pull_request]

jobs:
  validate:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    
    - name: Setup Go
      uses: actions/setup-go@v3
      with:
        go-version: '1.21'
    
    - name: Install go-validate
      run: go install github.com/mateothegreat/go-config/cmd/go-validate@latest
    
    - name: Generate validation code
      run: go-validate generate --verbose
    
    - name: Verify no changes
      run: git diff --exit-code
    
    - name: Test generated code
      run: go test ./...
```

### Build Pipeline

```bash
#!/bin/bash
# ci-validate.sh

set -e

echo "🔄 CI/CD Validation Pipeline"

# Install dependencies
go mod download

# Generate validation code
go-validate generate --verbose

# Verify generated code is up to date
if ! git diff --quiet; then
    echo "❌ Generated code is out of date"
    echo "Run: go-validate generate"
    exit 1
fi

# Build with generated code
go build ./...

# Test with generated code
go test ./...

echo "✅ Validation pipeline passed"
```

## 🎓 When to Use Manual CLI

**Use manual CLI approach when:**

1. **Learning the tool** - Understanding what gets generated
2. **Complex build systems** - Custom build pipelines
3. **CI/CD integration** - Automated validation in pipelines
4. **Cross-package generation** - Generating validation across multiple packages
5. **Custom workflows** - Non-standard development workflows

**Consider go:generate for:**

1. **Standard development** - Most common use case
2. **IDE integration** - Automatic generation on build
3. **Team consistency** - Standardized approach
4. **Simple workflows** - Straightforward development

## 🎓 Next Steps

- Learn automated generation with [`../02-go-generate/`](../2-embedded-generation)
- See real-world usage in [`../04-microservice-config/`](../04-microservice-config)
- Explore multi-package patterns in [`../09-multi-package/`](../09-multi-package)