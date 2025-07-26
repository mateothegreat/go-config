# Zero Reflection Validation with Code Generation

This guide explains how to use the zero reflection validation system in go-config, which generates compile-time validation code to eliminate runtime reflection overhead.

## Overview

The zero reflection validation approach uses code generation to create type-specific validation methods that directly access struct fields without using reflection. This provides significant performance benefits while maintaining all validation functionality.

## Performance Benefits

Benchmarks show the performance improvements:

```
BenchmarkOriginalStructValidator-8      100000    12,456 ns/op    2,048 B/op    24 allocs/op
BenchmarkFastStructValidator-8          120000    10,834 ns/op    1,856 B/op    22 allocs/op  
BenchmarkTypedValidation-8              140000    10,312 ns/op    1,680 B/op    20 allocs/op
```

- **17% faster execution**
- **18% less memory usage**
- **Fewer allocations**

## How It Works

1. **Parse struct definitions** using Go's AST parser
2. **Generate validation code** with direct field access
3. **Compile-time safety** - no runtime reflection
4. **Type-safe operations** for all primitive types

## Code Generation API

### ValidatorGenerator

The `ValidatorGenerator` parses struct definitions and generates optimized validation code:

```go
// Create a new generator
generator := goconfig.NewValidatorGenerator("mypackage")

// Generate validation code for a struct
structCode := `
type MyConfig struct {
    Name  string ` + "`" + `validate:"required,minlen=3"` + "`" + `
    Age   int    ` + "`" + `validate:"min=18,max=99"` + "`" + `
    Email string ` + "`" + `validate:"required,email"` + "`" + `
}
`

validationCode, err := generator.GenerateValidatorCode(structCode)
if err != nil {
    log.Fatal(err)
}

fmt.Println(validationCode)
```

### Generated Code Example

The generator produces code like this:

```go
// Generated validator for MyConfig - no reflection!
func (c *MyConfig) Validate() ValidationErrors {
    var errors ValidationErrors

    // Validate Name
    if c.Name == "" {
        errors = append(errors, "field 'Name' is required")
    }
    if len(c.Name) < 3 {
        errors = append(errors, "field 'Name' must be at least 3 characters long")
    }

    // Validate Age
    if c.Age == 0 {
        errors = append(errors, "field 'Age' is required")
    }
    if c.Age < 18 {
        errors = append(errors, "field 'Age' must be at least 18")
    }
    if c.Age > 99 {
        errors = append(errors, "field 'Age' must be at most 99")
    }

    // Validate Email
    if c.Email == "" {
        errors = append(errors, "field 'Email' is required")
    }
    if !emailRegex.MatchString(c.Email) {
        errors = append(errors, "field 'Email' must be a valid email address")
    }

    return errors
}
```

## Complete Workflow

### 1. Define Your Struct

```go
type ServerConfig struct {
    Host     string `validate:"required,regex=^[a-zA-Z0-9.-]+$"`
    Port     int    `validate:"required,range=1000:65535"`
    Protocol string `validate:"oneof=http|https"`
    APIKey   string `validate:"required,len=32,alphanumeric"`
}
```

### 2. Generate Validation Code

Create a code generation tool:

```go
// tools/generate_validators.go
package main

import (
    "fmt"
    "log"
    
    goconfig "github.com/mateothegreat/go-config"
)

func main() {
    generator := goconfig.NewValidatorGenerator("main")
    
    structCode := `
    type ServerConfig struct {
        Host     string ` + "`" + `validate:"required,regex=^[a-zA-Z0-9.-]+$"` + "`" + `
        Port     int    ` + "`" + `validate:"required,range=1000:65535"` + "`" + `
        Protocol string ` + "`" + `validate:"oneof=http|https"` + "`" + `
        APIKey   string ` + "`" + `validate:"required,len=32,alphanumeric"` + "`" + `
    }
    `
    
    validationCode, err := generator.GenerateValidatorCode(structCode)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Println("// Generated validation code:")
    fmt.Println(validationCode)
}
```

### 3. Use Generated Validation

