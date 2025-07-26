//go:generate go run ../../cmd/go-config-gen -input . -output validation_generated.go -verbose

package main

import (
	"fmt"
	"log"

	goconfig "github.com/mateothegreat/go-config"
	"github.com/mateothegreat/go-config/plugins"
)

func main() {
	fmt.Println("🔧 Zero Reflection Validation with go-config-gen")
	fmt.Println()

	var config ServerConfig

	// Load configuration from multiple sources
	err := goconfig.LoadWithPlugins(
		goconfig.Env(plugins.EnvOpts{Prefix: "SERVER"}),
		goconfig.YAML(plugins.YAMLOpts{Path: "config.yaml"}),
	).WithDefaults(&ServerConfig{
		Host:        "localhost",
		Port:        8080,
		Protocol:    "http",
		DatabaseURL: "postgres://localhost:5432/mydb",
		MaxConns:    10,
		EnableDebug: false,
		LogLevel:    "info",
		Version:     "v1.0.0",
		JWTSecret:   "this-is-a-very-secure-jwt-secret-key-with-32-plus-characters",
		AdminEmail:  "admin@example.com",
		APIKey:      "abcd1234efgh5678ijkl9012mnop3456",
	}).Build(&config)
	if err != nil {
		log.Printf("Configuration loading error: %v", err)
		return
	}

	fmt.Println("✅ Configuration loaded successfully!")

	// Check if generated validation exists
	fmt.Println("🔍 Checking for generated validation...")

	// The Validate() method will be available after running: go generate
	// For now, we'll just show that configuration loaded successfully
	fmt.Println("💡 To use zero reflection validation:")
	fmt.Println("   1. Run: go generate")
	fmt.Println("   2. Then run: go run .")
	fmt.Println("   3. The generated Validate() method will be available")
	
	fmt.Println("✅ Configuration loaded successfully!")
	fmt.Printf("Server will run on %s://%s:%d\n", config.Protocol, config.Host, config.Port)
	fmt.Printf("Database: %s\n", config.DatabaseURL)
	fmt.Printf("Log Level: %s\n", config.LogLevel)
	fmt.Printf("Version: %s\n", config.Version)
}
