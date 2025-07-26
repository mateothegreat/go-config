package main

import (
	"fmt"
	"log"

	goconfig "github.com/mateothegreat/go-config"
	"github.com/mateothegreat/go-config/plugins"
)

// Step 1: Define your configuration struct with validation tags
// This is the starting point for zero reflection validation

type APIConfig struct {
	// Server configuration
	Host string `config:"host" validate:"required,regex=^[a-zA-Z0-9.-]+$"`
	Port int    `config:"port" validate:"required,range=1000:65535"`

	// Database
	DatabaseURL string `config:"database_url" validate:"required,url"`
	MaxConns    int    `config:"max_conns" validate:"min=1,max=100"`

	// Security
	APIKey    string `config:"api_key" validate:"required,len=32,alphanumeric"`
	JWTSecret string `config:"jwt_secret" validate:"required,minlen=32"`

	// Features
	LogLevel string `config:"log_level" validate:"oneof=debug|info|warn|error"`
	Version  string `config:"version" validate:"required,regex=^v[0-9]+\\.[0-9]+\\.[0-9]+$"`
}

func main() {
	fmt.Println("🏗️  Step 1: Define Configuration Struct")
	fmt.Println()

	// Load config using traditional reflection-based validation
	var config APIConfig

	err := goconfig.LoadWithPlugins(
		goconfig.Env(plugins.EnvOpts{Prefix: "API"}),
	).WithDefaults(&APIConfig{
		Host:     "localhost",
		Port:     8080,
		LogLevel: "info",
		Version:  "v1.0.0",
	}).Build(&config)

	if err != nil {
		log.Printf("Configuration error: %v", err)
		return
	}

	fmt.Printf("✅ Configuration loaded successfully!\n")
	fmt.Printf("   Host: %s\n", config.Host)
	fmt.Printf("   Port: %d\n", config.Port)
	fmt.Printf("   Version: %s\n", config.Version)
	fmt.Printf("   Log Level: %s\n", config.LogLevel)
	fmt.Println()
	fmt.Println("📋 Validation tags defined:")
	fmt.Println("   - host: required, regex pattern")
	fmt.Println("   - port: required, range 1000-65535")
	fmt.Println("   - database_url: required, valid URL")
	fmt.Println("   - max_conns: min=1, max=100")
	fmt.Println("   - api_key: required, exactly 32 alphanumeric chars")
	fmt.Println("   - jwt_secret: required, minimum 32 chars")
	fmt.Println("   - log_level: one of debug|info|warn|error")
	fmt.Println("   - version: required, semantic version format")
	fmt.Println()
	fmt.Println("▶️  Next step: Run step2_generate.go to generate validation code")
}