# Custom Validators Example

This example demonstrates how to create, register, and use custom validation plugins with the `go-validate` framework.

## 📋 What You'll Learn

- Creating custom validator plugins
- Plugin registration and metadata
- Integration with generated validation code
- Built-in custom validators usage
- Advanced validation patterns

## 🏗️ Files

- `validators/` - Custom validator implementations
- `config.go` - Configuration using custom validators
- `main.go` - Registration and usage example
- `validators_test.go` - Testing custom validators

## 🚀 Quick Start

1. **Generate validation code:**

   ```bash
   go-validate generate
   ```

2. **Run the example:**

   ```bash
   go run .
   ```

## 📖 Custom Validator Examples

### 1. Domain Name Validator

```go
type DomainValidator struct{}

func (v *DomainValidator) Validate(field interface{}) error {
    domain, ok := field.(string)
    if !ok {
        return fmt.Errorf("domain validation requires string input")
    }
    
    if !isValidDomain(domain) {
        return fmt.Errorf("invalid domain format")
    }
    
    return nil
}
```

### 2. Credit Card Validator (Luhn Algorithm)

```go
type CreditCardValidator struct{}

func (v *CreditCardValidator) Validate(field interface{}) error {
    cardNumber, ok := field.(string)
    if !ok {
        return fmt.Errorf("credit card validation requires string")
    }
    
    if !luhnCheck(cardNumber) {
        return fmt.Errorf("invalid credit card number")
    }
    
    return nil
}
```

### 3. Business Rules Validator

```go
type BusinessHoursValidator struct{}

func (v *BusinessHoursValidator) Validate(field interface{}) error {
    hours, ok := field.(string)
    if !ok {
        return fmt.Errorf("business hours validation requires string")
    }
    
    // Validate format: "09:00-17:00"
    if !businessHoursPattern.MatchString(hours) {
        return fmt.Errorf("business hours must be in format HH:MM-HH:MM")
    }
    
    return nil
}
```

## 🔧 Plugin Registration

### Simple Registration

```go
func init() {
    goconfig.RegisterValidator("domain", &DomainValidator{})
    goconfig.RegisterValidator("creditcard", &CreditCardValidator{})
    goconfig.RegisterValidator("business_hours", &BusinessHoursValidator{})
}
```

### Registration with Metadata

```go
func init() {
    goconfig.RegisterValidatorWithMetadata("domain", &DomainValidator{}, 
        goconfig.PluginMetadata{
            Name:        "Domain Validator",
            Version:     "1.0.0",
            Description: "Validates domain name format",
            Author:      "Your Team",
            SupportedTypes: []string{"string"},
        })
}
```

## 🎯 Usage in Structs

```go
type CompanyConfig struct {
    // Built-in custom validators
    Website     string `validate:"required,url"`
    Email       string `validate:"required,email"`
    Phone       string `validate:"required,phone"`
    
    // Custom validators
    Domain      string `validate:"required,domain"`
    CreditCard  string `validate:"creditcard"`
    Hours       string `validate:"business_hours"`
    
    // Combination of validators
    AdminEmail  string `validate:"required,email,domain"`
}
```

## 🔍 Built-in Custom Validators

The framework includes several built-in custom validators:

### IP Address Validator

```go
IPAddress string `validate:"ip"`  // IPv4 or IPv6
```

### UUID Validator

```go
UserID string `validate:"uuid"`  // RFC 4122 format
```

### Phone Number Validator

```go
ContactPhone string `validate:"phone"`  // International format
```

### Credit Card Validator

```go
PaymentCard string `validate:"creditcard"`  // Luhn algorithm
```

## 🧪 Testing Custom Validators

```go
func TestDomainValidator(t *testing.T) {
    validator := &DomainValidator{}
    
    tests := []struct {
        name    string
        input   string
        wantErr bool
    }{
        {
            name:    "valid domain",
            input:   "example.com",
            wantErr: false,
        },
        {
            name:    "invalid domain",
            input:   "not..a..domain",
            wantErr: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := validator.Validate(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

## 🔄 Plugin Management

### Listing Registered Validators

```go
validators := goconfig.ListValidators()
for name, validator := range validators {
    fmt.Printf("Validator: %s\n", name)
    
    if metadata, exists := goconfig.GetValidatorMetadata(name); exists {
        fmt.Printf("  Version: %s\n", metadata.Version)
        fmt.Printf("  Description: %s\n", metadata.Description)
    }
}
```

### Runtime Validation

```go
// Validate with specific plugin
err := goconfig.ValidateWithPlugin("domain", "example.com")
if err != nil {
    log.Printf("Domain validation failed: %v", err)
}
```

## ⚡ Performance Considerations

### Pre-compiled Patterns

Store regex patterns as package-level variables:

```go
var (
    domainPattern = regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(\.[a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$`)
    phonePattern  = regexp.MustCompile(`^\+?[1-9]\d{1,14}$`)
)
```

### Efficient Implementation

```go
func (v *DomainValidator) Validate(field interface{}) error {
    domain := field.(string) // Type assertion (fast)
    
    // Quick length check first
    if len(domain) == 0 || len(domain) > 253 {
        return fmt.Errorf("invalid domain length")
    }
    
    // Then regex check
    if !domainPattern.MatchString(domain) {
        return fmt.Errorf("invalid domain format")
    }
    
    return nil
}
```

## 🎓 Advanced Patterns

### Contextual Validation

```go
type ContextualValidator struct {
    context map[string]interface{}
}

func (v *ContextualValidator) Validate(field interface{}) error {
    // Access other configuration values for validation
    if serverType := v.context["server_type"]; serverType == "production" {
        // Stricter validation for production
    }
    return nil
}
```

### Chain Validation

```go
type ChainValidator struct {
    validators []goconfig.ValidatorPlugin
}

func (v *ChainValidator) Validate(field interface{}) error {
    for _, validator := range v.validators {
        if err := validator.Validate(field); err != nil {
            return err
        }
    }
    return nil
}
```

## 🎓 Next Steps

- Explore embedding patterns in [`../08-embedding-composition/`](../examples/08-embedding-composition)
- See multi-package examples in [`../09-multi-package/`](../examples/09-multi-package)
- Learn error handling in [`../10-error-handling/`](../examples/10-error-handling)
