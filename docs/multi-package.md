# Multi-Package Validation Example

This example demonstrates validation code generation across multiple packages in a large codebase, showcasing organization patterns for enterprise applications.

## 📋 What You'll Learn

- Multi-package validation architecture
- Cross-package validation dependencies
- Separate validation packages
- Large codebase organization patterns
- Package-specific generation strategies

## 🏗️ Project Structure

```
09-multi-package/
├── cmd/
│   └── server/              # Main application
│       ├── main.go
│       └── config.go
├── internal/
│   ├── config/              # Application configuration
│   │   ├── app.go
│   │   └── app_validation_generated.go
│   ├── database/            # Database configuration
│   │   ├── config.go
│   │   └── config_validation_generated.go
│   └── api/                 # API configuration
│       ├── config.go
│       └── config_validation_generated.go
├── pkg/
│   └── validation/          # Shared validation utilities
│       └── utils.go
├── go.mod
├── Makefile                 # Multi-package build automation
└── README.md
```

## 🚀 Quick Start

1. **Generate validation for all packages:**

   ```bash
   make generate-all
   ```

2. **Generate for specific package:**

   ```bash
   make generate-config
   make generate-database
   make generate-api
   ```

3. **Run the application:**

   ```bash
   go run ./cmd/server
   ```

## 📖 Package-Specific Generation

### 1. Application Configuration Package

```bash
# Generate validation for internal/config package
go-validate generate \
    --input-dir ./internal/config \
    --output-dir ./internal/config \
    --package config \
    --verbose
```

### 2. Database Configuration Package

```bash
# Generate validation for internal/database package
go-validate generate \
    --input-dir ./internal/database \
    --output-dir ./internal/database \
    --package database \
    --verbose
```

### 3. API Configuration Package

```bash
# Generate validation for internal/api package
go-validate generate \
    --input-dir ./internal/api \
    --output-dir ./internal/api \
    --package api \
    --verbose
```

### 4. Centralized Validation Package

```bash
# Generate all validation to shared package
go-validate generate \
    --input-dir . \
    --output-dir ./pkg/validation \
    --package validation \
    --multi \
    --verbose
```

## 🔧 Package Organization Patterns

### Pattern 1: In-Package Validation (Recommended)

Each package contains its own validation code:

```go
// internal/config/app.go
package config

type AppConfig struct {
    Name string `validate:"required,minlen=3"`
}

// internal/config/app_validation_generated.go
package config

func (a *AppConfig) Validate() error {
    // Generated validation code
}
```

### Pattern 2: Centralized Validation Package

All validation code in a shared package:

```go
// pkg/validation/app_validation_generated.go
package validation

import "myapp/internal/config"

func ValidateAppConfig(cfg *config.AppConfig) error {
    // Generated validation code
}
```

### Pattern 3: Hybrid Approach

Core validation in packages, shared utilities centralized:

```go
// internal/config/app.go - contains Validate() method
// pkg/validation/utils.go - contains shared validation utilities
```

## 🎯 Advanced Multi-Package Scenarios

### 1. Cross-Package Dependencies

```go
// internal/config/app.go
package config

import "myapp/internal/database"

type AppConfig struct {
    Name     string `validate:"required"`
    Database database.Config // Embedded from another package
}

func (a *AppConfig) ValidateComplete() error {
    // Validate this package
    if err := a.Validate(); err != nil {
        return err
    }
    
    // Validate embedded config from another package
    if err := a.Database.Validate(); err != nil {
        return fmt.Errorf("database config: %w", err)
    }
    
    return nil
}
```

### 2. Interface-Based Validation

```go
// pkg/validation/interfaces.go
package validation

type Validator interface {
    Validate() error
}

type ConfigProvider interface {
    Validator
    GetEnvironment() string
}

// Validate any config that implements the interface
func ValidateConfig(cfg Validator) error {
    return cfg.Validate()
}
```

### 3. Package-Specific Custom Validators

```go
// internal/database/validators.go
package database

import "github.com/mateothegreat/go-config"

func init() {
    // Register database-specific validators
    goconfig.RegisterValidator("connection_string", &ConnectionStringValidator{})
    goconfig.RegisterValidator("table_name", &TableNameValidator{})
}
```

## 🔄 Build Automation Strategies

### Makefile for Multi-Package Generation

```makefile
# Generate validation for all packages
generate-all:
 @echo "🔄 Generating validation for all packages..."
 $(MAKE) generate-config
 $(MAKE) generate-database
 $(MAKE) generate-api
 @echo "✅ All packages generated"

# Generate for specific packages
generate-config:
 go-validate generate --input-dir ./internal/config --package config

generate-database:
 go-validate generate --input-dir ./internal/database --package database

generate-api:
 go-validate generate --input-dir ./internal/api --package api

# Generate to centralized validation package
generate-centralized:
 go-validate generate \
  --output-dir ./pkg/validation \
  --package validation \
  --multi
```

### Go Generate Directives per Package

```go
// internal/config/app.go
//go:generate go-validate generate --input-dir . --package config

// internal/database/config.go
//go:generate go-validate generate --input-dir . --package database

// internal/api/config.go
//go:generate go-validate generate --input-dir . --package api
```

