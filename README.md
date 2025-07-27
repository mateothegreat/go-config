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

```go
package main

import (
    "context"
    goconfig "github.com/mateothegreat/go-config"
    "github.com/mateothegreat/go-config/plugins/sources"
)

type ServerConfig struct {
    Name     string `validate:"required,minlen=3" yaml:"name"`
    Port     int    `validate:"min=1,max=65535" yaml:"port"`
    Host     string `validate:"required" yaml:"host"`
    Email    string `validate:"required,email" yaml:"email"`
    LogLevel string `yaml:"log_level"`
    Debug    bool   `yaml:"debug"`
}

func main() {
    config := &ServerConfig{}
    
    // Method 1: Fluent API
    err := goconfig.LoadWithPlugins(
        goconfig.FromYAML(sources.YAMLOpts{Path: "config.yaml"}),
        goconfig.FromEnv(sources.EnvOpts{Prefix: "SERVER"}),
    ).Build(config)
    
    if err != nil {
        // Enhanced error handling with suggestions
        if correlatedErr, ok := err.(*goconfig.CorrelatedError); ok {
            fmt.Printf("Error: %s\n", correlatedErr.Error())
            for _, suggestion := range correlatedErr.GetSuggestions() {
                fmt.Printf("💡 %s\n", suggestion)
            }
        }
        return
    }
    
    fmt.Printf("✅ Loaded config: %+v\n", config)
}
```

### Loader API (Advanced Usage)

```go
func main() {
    config := &ServerConfig{}
    
    // Method 2: Loader API for more control
    loader := goconfig.NewLoader(config)
    
    // Add sources
    yamlPlugin, _ := goconfig.CreateSourcePlugin("yaml", sources.YAMLOpts{Path: "config.yaml"})
    envPlugin, _ := goconfig.CreateSourcePlugin("env", sources.EnvOpts{Prefix: "SERVER"})
    
    loader.Use(yamlPlugin)
    loader.Use(envPlugin)
    
    // Set defaults
    loader.SetDefaults(&ServerConfig{
        Host:     "localhost",
        Port:     8080,
        LogLevel: "info",
        Debug:    false,
    })
    
    // Load and validate
    if err := loader.Load(context.Background()); err != nil {
        // Handle error with correlation
        return
    }
    
    // Inspect configuration sources
    fmt.Println("Sources used:", loader.Sources())
    fmt.Println("Merged config:", loader.Inspect())
}
```

## Code Generation

Generate high-performance validation code:

```bash
go run cmd/go-config-gen/main.go -input . -output validation_generated.go -verbose
```

This generates zero-reflection validation methods for your structs.

## Architecture

The library is organized into well-separated packages:

- **`config/`** - Core configuration loading and hydration
- **`validate/`** - Validation strategies and unified validator  
- **`plugins/`** - Plugin system for sources and validators
- **`errors/`** - Advanced error handling and correlation
- **`generate/`** - Code generation utilities

## Examples

See the [`examples/`](./examples/) directory for complete working examples:

- [`1-basic-validation/`](./examples/1-basic-validation/) - Basic configuration loading and validation
- [`2-embedded-generation/`](./examples/2-embedded-generation/) - Code generation with nested structs
- [`3-cli-usage/`](./examples/3-cli-usage/) - Command-line integration

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

## Contributing

1. Fork the repository
2. Create a feature branch
3. Add tests for new functionality
4. Run the test suite: `go test -v ./...`
5. Submit a pull request

## License

This project is licensed under the MIT License.