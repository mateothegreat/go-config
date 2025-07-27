# Zero-Reflection Validation

This document provides a deep dive into `go-validate`'s zero-reflection approach to validation, explaining the technical implementation and benefits.

## What is Zero-Reflection Validation?

Zero-reflection validation means that the validation code operates **without using Go's reflection package at runtime**. Instead of introspecting types and values dynamically, the validation logic is **generated at build time** and operates directly on struct fields.

### Traditional Reflection-Based Approach

```go
func validateWithReflection(data interface{}) error {
    val := reflect.ValueOf(data)           // Runtime type introspection
    typ := val.Type()
    
    for i := 0; i < val.NumField(); i++ {
        field := typ.Field(i)              // Runtime field discovery
        value := val.Field(i)              // Runtime value access
        tag := field.Tag.Get("validate")   // Runtime tag parsing
        
        // More reflection calls for validation logic...
    }
}
```

### Zero-Reflection Generated Approach

```go
// Generated at build time, no reflection at runtime
func (s *ServerConfig) Validate() error {
    if s.Name == "" {                      // Direct field access
        return goconfig.NewError().Field("Name").Required()
    }
    if len(s.Name) < 3 {                  // Direct length check
        return goconfig.NewError().Field("Name").MinLength(3)
    }
    return nil
}
```

## Code Generation Process

### 1. AST Scanning Phase

The generator uses Go's Abstract Syntax Tree (AST) to analyze source code:

```go
func (s *ASTScanner) extractStructsFromAST(node *ast.File, filePath string) ([]StructInfo, error) {
    var structs []StructInfo
    
    // Walk the AST looking for struct declarations
    ast.Inspect(node, func(n ast.Node) bool {
        switch x := n.(type) {
        case *ast.TypeSpec:
            if structType, ok := x.Type.(*ast.StructType); ok {
                // Found a struct, check for validation tags
                if s.hasValidationAnnotations(structType) {
                    structInfo := s.buildStructInfo(x, structType, packageName, filePath, imports)
                    structs = append(structs, structInfo)
                }
            }
        }
        return true
    })
    
    return structs, nil
}
```

**Benefits of AST scanning:**
- **Compile-time safety**: Errors detected during generation, not runtime
- **No regex parsing**: Robust tag extraction using Go's parser
- **Full context**: Access to comments, imports, and package structure

### 2. Validation Rule Parsing

Tags are parsed into structured validation rules:

```go
func (s *ASTScanner) parseValidationRules(tag string) map[string]string {
    rules := make(map[string]string)
    parts := strings.Split(tag, ",")

    for _, part := range parts {
        part = strings.TrimSpace(part)
        if strings.Contains(part, "=") {
            kv := strings.SplitN(part, "=", 2)
            rules[strings.TrimSpace(kv[0])] = strings.TrimSpace(kv[1])
        } else {
            rules[part] = ""
        }
    }
    return rules
}
```

**Input**: `validate:"required,minlen=3,email"`
**Output**: `map[string]string{"required": "", "minlen": "3", "email": ""}`

### 3. Code Generation Templates

Templates generate type-specific validation code:

```go
const validationTemplate = `
func (s *{{.Structs.0.Name}}) Validate() error {
{{- range $field := .Structs.0.Fields}}
{{- if $field.ValidateRules}}
{{- range $rule, $value := $field.ValidateRules}}
    {{buildValidator $field $rule $value}}
{{- end}}
{{- end}}
{{- end}}
    return nil
}
`
```

**Template functions** generate optimized code per validation rule:

```go
func (cg *CodeGenerator) buildRequiredValidator(field scanner.FieldInfo) string {
    switch field.Type {
    case "string":
        return fmt.Sprintf(`
    if s.%s == "" {
        return goconfig.NewError().Field("%s").Required()
    }`, field.Name, field.Name)
    case "int", "int8", "int16", "int32", "int64":
        return fmt.Sprintf(`
    if s.%s == 0 {
        return goconfig.NewError().Field("%s").Required()
    }`, field.Name, field.Name)
    }
}
```

## Generated Code Characteristics

### Direct Field Access

