# Embedding and Composition Example

This example demonstrates advanced struct composition patterns with `go-validate`, including embedding, interfaces, and complex validation scenarios.

## 📋 What You'll Learn

- Struct embedding with validation
- Interface-based configuration
- Composition patterns for reusable config
- Validation inheritance and override
- Complex validation scenarios

## 🏗️ Files

- `base/` - Base configuration structs
- `services/` - Service-specific configurations
- `main.go` - Composition and validation example
- `interfaces.go` - Interface definitions

## 🚀 Quick Start

1. **Generate validation code:**

   ```bash
   go-validate generate
   ```

2. **Run the example:**

   ```bash
   go run .
   ```

## 📖 Embedding Patterns

### 1. Base Configuration Embedding

```go
// Base configuration that other configs embed
type BaseConfig struct {
    Name        string `validate:"required,minlen=3"`
    Environment string `validate:"oneof=dev|staging|prod"`
    Version     string `validate:"required,regex=^v[0-9]+\\.[0-9]+\\.[0-9]+$"`
}

// Service configuration embeds base config
type WebServiceConfig struct {
    BaseConfig  // Embedded struct inherits validation
    
    Port        int    `validate:"min=1,max=65535"`
    Host        string `validate:"required"`
    TLSEnabled  bool
}
```

### 2. Network Configuration Composition

```go
type NetworkConfig struct {
    Interface string `validate:"required"`
    MTU       int    `validate:"min=576,max=9000"`
    Timeout   int    `validate:"min=1,max=300"`
}

type ServerConfig struct {
    BaseConfig    // Application-level config
    NetworkConfig // Network-level config
    
    // Service-specific fields
    MaxConnections int `validate:"min=1,max=10000"`
}
```

### 3. Conditional Validation with Embedding

```go
type TLSConfig struct {
    Enabled  bool   `validate:""`
    CertFile string `validate:"required_if=Enabled:true"`
    KeyFile  string `validate:"required_if=Enabled:true"`
    MinVersion string `validate:"oneof=1.0|1.1|1.2|1.3"`
}

type HTTPSServerConfig struct {
    BaseConfig
    TLSConfig
    
    RedirectHTTP bool `validate:""`
}
```

## 🔧 Interface-Based Configuration

### Configuration Provider Interface

```go
type ConfigProvider interface {
    GetName() string
    GetEnvironment() string
    Validate() error
}

type Configurable interface {
    LoadConfig() error
    ValidateConfig() error
}
```

### Implementation with Embedding

```go
type DatabaseService struct {
    BaseConfig
    
    Host     string `validate:"required"`
    Port     int    `validate:"min=1,max=65535"`
    Database string `validate:"required"`
    Username string `validate:"required"`
    Password string `validate:"required,minlen=8"`
}

func (d *DatabaseService) GetName() string {
    return d.Name
}

func (d *DatabaseService) GetEnvironment() string {
    return d.Environment
}

// Validation method will be generated for the embedded BaseConfig
// and the DatabaseService fields
```

## 🎯 Advanced Composition Patterns

### 1. Mixin Pattern

```go
// Reusable configuration mixins
type ObservabilityMixin struct {
    MetricsEnabled bool   `validate:""`
    MetricsPort    int    `validate:"min=1,max=65535"`
    TracingEnabled bool   `validate:""`
    LogLevel       string `validate:"oneof=debug|info|warn|error"`
}

type SecurityMixin struct {
    AuthEnabled    bool   `validate:""`
    JWTSecret      string `validate:"required_if=AuthEnabled:true,minlen=32"`
    SessionTimeout int    `validate:"min=60,max=86400"` // 1 minute to 24 hours
}

type FullServiceConfig struct {
    BaseConfig
    ObservabilityMixin
    SecurityMixin
    
    // Service-specific configuration
    BusinessLogicTimeout int `validate:"min=1,max=3600"`
}
```

### 2. Override Pattern

```go
type DefaultServerConfig struct {
    Host           string `validate:"required"`
    Port           int    `validate:"min=1,max=65535"`
    ReadTimeout    int    `validate:"min=1,max=300"`
    WriteTimeout   int    `validate:"min=1,max=300"`
}

type ProductionServerConfig struct {
    DefaultServerConfig
    
    // Override with stricter validation for production
    ReadTimeout  int `validate:"min=5,max=60"`   // Stricter timeouts
    WriteTimeout int `validate:"min=5,max=60"`   // Stricter timeouts
    
    // Additional production-only fields
    HealthCheckPath string `validate:"required"`
    MetricsPath     string `validate:"required"`
}
```

