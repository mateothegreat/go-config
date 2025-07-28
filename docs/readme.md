# go-config

**A unified, well-architected configuration management library for Go with built-in validation and error correlation.**

## Features

- 🏗️ **Well-architected design** with clear separation of concerns
- ⚡ **Multiple validation strategies** - Auto-detection, Generated, Reflection, and Fast validation
- 🔗 **Unified plugin system** for configuration sources (YAML, Environment variables, etc.)
- 🎯 **Advanced error correlation** with contextual suggestions
- 📝 **Fluent API** for both configuration loading and validation
- 🚀 **Zero-reflection code generation** for high-performance validation
- 💡 **Intelligent field mapping** with automatic tag detection

## Quick Start

### Installation

```bash
go get github.com/mateothegreat/go-config
```

### Basic Usage

Let's first setup our configuration structs. Create a file called `config.go` in the root of your project with the following content:

```go
package main

type ServerConfig struct {
 Name string `validate:"required,minlen=3,maxlen=50" yaml:"name"`
 Host string `validate:"required" yaml:"host"`
 Port int    `validate:"min=1,max=65535" yaml:"port"`

 Environment string `validate:"oneof=dev|staging|prod" yaml:"environment"`
 Version     string `validate:"required,regex=^v[0-9]+\\.[0-9]+\\.[0-9]+$" yaml:"version"`

 Debug   bool `yaml:"debug"`
 Metrics bool `yaml:"metrics"`
 AdminEmail string `validate:"required,email" yaml:"admin_email"`
}

type DatabaseConfig struct {
 Host     string `validate:"required" yaml:"host"`
 Port     int    `validate:"min=1,max=65535" yaml:"port"`
 Database string `validate:"required,minlen=1,maxlen=63" yaml:"database"`
 Username string `validate:"required" yaml:"username"`
 Password string `validate:"required,minlen=8" yaml:"password"`
 MaxConnections int `validate:"min=1,max=1000" yaml:"max_connections"`
 Timeout        int `validate:"min=1,max=300" yaml:"timeout"`
 SSLMode string `validate:"oneof=disable|require|verify-ca|verify-full" yaml:"ssl_mode"`
}

type APIConfig struct {
 BaseURL string `validate:"required,url" yaml:"base_url"`
 APIKey  string `validate:"required,len=32,alphanumeric" yaml:"api_key"`
 Timeout int    `validate:"min=1,max=120" yaml:"timeout"`
 RateLimit        int  `validate:"min=1,max=10000" yaml:"rate_limit"`
 RateLimitEnabled bool `yaml:"rate_limit_enabled"`
 TLSVerify bool `yaml:"tls_verify"`
}

type Config struct {
 Server   ServerConfig   `yaml:"server"`
 Database DatabaseConfig `yaml:"database"`
 API      APIConfig      `yaml:"api"`
}
```

### Loader API (Advanced Usage)

```go
func main() {
  cfg := &Config{}

  err := config.LoadWithPlugins(
    config.FromYAML(sources.YAMLOpts{Path: "config.yaml"}),
    config.FromEnv(sources.EnvOpts{Prefix: "APP"}),
  ).WithValidationStrategy(validation.StrategyAuto).Build(cfg)
  if err != nil {
    doSomethingWithTheError(err)
  }

  fmt.Printf("✅ Configuration loaded: %+v\n", cfg) // or cfg.Server.Environment, cfg.Database.Host, cfg.API.BaseURL, etc.
}
```

## Examples

See the [`examples/`](../examples) directory for complete working examples:

- [`1-basic-validation/`](../examples/1-basic-validation) - Basic configuration loading and validation
- [`2-embedded-generation/`](../examples/2-embedded-generation) - Code generation with nested structs
- [`3-cli-usage/`](../examples/3-cli-usage) - Command-line integration

## Validation Strategies

The library supports multiple validation approaches:

- **Auto** - Automatically detects the best strategy
- **Generated** - Uses pre-generated validation code (fastest)
- **Reflection** - Runtime reflection-based validation
- **Fast** - Type-specific optimized validation

## Error Correlation

Advanced error correlation provides contextual suggestions:

```go
if correlatedErr, ok := err.(*goconfig.CorrelatedError); ok {
    fmt.Println("Error:", correlatedErr.Error())
    fmt.Println("Suggestions:")
    for _, suggestion := range correlatedErr.GetSuggestions() {
        fmt.Printf("  - %s\n", suggestion)
    }
}
```