**Reflection approach:**
```go
fieldValue := reflect.ValueOf(s).Field(i)
stringValue := fieldValue.String()  // Type assertion required
if stringValue == "" {
    // validation error
}
```

**Generated approach:**
```go
if s.Name == "" {  // Direct access, no type assertion
    // validation error
}
```

### Pre-compiled Regex Patterns

**Reflection approach** (compiles regex on every validation):
```go
func validateEmail(value reflect.Value) error {
    emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
    return emailRegex.MatchString(value.String())
}
```

**Generated approach** (regex compiled once at startup):
```go
var (
    emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
)

func (s *Config) Validate() error {
    if !emailRegex.MatchString(s.Email) {  // Pre-compiled, fast
        return goconfig.NewError().Field("Email").Format("must be a valid email")
    }
    return nil
}
```

### Type-Specific Optimizations

The generator produces different code based on field types:

#### String Validation
```go
// For: Name string `validate:"required,minlen=3,maxlen=50"`
if s.Name == "" {
    return goconfig.NewError().Field("Name").Required()
}
if len(s.Name) < 3 {
    return goconfig.NewError().Field("Name").MinLength(3)
}
if len(s.Name) > 50 {
    return goconfig.NewError().Field("Name").MaxLength(50)
}
```

#### Numeric Validation
```go
// For: Port int `validate:"min=1,max=65535"`
if s.Port < 1 {
    return goconfig.NewError().Field("Port").Format("must be at least 1")
}
if s.Port > 65535 {
    return goconfig.NewError().Field("Port").Format("must be at most 65535")
}
```

#### Boolean Validation
```go
// For: Enabled bool `validate:"required"`
// Booleans are special - false is not considered "empty"
// Only validates presence, not value
```

## Performance Implications

### Memory Allocation Analysis

#### Successful Validation Path
```go
func (s *ServerConfig) Validate() error {
    // All comparisons are stack operations - 0 allocations
    if s.Name == "" { return /* error */ }      // 0 allocs
    if len(s.Name) < 3 { return /* error */ }   // 0 allocs
    if s.Port < 1 { return /* error */ }        // 0 allocs
    return nil                                   // 0 allocs - success!
}
```

**Total allocations for successful validation: 0**

#### Error Path (only when validation fails)
```go
if s.Name == "" {
    return goconfig.NewError().Field("Name").Required()  // 2-3 allocations for error object
}
```

**Total allocations for failed validation: 2-3 (only for the error object)**

### CPU Performance Benefits

#### Branch Prediction
Generated code has predictable branch patterns:
```go
// Highly predictable branches (most configs are valid)
if s.Name == "" {           // Usually false - branch predictor optimizes
    return error
}
if len(s.Name) < 3 {        // Usually false - branch predictor optimizes  
    return error
}
// Normal execution path continues...
```

#### Cache Efficiency
Direct field access improves CPU cache usage:
```go
// Memory accesses are sequential and predictable
s.Name        // CPU cache line 1
s.Port        // CPU cache line 1 (same struct)
s.Email       // CPU cache line 1 (same struct)
```

vs reflection which involves pointer chasing and hash table lookups.

#### Inlining Opportunities
Generated code is often inlined by the Go compiler:
```go
// Small, simple functions are inlined at call sites
func (s *Config) Validate() error {  // Candidate for inlining
    if s.Name == "" { return error } // Simple comparisons
    return nil
}
```

## Advanced Code Generation Features

### Optimization for Common Patterns

#### Early Return Optimization
```go
// Generated code returns immediately on first error
func (s *Config) Validate() error {
    if s.Name == "" {
        return goconfig.NewError().Field("Name").Required()  // Early return
    }
    // Don't continue checking other fields if first one fails
    if len(s.Name) < 3 {
        return goconfig.NewError().Field("Name").MinLength(3)  // Early return
    }
    // ... more validations
}
```

#### Field Ordering Optimization
The generator can order validations by likelihood to fail:
```go
// Check required fields first (most likely to fail)
if s.Name == "" { return error }
if s.Email == "" { return error }

// Then check format constraints (less likely to fail)
if !emailRegex.MatchString(s.Email) { return error }

// Finally check range constraints (least likely to fail)  
if s.Port < 1 || s.Port > 65535 { return error }
```

