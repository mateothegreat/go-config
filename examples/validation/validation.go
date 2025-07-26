package main

import (
	"log"

	goconfig "github.com/mateothegreat/go-config"
	"github.com/mateothegreat/go-config/plugins"
)

// ValidtationConfig is a config struct with strict validation rules.
// It is used to test the validation rules.
type ValidtationConfig struct {
	// Server configuration
	Host     string `config:"host" validate:"required,regex=^[a-zA-Z0-9.-]+$"`
	Port     int    `config:"port" validate:"required,range=1000:65535"`
	Protocol string `config:"protocol" validate:"oneof=http|https"`

	// Database configuration
	DatabaseURL string `config:"database_url" validate:"required,url"`
	MaxConns    int    `config:"max_conns" validate:"min=1,max=100"`

	// Authentication
	JWTSecret  string `config:"jwt_secret" validate:"required,minlen=32"`
	AdminEmail string `config:"admin_email" validate:"required,email"`

	// Features
	EnableDebug bool   `config:"enable_debug"`
	LogLevel    string `config:"log_level" validate:"oneof=debug|info|warn|error"`
	Version     string `config:"version" validate:"required,regex=^v[0-9]+\\.[0-9]+\\.[0-9]+$"`

	// API Keys
	APIKey        string `config:"api_key" validate:"required,len=40,alphanumeric"`
	WebhookSecret string `config:"webhook_secret" validate:"required,minlen=16,maxlen=64"`
}

func main() {
	var cfg ValidtationConfig

	err := goconfig.LoadWithPlugins(
		goconfig.Env(plugins.EnvOpts{}),
		goconfig.YAML(plugins.YAMLOpts{Path: "validation.yaml"}),
	).Build(&cfg)

	if err != nil {
		if validationErrors, ok := goconfig.AsValidationErrors(err); ok {
			log.Println("Validation failed with the following errors:")
			for _, e := range validationErrors {
				log.Printf("  - %s", e)
			}
		} else {
			log.Printf("Error: %v", err)
		}
		log.Fatal("failed to load config")
	}

	log.Println("✅ Configuration loaded and validated successfully!")
	log.Printf("Config: Host=%s, Port=%d, Protocol=%s", cfg.Host, cfg.Port, cfg.Protocol)
}
