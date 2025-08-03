# AST-Based Validation Generation Example

This example demonstrates the high-performance AST-based validation code generation that integrates with the [go-validation](https://github.com/mateothegreat/go-validation) library.

## Features

- **Zero Reflection at Runtime**: All validation code is generated at compile time
- **Full go-validation Integration**: Leverages the validation library's registry and built-in validators
- **AST-Based Generation**: Uses Go's AST package instead of text templates for better performance and type safety
- **Comprehensive Validation Rules**: Supports all go-validation rules including:
  - Basic: required, email, url, alpha, alphanum, numeric
  - Numeric: min, max, range
  - String: minlen, maxlen, len
  - Network: ip, ipv4, ipv6, cidr, mac, hostname
  - Format: uuid, datetime, base64, creditcard, phone
  - Conditional: oneof, required_if, required_unless

## Usage

1. Generate the validation code:
   ```bash
   go generate
   ```

2. Run the example:
   ```bash
   go run .
   ```

## Generated Code

The AST generator creates optimized validation methods that:

1. Register validators with go-validation's registry
2. Create type-specific validation methods
3. Use the fluent error builder for clear error messages
4. Avoid any runtime reflection

Example generated validation:

```go
// init registers generated validators with the go-validation registry
func init() {
    validation.RegisterValidation("serverconfig", ValidateServerConfig)
    validation.RegisterValidation("databaseconfig", ValidateDatabaseConfig)
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
    // ... more validations
    
    if errs.HasErrors() {
        return errs
    }
    return nil
}
```

## Performance

The AST-based approach provides several performance benefits:

1. **Compile-Time Generation**: No runtime parsing or reflection
2. **Direct Function Calls**: Validators are called directly, not through reflection
3. **Cached Expressions**: Validator expressions are cached during generation
4. **Minimal Allocations**: Error objects are only created when validation fails

## Customization

You can extend the validation system by:

1. Adding custom validators to the go-validation registry
2. Extending the ValidatorRegistry with new rules
3. Creating specialized AST generators for specific use cases 