# Basic Validation Example

This example demonstrates the fundamental usage of `go-validate` with a simple configuration struct.

## 📋 What You'll Learn

- Basic validation tags and rules
- Manual validation code generation
- Error handling patterns
- Performance benefits of zero-reflection validation

## 🏗️ Files

- `config.go` - Configuration struct with validation tags
- `main.go` - Example usage and error handling
- `config.yaml` - Sample configuration file
- `validation_generated.go` - Generated validation code (after running CLI)

## 🚀 Quick Start

1. **Generate validation code:**
   ```bash
   go-validate generate
   ```

2. **Run the example:**
   ```bash
   go run .
   ```

3. **Try with invalid data:**
   ```bash
   # Edit config.yaml with invalid values and run again
   go run .
   ```

## 📖 Code Walkthrough

### Configuration Struct

The `ServerConfig` struct demonstrates common validation patterns:

```go
type ServerConfig struct {
    // Required string with minimum length
    Name string `validate:"required,minlen=3"`
    
    // Numeric range validation
    Port int `validate:"min=1,max=65535"`
    
    // Email format validation
    Email string `validate:"required,email"`
    
    // Enum validation
    LogLevel string `validate:"oneof=debug|info|warn|error"`
}
```

### Generated Validation

The CLI generates zero-reflection validation code:

```go
func (s *ServerConfig) Validate() error {
    if s.Name == "" {
        return errors.NewError().Field("Name").Required()
    }
    if len(s.Name) < 3 {
        return errors.NewError().Field("Name").MinLength(3)
    }
    // ... more validations
    return nil
}
```

### Usage Pattern

```go
config := &ServerConfig{/* loaded from file/env */}

if err := config.Validate(); err != nil {
    // Handle validation errors
    log.Fatal(err)
}

// Config is guaranteed valid
```

## 🎯 Key Benefits Demonstrated

1. **Zero Reflection**: Direct field access, no runtime introspection
2. **Zero Allocations**: No memory allocations for successful validation
3. **Compile-time Safety**: Generated code is checked by Go compiler
4. **Fast Execution**: 79x faster than reflection-based validation

## 🔄 Try Different Scenarios

### Valid Configuration
```yaml
name: "my-server"
port: 8080
email: "admin@example.com"
log_level: "info"
```

### Invalid Port
```yaml
name: "my-server"
port: 70000  # Invalid: exceeds maximum
email: "admin@example.com"
log_level: "info"
```

### Missing Required Field
```yaml
# name: "my-server"  # Missing required field
port: 8080
email: "admin@example.com"
log_level: "info"
```

### Invalid Email
```yaml
name: "my-server"
port: 8080
email: "not-an-email"  # Invalid format
log_level: "info"
```

## 📊 Performance Comparison

Run the included benchmark to see the performance difference:

```bash
go test -bench=. -benchmem
```

Expected results:
```
BenchmarkGeneratedValidation-8    50000000    23.4 ns/op    0 B/op   0 allocs/op
BenchmarkReflectionValidation-8    1000000   1847 ns/op  512 B/op  23 allocs/op
```

## 🎓 Next Steps

- Learn automated generation with [`../02-go-generate/`](../2-embedded-generation)
- Explore real-world usage in [`../04-microservice-config/`](../04-microservice-config)
- See custom validators in [`../07-custom-validators/`](../../docs/07-custom-validators)