# go-validate

**Fast, fluent, and flexible validation for Go configs.**

`go-validate` is a powerful code generation tool built on three foundational pillars:

- ⚡ **Performance**: Generates zero-reflection, allocation-conscious code
- ✨ **Ergonomics**: Offers intuitive interfaces and seamless config integration  
- 🔗 **Fluency**: Enables chainable APIs for expressive validation and error handling

Unlike typical reflection-based libraries, `go-validate` generates optimized validation code directly from your Go structs—ensuring lightning-fast execution and cold-start performance.

## 🚀 Quick Start

### Installation

```bash
go install github.com/mateothegreat/go-config/cmd/go-validate@latest
```

### Basic Usage

1. **Define your config struct with validation tags:**

```go
package main

type ServerConfig struct {
    Name     string `validate:"required,minlen=3"`
    Port     int    `validate:"min=1,max=65535"`
    Email    string `validate:"required,email"`
    Debug    bool   `validate:""`
}
```

2. **Generate validation code:**

```bash
go-validate generate
```

3. **Use the generated validation:**

```go
config := &ServerConfig{
    Name:  "my-server",
    Port:  8080,
    Email: "admin@example.com",
    Debug: true,
}

if err := config.Validate(); err != nil {
    log.Fatal(err)
}
```

## 🏗️ Core Features

### ✅ Zero-Reflection Validation

Generates direct method calls like:

```go
func (cfg *MyConfig) Validate() error {
    if cfg.Name == "" {
        return goconfig.NewError().Field("Name").Required()
    }
    if len(cfg.Name) < 3 {
        return goconfig.NewError().Field("Name").MinLength(3)
    }
    return nil
}
```

**Performance Benefits:**
- No reflection overhead
- Zero allocations for successful validation
- Pre-compiled regex patterns
- Direct field access

### 🔍 AST-Based Scanning

- Uses Go's native AST for stable parsing
- Detects tags like `validate:"min=1,max=10"` without fragile regex
- Supports multi-tag fields: `json:"id" validate:"min=1" yaml:"id"`
- Automatically discovers structs with validation tags
- Caches parsed AST for improved performance

### 🔄 Dry Runs & Previews

Preview generated output with `--dry-run`:

```bash
go-validate generate --dry-run --verbose
```

Perfect for:
- Debugging code generation
- CI pipeline validation
- Understanding generated output

### 🧩 Custom Validators & Plugins

#### Built-in Custom Validators

The framework includes several built-in custom validators:

```go
type UserConfig struct {
    IPAddress    string `validate:"ip"`           // IP address validation
    UserID       string `validate:"uuid"`         // UUID format validation
    CreditCard   string `validate:"creditcard"`   // Luhn algorithm validation
    PhoneNumber  string `validate:"phone"`        // Phone number format
}
```

#### Creating Custom Validators

Register user-defined validation rules:

```go
import "github.com/mateothegreat/go-config"

// Define your custom validator
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

// Register with metadata
func init() {
    goconfig.RegisterValidatorWithMetadata("domain", &DomainValidator{}, 
        goconfig.PluginMetadata{
            Name:        "Domain Validator",
            Version:     "1.0.0",
            Description: "Validates domain name format",
            Author:      "Your Name",
            SupportedTypes: []string{"string"},
        })
}
```

### 📦 Multi-Struct & Multi-Package Support

#### Single File Generation (default)
```bash
go-validate generate
# Creates validation_generated.go with all struct validators
```

#### Multi-File Generation
```bash
go-validate generate --multi
# Creates separate files: config_validation_generated.go, settings_validation_generated.go
```

#### Struct Filtering
```bash
go-validate generate --structs Config,Settings
# Only generates validation for specified structs
```

#### Package Targeting
```bash
go-validate generate --input-dir ./internal/config --package config
```

### 💬 Fluent Error Composition

#### Single Field Errors
```go
return goconfig.NewError().
    Field("Email").
    Value(email).
    Tag("email").
    Code("INVALID_EMAIL").
    Format("must be a valid email address")
```

#### Multiple Field Errors
```go
return goconfig.NewError().
    Field("Email").Format("invalid email format").
    Field("Age").Format("must be ≥ 18").
    Field("Name").Required()
```

#### Multi-Error Pattern
```go
fieldErrors := map[string]error{
    "email": goconfig.NewError().Field("Email").Format("invalid"),
    "age":   goconfig.NewError().Field("Age").Format("too young"),
}
return goconfig.NewMultiError(fieldErrors)
```

#### Error Inspection
```go
if err := config.Validate(); err != nil {
    if validationErrors, ok := goconfig.AsValidationErrors(err); ok {
        for _, verr := range validationErrors.Errors() {
            fmt.Printf("Field: %s, Value: %v, Tag: %s, Message: %s\n", 
                verr.Field, verr.Value, verr.Tag, verr.Message)
        }
    }
}
```

