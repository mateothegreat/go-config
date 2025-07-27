package main

// ServerConfig demonstrates basic validation with common rules
type ServerConfig struct {
	// Server identification
	Name string `validate:"required,minlen=3,maxlen=50" yaml:"name" json:"name"`

	// Network settings
	Port int    `validate:"min=1,max=65535" yaml:"port" json:"port"`
	Host string `validate:"required" yaml:"host" json:"host"`

	// Contact information
	Email string `validate:"required,email" yaml:"email" json:"email"`

	// Operational settings
	LogLevel string `validate:"oneof=debug|info|warn|error" yaml:"log_level" json:"log_level"`
	Debug    bool   `yaml:"debug" json:"debug"`

	// Version information
	Version string `validate:"required,regex=^v[0-9]+\\.[0-9]+\\.[0-9]+$" yaml:"version" json:"version"`
}
