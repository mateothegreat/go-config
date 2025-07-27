# Advanced Error Handling Example

This example demonstrates sophisticated error handling patterns with `go-validate`, including structured errors, error aggregation, and user-friendly error reporting.

## 📋 What You'll Learn

- Fluent error composition patterns
- Structured validation error handling
- Error aggregation and reporting
- User-friendly error messages
- Error context and metadata
- Recovery and fallback strategies

## 🏗️ Files

- `config.go` - Configuration structs with validation
- `errors.go` - Custom error handling utilities
- `main.go` - Comprehensive error handling examples
- `reporting.go` - Error reporting and formatting
- `recovery.go` - Error recovery strategies

## 🚀 Quick Start

1. **Generate validation code:**

   ```bash
   go-validate generate
   ```

2. **Run error handling examples:**

   ```bash
   go run .
   ```

3. **Run with invalid config:**

   ```bash
   cp config-invalid.yaml config.yaml
   go run .
   ```

## 📖 Error Handling Patterns

### 1. Fluent Error Composition

```go
// Single field error with context
return goconfig.NewError().
    Field("Email").
    Value(config.Email).
    Tag("email").
    Code("INVALID_EMAIL_FORMAT").
    Format("must be a valid email address")

// Multiple field errors in one call
return goconfig.NewError().
    Field("Email").Format("invalid email format").
    Field("Port").Format("must be between 1 and 65535").
    Field("Name").Required()
```

### 2. Structured Error Inspection

```go
if err := config.Validate(); err != nil {
    // Check if it's a validation error
    if validationErrors, ok := goconfig.AsValidationErrors(err); ok {
        for _, verr := range validationErrors.Errors() {
            fmt.Printf("Field: %s\n", verr.Field)
            fmt.Printf("Value: %v\n", verr.Value)
            fmt.Printf("Tag: %s\n", verr.Tag)
            fmt.Printf("Message: %s\n", verr.Message)
            fmt.Printf("Code: %s\n", verr.Code)
        }
    }
}
```

### 3. Error Aggregation

```go
type ConfigValidator struct {
    errors []error
}

func (v *ConfigValidator) AddError(err error) {
    v.errors = append(v.errors, err)
}

func (v *ConfigValidator) HasErrors() bool {
    return len(v.errors) > 0
}

func (v *ConfigValidator) Error() error {
    if !v.HasErrors() {
        return nil
    }
    return goconfig.NewMultiError(v.errors)
}
```

### 4. Contextual Error Messages

```go
func validateWithContext(config *Config, environment string) error {
    if err := config.Validate(); err != nil {
        return fmt.Errorf("validation failed for %s environment: %w", 
            environment, err)
    }
    return nil
}
```

## 🎯 Advanced Error Patterns

### 1. Error Severity Levels

```go
type ErrorSeverity int

const (
    ErrorSeverityWarning ErrorSeverity = iota
    ErrorSeverityError
    ErrorSeverityCritical
)

type ValidationError struct {
    Field    string
    Message  string
    Severity ErrorSeverity
    Code     string
}

func (e *ValidationError) IsWarning() bool {
    return e.Severity == ErrorSeverityWarning
}

func (e *ValidationError) IsCritical() bool {
    return e.Severity == ErrorSeverityCritical
}
```

### 2. Error Recovery Strategies

```go
type ConfigWithFallback struct {
    Primary  *Config
    Fallback *Config
}

func (c *ConfigWithFallback) ValidateWithRecovery() error {
    // Try primary config first
    if err := c.Primary.Validate(); err == nil {
        return nil
    }
    
    // Fall back to secondary config
    if c.Fallback != nil {
        if err := c.Fallback.Validate(); err == nil {
            log.Warn("Using fallback configuration due to primary validation failure")
            *c.Primary = *c.Fallback
            return nil
        }
    }
    
    // Both failed, return structured error
    return goconfig.NewError().
        Field("configuration").
        Format("both primary and fallback configurations are invalid")
}
```

### 3. Progressive Validation