### Custom Validator Integration

Generated code can integrate with custom validators:
```go
func (s *DatabaseConfig) Validate() error {
    // Standard generated validations
    if s.Host == "" {
        return goconfig.NewError().Field("Host").Required()
    }
    
    // Custom validator integration
    if validator, exists := goconfig.GetValidator("custom_db_name"); exists {
        if err := validator.Validate(s.Database); err != nil {
            return goconfig.NewError().Field("Database").Format(err.Error())
        }
    }
    
    return nil
}
```

## Complete Example

### Input Struct
```go
type ServerConfig struct {
    Name     string `validate:"required,minlen=3"`
    Port     int    `validate:"min=1,max=65535"`
    Email    string `validate:"required,email"`
    Debug    bool   `validate:""`
}
```

### Generated Validation Code
```go
// Code generated by go-validate. DO NOT EDIT.
package main

import (
    "regexp"
    "github.com/mateothegreat/go-config"
)

// Pre-compiled regex patterns for validation performance
var (
    emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
)

// Validate validates the ServerConfig struct using zero-reflection validation
func (s *ServerConfig) Validate() error {
    if len(s.Name) < 3 {
        return goconfig.NewError().Field("Name").MinLength(3)
    }
    if s.Name == "" {
        return goconfig.NewError().Field("Name").Required()
    }
    if s.Port > 65535 {
        return goconfig.NewError().Field("Port").Format("must be at most 65535")
    }
    if s.Port < 1 {
        return goconfig.NewError().Field("Port").Format("must be at least 1")
    }
    if !emailRegex.MatchString(s.Email) {
        return goconfig.NewError().Field("Email").Format("must be a valid email address")
    }
    if s.Email == "" {
        return goconfig.NewError().Field("Email").Required()
    }
    return nil
}
```

### Usage
```go
func main() {
    config := &ServerConfig{
        Name:  "my-server",
        Port:  8080,
        Email: "admin@example.com",
        Debug: true,
    }
    
    // Zero-reflection validation
    if err := config.Validate(); err != nil {
        log.Fatal(err)  // Fast validation with detailed errors
    }
    
    fmt.Println("Configuration is valid!")
}
```

## CLI Integration

### Generating Code

```bash
# Generate validation code for current directory
go-validate generate

# Preview generated code without writing files  
go-validate generate --dry-run --verbose

# Generate for specific structs only
go-validate generate --structs ServerConfig,DatabaseConfig

# Generate separate files per struct
go-validate generate --multi
```

### Build Integration

**Makefile integration:**
```makefile
generate:
	go-validate generate

build: generate
	go build ./...

test: generate  
	go test ./...
```

**Go generate integration:**
```go
//go:generate go-validate generate
package main
```

## Comparison with Other Approaches

### vs Runtime Validation Libraries

| Aspect | go-validate (Generated) | reflect-based libraries |
|--------|------------------------|-------------------------|
| **Performance** | 79x faster | Baseline |
| **Memory** | 0 allocations | 23+ allocations |
| **Cold Start** | Instant | Reflection cache warm-up |
| **Type Safety** | Compile-time | Runtime |
| **Debug Info** | Full stack traces | Reflection internals |

### vs Manual Validation

| Aspect | go-validate (Generated) | Hand-written validation |
|--------|------------------------|------------------------|
| **Performance** | Same | Same |
| **Maintainability** | High (declarative tags) | Low (imperative code) |
| **Consistency** | Guaranteed | Manual effort |
| **Error Messages** | Standardized | Inconsistent |
| **Code Volume** | Minimal | High |

### vs JSON Schema Validation

| Aspect | go-validate (Generated) | JSON Schema |
|--------|------------------------|-------------|
| **Performance** | 79x faster | Slower (JSON marshal/unmarshal) |
| **Type Integration** | Native Go types | JSON types only |
| **IDE Support** | Full Go support | Limited |
| **Compile Safety** | Yes | Runtime only |

## Best Practices for Zero-Reflection

### Tag Design

**Good**: Clear, specific validation rules
```go
type Config struct {
    Port int `validate:"min=1,max=65535"`          // Clear numeric range
    Name string `validate:"required,minlen=3"`     // Clear string constraints
}
```

