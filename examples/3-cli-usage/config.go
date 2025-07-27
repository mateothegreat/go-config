package main

// Note: No go:generate directives - this example uses manual CLI commands

// ServerConfig demonstrates manual CLI validation generation
type ServerConfig struct {
	// Basic server settings
	Name string `validate:"required,minlen=3,maxlen=50" yaml:"name"`
	Host string `validate:"required" yaml:"host"`
	Port int    `validate:"min=1,max=65535" yaml:"port"`

	// Service configuration
	Environment string `validate:"oneof=dev|staging|prod" yaml:"environment"`
	Version     string `validate:"required,regex=^v[0-9]+\\.[0-9]+\\.[0-9]+$" yaml:"version"`

	// Feature flags
	Debug   bool `yaml:"debug"`
	Metrics bool `yaml:"metrics"`

	// Contact information
	AdminEmail string `validate:"required,email" yaml:"admin_email"`
}

// DatabaseConfig demonstrates validation with database-specific rules
type DatabaseConfig struct {
	// Connection settings
	Host     string `validate:"required" yaml:"host"`
	Port     int    `validate:"min=1,max=65535" yaml:"port"`
	Database string `validate:"required,minlen=1,maxlen=63" yaml:"database"`
	Username string `validate:"required" yaml:"username"`
	Password string `validate:"required,minlen=8" yaml:"password"`

	// Connection pool
	MaxConnections int `validate:"min=1,max=1000" yaml:"max_connections"`
	Timeout        int `validate:"min=1,max=300" yaml:"timeout"`

	// SSL settings
	SSLMode string `validate:"oneof=disable|require|verify-ca|verify-full" yaml:"ssl_mode"`
}

// APIConfig demonstrates API-specific validation rules
type APIConfig struct {
	// API settings
	BaseURL string `validate:"required,url" yaml:"base_url"`
	APIKey  string `validate:"required,len=32,alphanumeric" yaml:"api_key"`
	Timeout int    `validate:"min=1,max=120" yaml:"timeout"`

	// Rate limiting
	RateLimit    int  `validate:"min=1,max=10000" yaml:"rate_limit"`
	RateLimitEnabled bool `yaml:"rate_limit_enabled"`

	// Security
	TLSVerify bool `yaml:"tls_verify"`
}