### 3. Hierarchical Configuration

```go
type ClusterConfig struct {
    Name        string `validate:"required"`
    Environment string `validate:"oneof=dev|staging|prod"`
    Region      string `validate:"required"`
}

type NodeConfig struct {
    ClusterConfig // Inherits cluster-level settings
    
    NodeID   string `validate:"required,uuid"`
    Role     string `validate:"oneof=master|worker"`
    Resources ResourceConfig
}

type ResourceConfig struct {
    CPU    int `validate:"min=1,max=64"`
    Memory int `validate:"min=512,max=65536"` // MB
    Disk   int `validate:"min=1,max=1024"`    // GB
}

type ServiceConfig struct {
    NodeConfig // Inherits node and cluster settings
    
    ServiceName string `validate:"required"`
    Port        int    `validate:"min=1,max=65535"`
    Replicas    int    `validate:"min=1,max=100"`
}
```

## 🔄 Validation Strategies

### 1. Embedded Validation

```go
//go:generate go-validate generate --structs BaseConfig,WebServiceConfig

// Both structs get their own Validate() methods
func Example() {
    config := &WebServiceConfig{
        BaseConfig: BaseConfig{
            Name: "web-service",
            Environment: "prod",
            Version: "v1.0.0",
        },
        Port: 8080,
        Host: "localhost",
    }
    
    // Validates all fields including embedded BaseConfig
    if err := config.Validate(); err != nil {
        log.Fatal(err)
    }
}
```

### 2. Composite Validation

```go
func (s *ServiceConfig) ValidateComplete() error {
    // Validate the service config
    if err := s.Validate(); err != nil {
        return fmt.Errorf("service validation failed: %w", err)
    }
    
    // Additional business logic validation
    if s.Role == "master" && s.Replicas > 1 {
        return fmt.Errorf("master role cannot have multiple replicas")
    }
    
    if s.Environment == "prod" && s.Resources.CPU < 2 {
        return fmt.Errorf("production environment requires at least 2 CPU cores")
    }
    
    return nil
}
```

### 3. Factory Pattern with Validation

```go
type ConfigFactory struct{}

func (f *ConfigFactory) CreateServiceConfig(serviceType string) (ConfigProvider, error) {
    var config ConfigProvider
    
    switch serviceType {
    case "web":
        config = &WebServiceConfig{}
    case "database":
        config = &DatabaseService{}
    case "cache":
        config = &CacheService{}
    default:
        return nil, fmt.Errorf("unknown service type: %s", serviceType)
    }
    
    // All configs implement ConfigProvider which includes Validate()
    if err := config.Validate(); err != nil {
        return nil, fmt.Errorf("invalid config for %s: %w", serviceType, err)
    }
    
    return config, nil
}
```

## 🧪 Testing Composition

```go
func TestEmbeddedValidation(t *testing.T) {
    tests := []struct {
        name    string
        config  WebServiceConfig
        wantErr bool
    }{
        {
            name: "valid embedded config",
            config: WebServiceConfig{
                BaseConfig: BaseConfig{
                    Name: "test-service",
                    Environment: "dev",
                    Version: "v1.0.0",
                },
                Port: 8080,
                Host: "localhost",
            },
            wantErr: false,
        },
        {
            name: "invalid base config name",
            config: WebServiceConfig{
                BaseConfig: BaseConfig{
                    Name: "x", // Too short
                    Environment: "dev",
                    Version: "v1.0.0",
                },
                Port: 8080,
                Host: "localhost",
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

## 🎓 Best Practices

### 1. Clear Inheritance Hierarchy

```go
// Good: Clear inheritance chain
type Base struct { /* common fields */ }
type Service struct { Base; /* service fields */ }
type WebService struct { Service; /* web-specific fields */ }
```

### 2. Logical Composition

```go
// Good: Logical grouping
type ServerConfig struct {
    NetworkConfig
    SecurityConfig
    ObservabilityConfig
}
```

### 3. Avoid Deep Nesting

```go
// Avoid: Too many levels
type Config struct {
    Level1 struct {
        Level2 struct {
            Level3 struct {
                Value string
            }
        }
    }
}

// Better: Flatter structure
type Config struct {
    DatabaseConfig
    CacheConfig
    APIConfig
}
```

## 🎓 Next Steps

- See multi-package examples in [`../09-multi-package/`](../examples/09-multi-package)
- Learn error handling in [`../10-error-handling/`](../examples/10-error-handling)
- Explore CI/CD integration in [`../11-ci-cd-integration/`](../examples/11-ci-cd-integration)
