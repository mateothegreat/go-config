package main

import (
	"fmt"
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

func main() {
	fmt.Println("🔧 Manual CLI Usage Example")
	fmt.Println("============================")
	fmt.Println()

	// Show CLI commands used to generate validation
	showCLICommands()

	// Load and validate configuration
	config, err := loadConfig("config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Validate server config
	fmt.Println("🔍 Validating server configuration...")
	if err := config.Server.Validate(); err != nil {
		fmt.Printf("❌ Server validation failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✅ Server configuration valid")

	// Validate database config
	fmt.Println("🔍 Validating database configuration...")
	if err := config.Database.Validate(); err != nil {
		fmt.Printf("❌ Database validation failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✅ Database configuration valid")

	// Validate API config
	fmt.Println("🔍 Validating API configuration...")
	if err := config.API.Validate(); err != nil {
		fmt.Printf("❌ API validation failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✅ API configuration valid")

	fmt.Println()
	fmt.Println("🎉 All configurations validated successfully!")
	fmt.Printf("🚀 Server '%s' ready on %s:%d (env: %s)\n",
		config.Server.Name, config.Server.Host, config.Server.Port, config.Server.Environment)
}

// Config represents the complete application configuration
type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	API      APIConfig      `yaml:"api"`
}

func loadConfig(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	return &config, nil
}

func showCLICommands() {
	fmt.Println("📋 CLI Commands Used for This Example:")
	fmt.Println("======================================")
	fmt.Println()

	fmt.Println("1. 👀 Preview generated code:")
	fmt.Println("   go-validate generate --dry-run --verbose")
	fmt.Println()

	fmt.Println("2. 🔄 Generate validation code:")
	fmt.Println("   go-validate generate --verbose")
	fmt.Println()

	fmt.Println("3. 🎯 Generate for specific structs:")
	fmt.Println("   go-validate generate --structs ServerConfig,DatabaseConfig")
	fmt.Println()

	fmt.Println("4. 📁 Generate to custom directory:")
	fmt.Println("   go-validate generate --output-dir ./validation --multi")
	fmt.Println()

	fmt.Println("5. 🔍 Verify generation was successful:")
	fmt.Println("   go build -o /dev/null .")
	fmt.Println()
}