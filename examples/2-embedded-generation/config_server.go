//go:generate go-validate generate --structs ServerConfig

package main

// ServerConfig handles HTTP server configuration
type ServerConfig struct {
	// Network settings
	Host string `validate:"required" yaml:"host"`
	Port int    `validate:"min=1,max=65535" yaml:"port"`

	// TLS settings
	TLS     bool   `yaml:"tls"`
	CertFile string `yaml:"cert_file"`
	KeyFile  string `yaml:"key_file"`

	// Timeouts (in seconds)
	ReadTimeout  int `validate:"min=1,max=300" yaml:"read_timeout"`
	WriteTimeout int `validate:"min=1,max=300" yaml:"write_timeout"`

	// CORS settings
	CORS CORSConfig `yaml:"cors"`
}

// CORSConfig handles Cross-Origin Resource Sharing settings
type CORSConfig struct {
	Enabled     bool     `yaml:"enabled"`
	Origins     []string `yaml:"origins"`
	Methods     []string `yaml:"methods"`
	Headers     []string `yaml:"headers"`
	Credentials bool     `yaml:"credentials"`
}