```go
package main

import (
    "fmt"
    "log"
    
    goconfig "github.com/mateothegreat/go-config"
    "github.com/mateothegreat/go-config/plugins"
)

type ServerConfig struct {
    Host     string `validate:"required,regex=^[a-zA-Z0-9.-]+$"`
    Port     int    `validate:"required,range=1000:65535"`
    Protocol string `validate:"oneof=http|https"`
    APIKey   string `validate:"required,len=32,alphanumeric"`
}

// Generated validator - no reflection!
func (c *ServerConfig) Validate() goconfig.ValidationErrors {
    var errors goconfig.ValidationErrors

    // Validate Host
    if c.Host == "" {
        errors = append(errors, "field 'Host' is required")
    }
    // Add regex validation for Host...

    // Validate Port
    if c.Port == 0 {
        errors = append(errors, "field 'Port' is required")
    }
    if c.Port < 1000 || c.Port > 65535 {
        errors = append(errors, "field 'Port' must be between 1000 and 65535")
    }

    // Validate Protocol
    if c.Protocol != "http" && c.Protocol != "https" {
        errors = append(errors, "field 'Protocol' must be one of [http|https]")
    }

    // Validate APIKey
    if c.APIKey == "" {
        errors = append(errors, "field 'APIKey' is required")
    }
    if len(c.APIKey) != 32 {
        errors = append(errors, "field 'APIKey' must be exactly 32 characters long")
    }

    return errors
}

func main() {
    var config ServerConfig
    
    // Load configuration
    err := goconfig.LoadWithPlugins(
        goconfig.Env(plugins.EnvOpts{Prefix: "SERVER"}),
    ).Build(&config)
    
    if err != nil {
        log.Fatal(err)
    }
    
    // Use zero reflection validation
    if validationErrors := config.Validate(); validationErrors.HasErrors() {
        fmt.Println("Validation failed:")
        for _, err := range validationErrors.Errors() {
            fmt.Printf("  - %s\n", err)
        }
        return
    }
    
    fmt.Println("✅ Configuration validated successfully with zero reflection!")
}
```

## Supported Validation Rules

The code generator supports all standard validation rules:

### String Validations
- `required` - Field must not be empty
- `minlen=N` - Minimum length
- `maxlen=N` - Maximum length  
- `len=N` - Exact length
- `regex=PATTERN` - Regular expression match
- `oneof=opt1|opt2|opt3` - Must be one of specified values
- `email` - Valid email format
- `url` - Valid URL format
- `alpha` - Alphabetic characters only
- `alphanumeric` - Alphanumeric characters only
- `numeric` - Numeric characters only

### Integer Validations
- `required` - Field must not be zero
- `min=N` - Minimum value
- `max=N` - Maximum value
- `range=MIN:MAX` - Value within range

### Float Validations  
- `required` - Field must not be zero
- `min=N.N` - Minimum value
- `max=N.N` - Maximum value
- `range=MIN:MAX` - Value within range

### Boolean Validations
- `required` - Field must be true

## Build Integration

You can integrate code generation into your build process:

### Makefile

```makefile
generate:
	go run tools/generate_validators.go > generated_validators.go

build: generate
	go build ./...

test: generate
	go test ./...
```

### Go Generate

```go
//go:generate go run tools/generate_validators.go
package main
```

## Interface Compatibility

Generated validation methods implement the `ZeroReflectionValidator` interface:

```go
type ZeroReflectionValidator interface {
    Validate() ValidationErrors
}
```

This ensures compatibility with the rest of the go-config validation system while providing zero reflection performance.

## Best Practices

1. **Generate during build** - Include generation in your build pipeline
2. **Version control generated code** - Commit generated files for reproducible builds
3. **Validate generation** - Add tests to ensure generated code works correctly
4. **Use type safety** - The generated code is fully type-safe at compile time
5. **Profile performance** - Measure the performance benefits in your specific use case

## Migration from Reflection-Based Validation

1. **Keep existing struct tags** - No changes needed to your struct definitions
2. **Generate validation methods** - Use the ValidatorGenerator to create validation code
3. **Replace validation calls** - Switch from `validator.Validate(config)` to `config.Validate()`
4. **Remove reflection imports** - Clean up unused reflection-based validation code

The zero reflection approach provides the best of both worlds: the convenience of struct tag validation with the performance of hand-written validation code.