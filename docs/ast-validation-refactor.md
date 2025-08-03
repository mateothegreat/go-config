# AST-Based Validation Code Generation Refactor

## Overview

This document describes the refactoring of the go-config library's validation code generation system to use pure AST-based generation instead of text templates, integrating with the [go-validation](https://github.com/mateothegreat/go-validation) library.

## Key Changes

### 1. New AST-Based Code Generator (`generator/ast_codegen.go`)

- **Pure AST Generation**: All code is generated using Go's `go/ast` package
- **No Text Templates**: Completely eliminates string-based template generation
- **Type-Safe**: AST manipulation ensures syntactically correct code
- **Performance**: Zero allocations for cached validator expressions

### 2. Validator Registry (`generator/validator_registry.go`)

- **Centralized Validator Mapping**: Maps validation rules to go-validation library functions
- **Comprehensive Coverage**: Supports all go-validation built-in validators:
  - Basic: required, email, url, alpha, alphanum, numeric
  - Numeric: min, max, range
  - String: minlen, maxlen, len
  - Network: ip, ipv4, ipv6, cidr, mac, hostname
  - Format: uuid, datetime, base64, creditcard, phone
  - Conditional: oneof, required_if, required_unless, required_with, required_without
  - Cross-field: eqfield, nefield, gtfield, ltfield
  - Collections: dive
  - Patterns: regex

### 3. Integration with go-validation Library

- **Registry Integration**: Generated code registers validators with go-validation's registry
- **Direct Function Calls**: Validators are called directly without reflection
- **Fluent Error Builder**: Uses the fluent error interface for clear error messages

## Architecture

```
┌─────────────────────┐
│   Config Structs    │
│  (with validate     │
│    tags)            │
└──────────┬──────────┘
           │
           ▼
┌─────────────────────┐
│   AST Scanner       │
│  (scanner/ast.go)   │
└──────────┬──────────┘
           │
           ▼
┌─────────────────────┐
│  AST Code Generator │
│ (ast_codegen.go)    │
└──────────┬──────────┘
           │
           ▼
┌─────────────────────┐
│ Validator Registry  │
│(validator_registry) │
└──────────┬──────────┘
           │
           ▼
┌─────────────────────┐
│  Generated Go Code  │
│  (Zero Reflection)  │
└─────────────────────┘
```

## Generated Code Example

For a struct like:

```go
type ServerConfig struct {
    Name  string `validate:"required,minlen=3,maxlen=50"`
    Port  int    `validate:"required,min=1,max=65535"`
    Email string `validate:"required,email"`
}
```

The AST generator produces:

```go
package config

import (
    validation "github.com/mateothegreat/go-validation"
    "github.com/mateothegreat/go-validation/rules"
    "github.com/mateothegreat/go-config/errors"
)

// init registers generated validators with the go-validation registry
func init() {
    validation.RegisterValidation("serverconfig", ValidateServerConfig)
}

// Validate validates the ServerConfig struct using zero-reflection validation
func (s *ServerConfig) Validate() error {
    var errs *errors.FluentError = errors.NewError()
    
    // Validate Name
    if !rules.Required(validation.FieldLevel("Name", s.Name)) {
        errs.Field("Name").Required()
    }
    if !rules.MinLength(3)(validation.FieldLevel("Name", s.Name)) {
        errs.Field("Name").MinLength(3)
    }
    if !rules.MaxLength(50)(validation.FieldLevel("Name", s.Name)) {
        errs.Field("Name").MaxLength(50)
    }
    
    // Validate Port
    if !rules.Required(validation.FieldLevel("Port", s.Port)) {
        errs.Field("Port").Required()
    }
    if !rules.Min(1)(validation.FieldLevel("Port", s.Port)) {
        errs.Field("Port").Min(1)
    }
    if !rules.Max(65535)(validation.FieldLevel("Port", s.Port)) {
        errs.Field("Port").Max(65535)
    }
    
    // Validate Email
    if !rules.Required(validation.FieldLevel("Email", s.Email)) {
        errs.Field("Email").Required()
    }
    if !rules.Email(validation.FieldLevel("Email", s.Email)) {
        errs.Field("Email").Email()
    }
    
    if errs.HasErrors() {
        return errs
    }
    return nil
}

// ValidateServerConfig validates ServerConfig instances for use with go-validation registry
func ValidateServerConfig(fl validation.FieldLevel) bool {
    value := fl.Field()
    v, ok := value.(*ServerConfig)
    if !ok {
        return false
    }
    return v.Validate() == nil
}
```

## Performance Benefits

1. **Compile-Time Generation**: All validation code is generated at compile time
2. **Zero Runtime Reflection**: No reflection is used during validation
3. **Direct Function Calls**: Validators are called directly, not through interfaces
4. **Cached Expressions**: Validator AST expressions are cached during generation
5. **Minimal Allocations**: Error objects only created when validation fails

## Usage

### Command Line

```bash
# Generate with AST (default)
go-config-gen

# Generate for specific structs
go-config-gen -structs ServerConfig,DatabaseConfig

# Generate separate files for each struct
go-config-gen -multi

# Preview generated code
go-config-gen -dry-run -verbose
```

### Go Generate

```go
//go:generate go-config-gen -ast
```

## Extending the System

### Adding Custom Validators

1. Register the validator with go-validation:
   ```go
   validation.RegisterValidation("myvalidator", MyValidatorFunc)
   ```

2. Add to the ValidatorRegistry:
   ```go
   r.validators["myvalidator"] = ValidatorInfo{
       Name:         "myvalidator",
       PackagePath:  "mypackage",
       FunctionName: "MyValidator",
       RequiresArgs: false,
   }
   ```

### Creating Specialized Generators

The AST-based approach makes it easy to create specialized generators for specific use cases:

```go
type SpecializedGenerator struct {
    *ASTCodeGenerator
}

func (g *SpecializedGenerator) GenerateCustomValidation() ast.Expr {
    // Create custom AST nodes
}
```

## Migration from Template-Based Generation

1. Update go-config-gen command flags (AST is now default)
2. Regenerate validation files
3. No changes needed to struct definitions
4. Generated code is fully compatible

## Future Enhancements

1. **Struct-Level Validation**: Support for complex business logic validation
2. **Custom Error Messages**: Support for custom error message templates
3. **Validation Groups**: Support for conditional validation groups
4. **Performance Optimizations**: Further optimize generated code paths 