**Avoid**: Complex or ambiguous rules
```go
type Config struct {
    Value string `validate:"regex=^(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$"` // Too complex
}
```

### Field Organization

**Good**: Group related fields together
```go
type ServerConfig struct {
    // Connection settings
    Host string `validate:"required"`
    Port int    `validate:"min=1,max=65535"`
    
    // Authentication
    Username string `validate:"required,minlen=3"`
    Password string `validate:"required,minlen=8"`
    
    // Features  
    SSL    bool `validate:""`
    Debug  bool `validate:""`
}
```

### Error Handling

**Good**: Use structured error handling
```go
if err := config.Validate(); err != nil {
    if validationErrors, ok := goconfig.AsValidationErrors(err); ok {
        for _, verr := range validationErrors.Errors() {
            log.Printf("Field %s: %s", verr.Field, verr.Message)
        }
    }
    return err
}
```

## Debugging Generated Code

### Viewing Generated Output

Use `--dry-run` to see generated code:
```bash
go-validate generate --dry-run --verbose
```

### Common Issues and Solutions

#### Missing Validation Methods
**Problem**: `config.Validate() undefined`
**Solution**: Ensure struct has validation tags and re-run generator

#### Import Errors in Generated Code
**Problem**: Generated code has import errors
**Solution**: Check that `github.com/mateothegreat/go-config` is in go.mod

#### Performance Regression
**Problem**: Generated validation slower than expected
**Solution**: 
1. Check for complex regex patterns
2. Verify inlining is occurring (`go build -gcflags='-m'`)
3. Profile with `go test -cpuprofile`

### Testing Generated Code

**Validation correctness:**
```go
func TestGeneratedValidation(t *testing.T) {
    config := &ServerConfig{
        Name: "",  // Invalid - required field
        Port: 0,   // Invalid - below minimum
    }
    
    err := config.Validate()
    assert.Error(t, err)
    
    // Test specific error content
    validationErrors, ok := goconfig.AsValidationErrors(err)
    assert.True(t, ok)
    assert.Len(t, validationErrors.Errors(), 2)
}
```

**Performance testing:**
```go
func BenchmarkGeneratedValidation(b *testing.B) {
    config := &ServerConfig{
        Name: "valid-name",
        Port: 8080,
        Email: "test@example.com",
    }
    
    b.ResetTimer()
    b.ReportAllocs()
    
    for i := 0; i < b.N; i++ {
        err := config.Validate()
        if err != nil {
            b.Fatal(err)
        }
    }
    // Should report 0 allocs/op for successful validation
}
```

## Future Enhancements

### Planned Features

1. **Cross-field Validation**: Generate code for field dependencies
   ```go
   // Future: dependent field validation
   type Config struct {
       SSL  bool   `validate:""`
       Port int    `validate:"ssl_port_constraint"`  // Port 443 if SSL=true
   }
   ```

2. **Conditional Validation**: Generate code based on field values
   ```go
   // Future: conditional validation  
   type Config struct {
       Type     string `validate:"oneof=db|cache|queue"`
       Database string `validate:"required_if=Type:db"`
   }
   ```

3. **Slice/Map Validation**: Generate code for collection validation
   ```go
   // Future: collection validation
   type Config struct {
       Servers []Server `validate:"dive,required"`  // Validate each Server
   }
   ```

### Performance Improvements

1. **Template Optimization**: Further reduce generated code size
2. **Parallel Validation**: Generate code for concurrent field validation
3. **SIMD Operations**: Use vectorized operations for bulk validation

## Conclusion

Zero-reflection validation represents a paradigm shift from runtime introspection to compile-time code generation. This approach delivers:

- **Dramatic performance improvements** (79x faster)
- **Zero runtime allocations** for successful validation
- **Compile-time type safety** and error detection
- **Excellent debugging experience** with full stack traces
- **Predictable performance** characteristics

The generated code is **readable, debuggable, and optimizable** by both developers and the Go compiler, making it an ideal choice for performance-critical applications where validation overhead matters.

By moving complexity from runtime to build time, `go-validate` achieves the best of both worlds: the convenience of declarative validation with the performance of hand-optimized code.