## 🛠 CLI Reference

### Commands

| Command | Description | Status |
|---------|-------------|---------|
| `generate` | Run one-time validator generation | ✅ Implemented |
| `watch` | Auto-regenerate on file changes | 🚧 Planned |
| `bench` | Generate and run benchmark tests | 🚧 Planned |
| `version` | Show version information | ✅ Implemented |

### Generate Command

```bash
go-validate generate [flags]
```

#### Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--input-dir` | string | `.` | Input directory to scan for structs |
| `--output-dir` | string | `.` | Custom base directory for output files |
| `--package` | string | auto-detect | Package name for generated code |
| `--structs` | []string | all | Target specific structs (comma-separated) |
| `--dry-run` | bool | false | Output to stdout without writing files |
| `--multi` | bool | false | Generate separate files per struct/package |
| `--verbose` | bool | false | Enable verbose output |

#### Examples

```bash
# Generate validation for all structs in current directory
go-validate generate

# Generate for specific structs only
go-validate generate --structs MyConfig,ServerConfig

# Preview generated code without writing files
go-validate generate --dry-run --verbose

# Generate separate files per struct
go-validate generate --multi

# Generate with custom output directory
go-validate generate --output-dir ./generated

# Scan specific directory with verbose output
go-validate generate --input-dir ./internal/config --verbose

# Generate for specific package
go-validate generate --package mypackage --output-dir ./gen
```

## 🏷 Validation Rules Reference

### Basic Rules

| Rule | Description | Types | Example |
|------|-------------|-------|---------|
| `required` | Field must have a non-zero value | all | `validate:"required"` |
| `min=N` | Minimum numeric value | int, float | `validate:"min=1"` |
| `max=N` | Maximum numeric value | int, float | `validate:"max=100"` |
| `range=min:max` | Numeric range validation | int, float | `validate:"range=1:100"` |

### String Rules

| Rule | Description | Example |
|------|-------------|---------|
| `minlen=N` | Minimum string length | `validate:"minlen=3"` |
| `maxlen=N` | Maximum string length | `validate:"maxlen=50"` |
| `len=N` | Exact string length | `validate:"len=10"` |
| `alpha` | Alphabetic characters only | `validate:"alpha"` |
| `alphanumeric` | Alphanumeric characters only | `validate:"alphanumeric"` |
| `numeric` | Numeric characters only | `validate:"numeric"` |
| `regex=pattern` | Custom regex pattern | `validate:"regex=^[A-Z]+$"` |
| `oneof=a\|b\|c` | Value must be one of options | `validate:"oneof=red\|green\|blue"` |

### Format Rules

| Rule | Description | Example |
|------|-------------|---------|
| `email` | Valid email format (RFC 5322) | `validate:"email"` |
| `url` | Valid URL format | `validate:"url"` |

### Custom Validator Rules

| Rule | Description | Example |
|------|-------------|---------|
| `ip` | IP address (IPv4/IPv6) | `validate:"ip"` |
| `uuid` | UUID format (RFC 4122) | `validate:"uuid"` |
| `creditcard` | Credit card (Luhn algorithm) | `validate:"creditcard"` |
| `phone` | Phone number format | `validate:"phone"` |

### Rule Combination

Rules can be combined with commas:

```go
type User struct {
    Email    string `validate:"required,email,maxlen=100"`
    Age      int    `validate:"required,min=18,max=120"`
    Username string `validate:"required,minlen=3,maxlen=20,alphanumeric"`
    Website  string `validate:"url,maxlen=255"`  // Optional URL
}
```

## 🏷 Struct-Level Annotations

### Go Generate Directives

Use Go directives to opt-in to generation:

```go
//go:generate go-validate -struct MyStruct
type MyStruct struct {
    Port int `validate:"min=1,max=65535"`
}
```

### Field Comments

Document validation behavior:

```go
type Config struct {
    // Port is the server port (1-65535)
    Port int `validate:"min=1,max=65535"`
    
    // Name must be at least 3 characters
    Name string `validate:"required,minlen=3"`
}
```

### Tag Integration

Supports multiple tag types:

```go
type APIConfig struct {
    Endpoint string `json:"endpoint" yaml:"endpoint" validate:"required,url"`
    APIKey   string `json:"api_key" yaml:"api_key" validate:"required,minlen=32"`
    Timeout  int    `json:"timeout" yaml:"timeout" validate:"min=1,max=300"`
}
```

## 🧠 Plugin System Architecture

### Plugin Interface

```go
type ValidatorPlugin interface {
    Validate(field interface{}) error
}

type PluginMetadata struct {
    Name           string
    Version        string
    Description    string
    Author         string
    SupportedTypes []string
}
```

### Plugin Registration