### Recursive Generation Script

```bash
#!/bin/bash
# generate-all.sh

set -e

echo "🔄 Multi-package validation generation"

# Find all packages with validation tags
PACKAGES=$(find . -name "*.go" -exec grep -l "validate:" {} \; | \
           xargs -I {} dirname {} | \
           sort -u | \
           grep -v "vendor\|\.git")

for package in $PACKAGES; do
    echo "📦 Generating validation for package: $package"
    go-validate generate \
        --input-dir "$package" \
        --output-dir "$package" \
        --package "$(basename "$package")" \
        --verbose
done

echo "✅ Multi-package generation complete"
```

## 🧪 Testing Multi-Package Validation

### Package-Specific Tests

```go
// internal/config/app_test.go
package config

func TestAppConfig_Validate(t *testing.T) {
    cfg := &AppConfig{
        Name: "test-app",
    }
    
    err := cfg.Validate()
    assert.NoError(t, err)
}
```

### Integration Tests

```go
// cmd/server/integration_test.go
package main

func TestCompleteConfigValidation(t *testing.T) {
    // Load configuration from all packages
    cfg := LoadConfig()
    
    // Validate all components
    assert.NoError(t, cfg.App.Validate())
    assert.NoError(t, cfg.Database.Validate())
    assert.NoError(t, cfg.API.Validate())
}
```

### Cross-Package Validation Tests

```go
// pkg/validation/integration_test.go
package validation

func TestCrossPackageValidation(t *testing.T) {
    appCfg := &config.AppConfig{/* ... */}
    dbCfg := &database.Config{/* ... */}
    
    // Test that configs work together
    assert.NoError(t, ValidateConfig(appCfg))
    assert.NoError(t, ValidateConfig(dbCfg))
}
```

## 📊 Performance Considerations

### Package Loading Optimization

```go
// Lazy loading of validation packages
var (
    configValidationOnce sync.Once
    dbValidationOnce     sync.Once
    apiValidationOnce    sync.Once
)

func ValidateConfig(cfg interface{}) error {
    switch c := cfg.(type) {
    case *config.AppConfig:
        configValidationOnce.Do(func() {
            // Initialize config validation
        })
        return c.Validate()
    }
}
```

### Parallel Validation

```go
func ValidateAllConfigs(configs ...Validator) error {
    var wg sync.WaitGroup
    errors := make(chan error, len(configs))
    
    for _, cfg := range configs {
        wg.Add(1)
        go func(c Validator) {
            defer wg.Done()
            if err := c.Validate(); err != nil {
                errors <- err
            }
        }(cfg)
    }
    
    wg.Wait()
    close(errors)
    
    // Collect errors
    var validationErrors []error
    for err := range errors {
        validationErrors = append(validationErrors, err)
    }
    
    if len(validationErrors) > 0 {
        return fmt.Errorf("validation failed: %v", validationErrors)
    }
    
    return nil
}
```

## 🔧 CI/CD for Multi-Package Projects

### GitHub Actions Workflow

```yaml
name: Multi-Package Validation

on: [push, pull_request]

jobs:
  validate:
    runs-on: ubuntu-latest
    strategy:
      matrix:
        package: [config, database, api]
    
    steps:
    - uses: actions/checkout@v3
    
    - name: Setup Go
      uses: actions/setup-go@v3
      with:
        go-version: '1.21'
    
    - name: Generate validation for ${{ matrix.package }}
      run: |
        go install github.com/mateothegreat/go-config/cmd/go-validate@latest
        go-validate generate --input-dir ./internal/${{ matrix.package }}
    
    - name: Test ${{ matrix.package }}
      run: go test ./internal/${{ matrix.package }}
```

### Monorepo Validation Strategy

```bash
# .github/workflows/validate-monorepo.sh
#!/bin/bash

# Find changed packages
CHANGED_PACKAGES=$(git diff --name-only HEAD~1 | \
                   grep "\.go$" | \
                   xargs -I {} dirname {} | \
                   sort -u)

# Generate validation only for changed packages
for package in $CHANGED_PACKAGES; do
    if [ -f "$package"/*.go ]; then
        echo "Validating changed package: $package"
        go-validate generate --input-dir "$package"
        go test "$package"
    fi
done
```

## 🎓 Best Practices for Multi-Package

### 1. Package Boundaries

- **Keep validation close to data** - Generate in the same package as structs
- **Minimize cross-package dependencies** - Avoid complex validation chains
- **Use interfaces for abstraction** - Enable testing and flexibility

### 2. Code Organization

- **Consistent naming** - Use `*_validation_generated.go` pattern
- **Clear package responsibilities** - Each package owns its validation
- **Shared utilities** - Common validation helpers in shared package

### 3. Build Process

- **Automate generation** - Use Makefile or scripts for consistency
- **Verify in CI** - Ensure generated code is up to date
- **Test thoroughly** - Include integration tests for cross-package scenarios

## 🎓 Next Steps

- Learn error handling patterns in [`../10-error-handling/`](../examples/10-error-handling)
- See production integration in [`../11-ci-cd-integration/`](../examples/11-ci-cd-integration)
- Explore performance testing in [`../13-performance-testing/`](../examples/13-performance-testing)
