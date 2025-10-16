package main

//go:generate go run ../../cmd/go-config-gen/main.go --ast

// ServerConfig demonstrates AST-based validation generation
type ServerConfig struct {
	// Server identification
	Name        string `yaml:"name" validate:"required,minlen=3,maxlen=50"`
	Environment string `yaml:"environment" validate:"required,oneof=dev staging prod"`

	// Network configuration
	Host      string `yaml:"host" validate:"required,ip|hostname"`
	Port      int    `yaml:"port" validate:"required,min=1,max=65535"`
	PublicURL string `yaml:"public_url" validate:"required,url"`

	// Security settings
	APIKey       string   `yaml:"api_key" validate:"required,minlen=32,base64"`
	SecretKey    string   `yaml:"secret_key" validate:"required,len=64"`
	AllowedCIDRs []string `yaml:"allowed_cidrs" validate:"dive,cidr"`

	// Database configuration
	Database DatabaseConfig `yaml:"database" validate:"required"`

	// Performance settings
	MaxConnections int     `yaml:"max_connections" validate:"min=10,max=10000"`
	Timeout        float64 `yaml:"timeout" validate:"min=0.1,max=300"`

	// Feature flags
	Features map[string]bool `yaml:"features"`
}

// DatabaseConfig contains database connection settings
type DatabaseConfig struct {
	Driver   string `yaml:"driver" validate:"required,oneof=postgres mysql sqlite"`
	Host     string `yaml:"host" validate:"required_if=Driver postgres,required_if=Driver mysql,hostname|ip"`
	Port     int    `yaml:"port" validate:"required_if=Driver postgres,required_if=Driver mysql,min=1,max=65535"`
	Username string `yaml:"username" validate:"required_if=Driver postgres,required_if=Driver mysql"`
	Password string `yaml:"password" validate:"required_if=Driver postgres,required_if=Driver mysql"`
	Database string `yaml:"database" validate:"required"`

	// Connection pool settings
	MaxOpenConns    int `yaml:"max_open_conns" validate:"min=1,max=100"`
	MaxIdleConns    int `yaml:"max_idle_conns" validate:"min=0,max=100"`
	ConnMaxLifetime int `yaml:"conn_max_lifetime" validate:"min=0"`
}

// EmailConfig demonstrates email validation
type EmailConfig struct {
	FromAddress  string   `yaml:"from_address" validate:"required,email"`
	ReplyTo      string   `yaml:"reply_to" validate:"email"`
	AdminEmails  []string `yaml:"admin_emails" validate:"required,min=1,dive,email"`
	SMTPHost     string   `yaml:"smtp_host" validate:"required,hostname"`
	SMTPPort     int      `yaml:"smtp_port" validate:"required,min=1,max=65535"`
	SMTPUsername string   `yaml:"smtp_username" validate:"required"`
	SMTPPassword string   `yaml:"smtp_password" validate:"required"`
	UseTLS       bool     `yaml:"use_tls"`
}