```go
// Simple registration
goconfig.RegisterValidator("custom", &MyValidator{})

// Registration with metadata
goconfig.RegisterValidatorWithMetadata("custom", &MyValidator{}, metadata)

// Get registered validator
if validator, exists := goconfig.GetValidator("custom"); exists {
    err := validator.Validate(fieldValue)
}

// List all validators
validators := goconfig.ListValidators()
```

### Plugin Management

```go
// Get plugin metadata
if metadata, exists := registry.GetMetadata("ip"); exists {
    fmt.Printf("Plugin: %s v%s by %s\n", 
        metadata.Name, metadata.Version, metadata.Author)
}

// Validate with specific plugin
err := registry.ValidateWithPlugin("custom", fieldValue)
```

## 📊 Performance Benchmarks

Generated validation significantly outperforms reflection-based approaches:

```text
BenchmarkGeneratedValidation-8     50000000    23.4 ns/op     0 B/op    0 allocs/op
BenchmarkReflectionValidation-8     1000000   1847 ns/op   512 B/op   23 allocs/op
BenchmarkFluentErrorBuilding-8     10000000    156 ns/op    64 B/op    2 allocs/op
```

**Performance characteristics:**
- **79x faster** than reflection-based validation
- **Zero allocations** for successful validation paths
- **Pre-compiled regex** for format validation
- **Direct field access** without reflection overhead

## 🔧 Integration Examples

### Basic Configuration Loading

```go
package main

import (
    "context"
    "log"
    
    "github.com/mateothegreat/go-config"
    "github.com/mateothegreat/go-config/plugins"
)

type AppConfig struct {
    Name        string `json:"name" validate:"required,minlen=3"`
    Port        int    `json:"port" validate:"min=1,max=65535"`
    DatabaseURL string `json:"database_url" validate:"required,url"`
    Debug       bool   `json:"debug"`
}

func main() {
    config := &AppConfig{}
    
    // Configure loader with multiple sources
    loader := goconfig.NewConfigLoader(config).
        Use(plugins.NewEnvPlugin("APP_")).
        Use(plugins.NewYAMLPlugin("config.yaml")).
        SetDefaults(&AppConfig{
            Port:  8080,
            Debug: false,
        })
    
    // Load and validate configuration
    if err := loader.Load(context.Background()); err != nil {
        log.Fatalf("Configuration error: %v", err)
    }
    
    fmt.Printf("Starting %s on port %d\n", config.Name, config.Port)
}
```

### Advanced Error Handling

```go
func validateConfig(config *AppConfig) {
    if err := config.Validate(); err != nil {
        switch e := err.(type) {
        case goconfig.ValidationErrors:
            fmt.Println("Validation failed:")
            for _, verr := range e.Errors() {
                fmt.Printf("  • %s: %s\n", verr.Field, verr.Message)
            }
        case *goconfig.MultiError:
            fmt.Println("Multiple validation errors:")
            for field, fieldErr := range e.Errors() {
                fmt.Printf("  • %s: %s\n", field, fieldErr.Error())
            }
        default:
            fmt.Printf("Validation error: %v\n", err)
        }
    }
}
```

### Custom Validation Integration

```go
type DatabaseConfig struct {
    Host     string `validate:"required"`
    Port     int    `validate:"min=1,max=65535"`
    Database string `validate:"required,custom_db_name"`
    SSL      bool   `validate:""`
}

// Custom validator for database names
type DBNameValidator struct{}

func (v *DBNameValidator) Validate(field interface{}) error {
    name, ok := field.(string)
    if !ok {
        return fmt.Errorf("database name must be a string")
    }
    
    if strings.Contains(name, " ") {
        return fmt.Errorf("database name cannot contain spaces")
    }
    
    if len(name) > 63 {
        return fmt.Errorf("database name too long (max 63 characters)")
    }
    
    return nil
}

func init() {
    goconfig.RegisterValidator("custom_db_name", &DBNameValidator{})
}
```

### Generated Code Example

For the struct:

```go
type ServerConfig struct {
    Name  string `validate:"required,minlen=3"`
    Port  int    `validate:"min=1,max=65535"`
    Email string `validate:"required,email"`
}
```

The generator produces:

```go
// Code generated by go-validate. DO NOT EDIT.
package main

import (
    "regexp"
    "github.com/mateothegreat/go-config"
)

var (
    emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
)

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

## 🧪 Testing

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -race -coverprofile=coverage.out ./...

# Run benchmarks
go test -bench=. ./...

# Run specific package tests
go test ./internal/generator -v

# Run integration tests
go test . -v
```

### Test Coverage

Current test coverage:
- **internal/errors**: 97.7%
- **internal/plugins**: 94.3%
- **internal/scanner**: 86.9%
- **internal/generator**: 40.2%
- **main package**: 58.0%

### Writing Tests for Generated Code

