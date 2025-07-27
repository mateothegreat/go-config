//go:generate go-validate generate --structs AppConfig --verbose

package main

// AppConfig represents the main application configuration
// The go:generate directive above will create validation code for this struct
type AppConfig struct {
	// Application identification
	Name        string `validate:"required,minlen=3,maxlen=50" yaml:"name"`
	Environment string `validate:"oneof=dev|staging|prod" yaml:"environment"`
	Version     string `validate:"required,regex=^v[0-9]+\\.[0-9]+\\.[0-9]+$" yaml:"version"`

	// Server configuration (embedded)
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`

	// Feature flags
	Features FeatureConfig `yaml:"features"`
}