```go
func ValidateProgressive(config *Config) error {
    validator := &ProgressiveValidator{}
    
    // Critical validations first (fail fast)
    validator.ValidateCritical(config)
    if validator.HasCriticalErrors() {
        return validator.CriticalError()
    }
    
    // Non-critical validations
    validator.ValidateWarnings(config)
    
    if validator.HasWarnings() {
        log.Warn("Configuration has warnings", validator.Warnings())
    }
    
    return validator.Error()
}
```

### 4. Error Localization

```go
type ErrorLocalizer struct {
    locale   string
    messages map[string]map[string]string
}

func (l *ErrorLocalizer) LocalizeError(err *goconfig.ValidationError) string {
    if messages, ok := l.messages[l.locale]; ok {
        if msg, exists := messages[err.Code]; exists {
            return fmt.Sprintf(msg, err.Field, err.Value)
        }
    }
    return err.Message // Fallback to default
}

// Usage
localizer := &ErrorLocalizer{
    locale: "es",
    messages: map[string]map[string]string{
        "es": {
            "REQUIRED": "El campo '%s' es obligatorio",
            "INVALID_EMAIL": "El campo '%s' debe ser un email válido",
        },
    },
}
```

## 🔧 Error Reporting Utilities

### 1. Human-Readable Error Reports

```go
func FormatValidationReport(err error) string {
    var report strings.Builder
    
    report.WriteString("Configuration Validation Report\n")
    report.WriteString("================================\n\n")
    
    if validationErrors, ok := goconfig.AsValidationErrors(err); ok {
        errorCount := len(validationErrors.Errors())
        report.WriteString(fmt.Sprintf("Found %d validation error(s):\n\n", errorCount))
        
        for i, verr := range validationErrors.Errors() {
            report.WriteString(fmt.Sprintf("%d. Field: %s\n", i+1, verr.Field))
            report.WriteString(fmt.Sprintf("   Error: %s\n", verr.Message))
            if verr.Value != nil {
                report.WriteString(fmt.Sprintf("   Current Value: %v\n", verr.Value))
            }
            if verr.Tag != "" {
                report.WriteString(fmt.Sprintf("   Validation Rule: %s\n", verr.Tag))
            }
            report.WriteString("\n")
        }
    }
    
    return report.String()
}
```

### 2. JSON Error Responses

```go
type APIErrorResponse struct {
    Success bool                    `json:"success"`
    Message string                  `json:"message"`
    Errors  []APIValidationError    `json:"errors,omitempty"`
    Code    string                  `json:"code"`
}

type APIValidationError struct {
    Field   string      `json:"field"`
    Message string      `json:"message"`
    Value   interface{} `json:"value,omitempty"`
    Code    string      `json:"code"`
}

func FormatAPIError(err error) *APIErrorResponse {
    response := &APIErrorResponse{
        Success: false,
        Message: "Validation failed",
        Code:    "VALIDATION_ERROR",
    }
    
    if validationErrors, ok := goconfig.AsValidationErrors(err); ok {
        for _, verr := range validationErrors.Errors() {
            response.Errors = append(response.Errors, APIValidationError{
                Field:   verr.Field,
                Message: verr.Message,
                Value:   verr.Value,
                Code:    verr.Code,
            })
        }
    }
    
    return response
}
```

### 3. Error Metrics and Monitoring

```go
type ErrorMetrics struct {
    TotalErrors   int64
    FieldErrors   map[string]int64
    ErrorCodes    map[string]int64
    mu            sync.RWMutex
}

func (m *ErrorMetrics) RecordError(err *goconfig.ValidationError) {
    m.mu.Lock()
    defer m.mu.Unlock()
    
    m.TotalErrors++
    
    if m.FieldErrors == nil {
        m.FieldErrors = make(map[string]int64)
    }
    m.FieldErrors[err.Field]++
    
    if m.ErrorCodes == nil {
        m.ErrorCodes = make(map[string]int64)
    }
    m.ErrorCodes[err.Code]++
}

func (m *ErrorMetrics) GetTopErrors(limit int) []ErrorStat {
    m.mu.RLock()
    defer m.mu.RUnlock()
    
    var stats []ErrorStat
    for field, count := range m.FieldErrors {
        stats = append(stats, ErrorStat{Field: field, Count: count})
    }
    
    sort.Slice(stats, func(i, j int) bool {
        return stats[i].Count > stats[j].Count
    })
    
    if len(stats) > limit {
        stats = stats[:limit]
    }
    
    return stats
}
```

