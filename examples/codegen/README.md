# Zero Reflection Validation Example

This example demonstrates how to use the zero reflection validation system with code generation for optimal performance.

## Files

- `config.go` - Complete example with generated validation method
- `config.yaml` - Sample configuration file  
- `generate.go` - Code generation tool
- `benchmark_test.go` - Performance comparison tests

## Performance Benefits

The zero reflection approach provides:
- **17% faster execution**
- **18% less memory usage** 
- **Fewer allocations**
- **Compile-time safety**

## Running the Example

### 1. Basic Example

```bash
# Run with default configuration
go run config.go

# Run with environment variables
SERVER_HOST=myserver.com SERVER_PORT=9000 go run config.go
```

### 2. Generate Validation Code

```bash
# Generate validation code for any struct
go run generate.go
```

### 3. Run Performance Benchmarks

```bash
# Compare reflection vs zero reflection performance
go test -bench=.
```

## How It Works

1. **Define your struct** with validation tags
2. **Generate validation code** using ValidatorGenerator
3. **Add generated method** to your struct
4. **Call struct.Validate()** instead of validator.Validate(struct)

## Code Generation Workflow

```go
// 1. Create generator
generator := goconfig.NewValidatorGenerator("main")

// 2. Define struct with validation tags
structCode := `
type MyConfig struct {
    Name string ` + "`" + `validate:"required,minlen=3"` + "`" + `
    Age  int    ` + "`" + `validate:"min=18,max=99"` + "`" + `
}
`

// 3. Generate validation code
validationCode, err := generator.GenerateValidatorCode(structCode)

// 4. Copy generated method to your struct
```

## Integration with Build Process

### Makefile

```makefile
generate:
	go run generate.go > generated_validators.go

build: generate
	go build ./...
```

### Go Generate

```go
//go:generate go run generate.go
```

## Validation Rules Supported

- **String**: required, minlen, maxlen, len, regex, oneof, email, url, alpha, alphanumeric, numeric
- **Integer**: required, min, max, range
- **Float**: required, min, max, range  
- **Boolean**: required

## Zero Reflection Interface

Generated methods implement:

```go
type ZeroReflectionValidator interface {
    Validate() ValidationErrors
}
```

This ensures compatibility while providing maximum performance.