```go
func TestGeneratedValidation(t *testing.T) {
    tests := []struct {
        name    string
        config  ServerConfig
        wantErr bool
    }{
        {
            name: "valid config",
            config: ServerConfig{
                Name:  "test-server",
                Port:  8080,
                Email: "admin@example.com",
            },
            wantErr: false,
        },
        {
            name: "invalid port",
            config: ServerConfig{
                Name:  "test-server",
                Port:  70000,
                Email: "admin@example.com",
            },
            wantErr: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := tt.config.Validate()
            if (err != nil) != tt.wantErr {
                t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

## 🔧 Development

### Building from Source

```bash
git clone https://github.com/mateothegreat/go-config
cd go-config
go build -o bin/go-validate ./cmd/go-validate
```

### Project Structure

```text
go-config/
├── cmd/
│   ├── go-config-gen/     # Legacy generator (deprecated)
│   └── go-validate/       # Main CLI tool
├── internal/
│   ├── errors/           # Fluent error composition
│   ├── generator/        # Code generation engine
│   ├── plugins/          # Plugin system
│   └── scanner/          # AST scanning
├── examples/             # Usage examples
├── plugins/              # Built-in plugins
└── docs/                # Documentation
```

## 📚 Examples and Learning Resources

### 🎯 Comprehensive Examples

The [`examples/`](examples/) directory contains complete, runnable examples for every use case:

#### 🚀 Getting Started
- [`01-basic-validation/`](examples/01-basic-validation/) - **Start here!** Simple struct validation with common rules
- [`02-go-generate/`](examples/02-go-generate/) - Using `//go:generate` for automated code generation
- [`03-cli-usage/`](examples/03-cli-usage/) - Manual CLI usage and dry-run previews

#### 🏗️ Real-World Applications  
- [`04-microservice-config/`](examples/04-microservice-config/) - Production microservice configuration
- [`05-api-request-validation/`](examples/05-api-request-validation/) - HTTP API request validation
- [`06-config-loading/`](examples/06-config-loading/) - Multi-source configuration loading (ENV, YAML, defaults)

#### 🔧 Advanced Features
- [`07-custom-validators/`](examples/07-custom-validators/) - **Custom validator plugins** - Create domain-specific validation rules
- [`08-embedding-composition/`](examples/08-embedding-composition/) - **Struct embedding patterns** - Complex validation hierarchies
- [`09-multi-package/`](examples/09-multi-package/) - Cross-package validation with separate generated files
- [`10-error-handling/`](examples/10-error-handling/) - Advanced error handling and reporting patterns

#### 🔄 Integration & Deployment
- [`11-ci-cd-integration/`](examples/11-ci-cd-integration/) - GitHub Actions, CI/CD pipeline integration
- [`12-build-automation/`](examples/12-build-automation/) - Makefile and build script integration  
- [`13-performance-testing/`](examples/13-performance-testing/) - Benchmarking and performance optimization

### 📖 Quick Learning Path

**New to go-validate?** Follow this path:
1. 🎯 **[Basic Validation](examples/01-basic-validation/)** - Learn the fundamentals
2. 🔄 **[Go Generate](examples/02-go-generate/)** - Automate code generation
3. 🏗️ **[Microservice Config](examples/04-microservice-config/)** - See real-world usage
4. 🔧 **[Custom Validators](examples/07-custom-validators/)** - Create custom rules
5. 🚀 **[CI/CD Integration](examples/11-ci-cd-integration/)** - Production deployment

### 📋 Documentation Files

- [`docs/performance.md`](docs/performance.md) - **Performance analysis** - 79x faster than reflection-based validation
- [`docs/zero-reflection-validation.md`](docs/zero-reflection-validation.md) - **Technical deep dive** - How zero-reflection works

### API Reference

The complete API documentation is available through:

```bash
go doc github.com/mateothegreat/go-config
```

### Migration Guides

If migrating from reflection-based validators:

1. Replace validation logic with struct tags
2. Generate validation code with `go-validate generate`
3. Update error handling to use fluent API
4. Remove reflection-based validator dependencies

## 📝 License

MIT License - see [LICENSE](LICENSENSE) file for details.

## 🤝 Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Write tests for your changes
4. Ensure all tests pass (`go test ./...`)
5. Run code generation tests
6. Commit your changes (`git commit -m 'Add some amazing feature'`)
7. Push to the branch (`git push origin feature/amazing-feature`)
8. Open a Pull Request

### Development Guidelines

- Follow Go best practices and idiomatic code
- Maintain test coverage above 80%
- Update documentation for new features
- Add examples for new validation rules
- Ensure generated code is optimal and readable

## 🙏 Acknowledgments

- Built with Go's powerful AST parsing capabilities
- Inspired by modern validation libraries while prioritizing performance
- Thanks to the Go community for feedback and contributions
- Special thanks to contributors who helped shape the fluent API design