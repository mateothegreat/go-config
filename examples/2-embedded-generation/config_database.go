//go:generate go-validate generate --structs DatabaseConfig

package main

// DatabaseConfig handles database connection configuration
type DatabaseConfig struct {
	// Connection settings
	Host     string `validate:"required" yaml:"host"`
	Port     int    `validate:"min=1,max=65535" yaml:"port"`
	Database string `validate:"required,minlen=1,maxlen=63" yaml:"database"`
	Username string `validate:"required" yaml:"username"`
	Password string `validate:"required,minlen=8" yaml:"password"`

	// Connection pool settings
	MaxConnections     int `validate:"min=1,max=1000" yaml:"max_connections"`
	MaxIdleConnections int `validate:"min=1,max=100" yaml:"max_idle_connections"`
	ConnMaxLifetime    int `validate:"min=0" yaml:"conn_max_lifetime"` // seconds, 0 = unlimited

	// TLS settings
	SSLMode string `validate:"oneof=disable|require|verify-ca|verify-full" yaml:"ssl_mode"`

	// Migration settings
	AutoMigrate bool `yaml:"auto_migrate"`
}
