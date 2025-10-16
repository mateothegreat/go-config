package main

// Embedded struct definitions
type BaseConfig struct {
	Host    string `yaml:"host" validate:"required"`
	Port    int    `yaml:"port" validate:"required,min=1,max=65535"`
	Default string `yaml:"default" validate:"required" default:"default"`
}

type AuthConfig struct {
	Username string `yaml:"username" validate:"required"`
	Password string `yaml:"password" validate:"required"`
}

type AppConfig struct {
	Name        string    `yaml:"name" validate:"required,minlen=3,maxlen=50"`
	Environment string    `yaml:"environment" validate:"oneof=dev staging prod"`
	Version     string    `yaml:"version" validate:"required,regex=^v[0-9]+\\.[0-9]+\\.[0-9]+$"`
	Sub         SubConfig `yaml:"sub" validate:"required"`
	// Test embedded structs - this should cause the issue
	Redis RedisConfig `yaml:"redis" validate:"required"`
}

type SubConfig struct {
	Foo string `yaml:"foo" validate:"required,minlen=3"`
}

// RedisConfig with embedded structs
type RedisConfig struct {
	BaseConfig        // Embedded - fields should be flattened
	AuthConfig        // Embedded - fields should be flattened
	Database   int    `yaml:"database" validate:"min=0,max=15"`
	Timeout    string `yaml:"timeout" validate:"required"`
}
