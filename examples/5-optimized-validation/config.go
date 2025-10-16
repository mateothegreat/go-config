package main

//go:generate go run ../../validator_codegen.go

// ServerConfig demonstrates optimized validation with zero reflection
type ServerConfig struct {
	// Server identification
	Name        string `validate:"required,minlen=3,maxlen=50"`
	Environment string `validate:"required,oneof=dev staging prod"`

	// Network configuration
	Host string `validate:"required,hostname"`
	Port int    `validate:"required,min=1,max=65535"`

	// Contact information
	Email       string   `validate:"required,email"`
	AdminEmails []string `validate:"required,min=1"`

	// Feature flags
	Debug   bool `validate:"required"`
	Enabled bool
}

// DatabaseConfig contains database connection settings
type DatabaseConfig struct {
	Driver   string `validate:"required,oneof=postgres mysql sqlite"`
	Host     string `validate:"required,hostname"`
	Port     int    `validate:"required,min=1,max=65535"`
	Username string `validate:"required"`
	Password string `validate:"required"`
	Database string `validate:"required"`

	// Connection pool settings
	MaxOpenConns int `validate:"min=1,max=100"`
	MaxIdleConns int `validate:"min=0,max=100"`
}

// APIConfig contains API-specific configuration
type APIConfig struct {
	BaseURL   string  `validate:"required,url"`
	APIKey    string  `validate:"required,base64"`
	Version   string  `validate:"required"`
	Timeout   float32 `validate:"min=0.1,max=300"`
	RateLimit int     `validate:"min=1,max=10000"`
}

// Config is the main configuration struct with nested configurations
type Config struct {
	Server   ServerConfig   `validate:"required"`
	Database DatabaseConfig `validate:"required"`
	API      APIConfig      `validate:"required"`

	// Global settings
	LogLevel string `validate:"oneof=debug info warn error"`
	Debug    bool
}