## 🧪 Testing Error Scenarios

### 1. Error Testing Patterns

```go
func TestConfigValidation_Errors(t *testing.T) {
    tests := []struct {
        name          string
        config        Config
        expectedFields []string
        expectedCodes  []string
    }{
        {
            name: "missing required fields",
            config: Config{
                // Name missing (required)
                Port: 8080,
            },
            expectedFields: []string{"Name"},
            expectedCodes:  []string{"REQUIRED"},
        },
        {
            name: "invalid email and port",
            config: Config{
                Name:  "test",
                Email: "invalid-email",
                Port:  70000, // exceeds max
            },
            expectedFields: []string{"Email", "Port"},
            expectedCodes:  []string{"INVALID_EMAIL", "MAX_VALUE_EXCEEDED"},
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := tt.config.Validate()
            require.Error(t, err)
            
            validationErrors, ok := goconfig.AsValidationErrors(err)
            require.True(t, ok)
            
            // Check expected fields
            actualFields := make([]string, len(validationErrors.Errors()))
            for i, verr := range validationErrors.Errors() {
                actualFields[i] = verr.Field
            }
            assert.ElementsMatch(t, tt.expectedFields, actualFields)
            
            // Check expected error codes
            actualCodes := make([]string, len(validationErrors.Errors()))
            for i, verr := range validationErrors.Errors() {
                actualCodes[i] = verr.Code
            }
            assert.ElementsMatch(t, tt.expectedCodes, actualCodes)
        })
    }
}
```

### 2. Error Recovery Testing

```go
func TestErrorRecovery(t *testing.T) {
    primary := &Config{
        Name: "", // Invalid
        Port: 8080,
    }
    
    fallback := &Config{
        Name: "fallback-server",
        Port: 9090,
    }
    
    configWithFallback := &ConfigWithFallback{
        Primary:  primary,
        Fallback: fallback,
    }
    
    err := configWithFallback.ValidateWithRecovery()
    assert.NoError(t, err)
    assert.Equal(t, "fallback-server", primary.Name)
}
```

## 🎯 Production Error Handling

### 1. Graceful Degradation

```go
func StartServerWithValidation(config *Config) error {
    // Validate critical configuration
    if err := validateCritical(config); err != nil {
        return fmt.Errorf("critical configuration error: %w", err)
    }
    
    // Validate non-critical configuration (warn only)
    if err := validateNonCritical(config); err != nil {
        log.Warn("Non-critical configuration issues detected",
            "errors", FormatValidationReport(err))
    }
    
    return startServer(config)
}
```

### 2. Configuration Health Checks

```go
func HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
    config := getCurrentConfig()
    
    if err := config.Validate(); err != nil {
        w.WriteHeader(http.StatusServiceUnavailable)
        json.NewEncoder(w).Encode(map[string]interface{}{
            "status": "unhealthy",
            "reason": "configuration_invalid",
            "errors": FormatAPIError(err),
        })
        return
    }
    
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]string{
        "status": "healthy",
    })
}
```

## 🎓 Best Practices

### 1. Error Message Guidelines

- **Be specific** - "Port must be between 1 and 65535" vs "Invalid port"
- **Include context** - Show current value when helpful
- **Provide guidance** - Suggest valid alternatives when possible
- **Use consistent terminology** - Standardize error message patterns

### 2. Error Handling Strategy

- **Fail fast for critical errors** - Stop validation on security issues
- **Collect all validation errors** - Show all problems at once when possible
- **Provide recovery options** - Offer fallback configurations
- **Log appropriately** - Different log levels for different error types

### 3. User Experience

- **Progressive disclosure** - Show critical errors first
- **Clear action items** - Tell users exactly what to fix
- **Helpful documentation** - Link to configuration guides
- **Consistent formatting** - Use standard error report formats

## 🎓 Next Steps

- Explore CI/CD integration in [`../11-ci-cd-integration/`](../examples/11-ci-cd-integration)
- See performance testing in [`../13-performance-testing/`](../examples/13-performance-testing)
- Learn production patterns in [`../04-microservice-config/`](../examples/04-microservice-config)
