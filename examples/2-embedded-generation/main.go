package main

import (
	"fmt"
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

func main() {
	fmt.Println("🔄 Go Generate Validation Example")
	fmt.Println("==================================")

	// Load configuration
	config, err := loadConfig("config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	fmt.Println("📋 Validating application configuration...")
	
	// Validate main config (has generated validation)
	if err := config.Validate(); err != nil {
		fmt.Printf("❌ AppConfig validation failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✅ AppConfig validation passed")

	// Validate server config (has generated validation)
	if err := config.Server.Validate(); err != nil {
		fmt.Printf("❌ ServerConfig validation failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✅ ServerConfig validation passed")

	// Validate database config (has generated validation)
	if err := config.Database.Validate(); err != nil {
		fmt.Printf("❌ DatabaseConfig validation failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✅ DatabaseConfig validation passed")

	// Note: FeatureConfig doesn't have validation (no go:generate directive)
	fmt.Println("ℹ️  FeatureConfig has no validation (no go:generate directive)")

	fmt.Println("\n🎉 All validations passed!")
	fmt.Printf("🚀 Application '%s' (%s) ready to start on %s:%d\n",
		config.Name, config.Version, config.Server.Host, config.Server.Port)
}

func loadConfig(filename string) (*AppConfig, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config AppConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	return &config, nil
}