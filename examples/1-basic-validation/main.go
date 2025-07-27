package main

import (
	"fmt"
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

func main() {
	fmt.Println("🚀 Basic Validation Example")
	fmt.Println("============================")

	// Load configuration from YAML file
	config, err := LoadConfigFromFile("config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	fmt.Printf("Loaded config: %+v\n\n", config)

	// Validate the configuration
	fmt.Println("🔍 Validating configuration...")
	err = config.Validate()
	if err != nil {
		fmt.Println("❌ Validation failed!")

		// Handle validation error - in this basic example, just show the error message
		fmt.Printf("Error: %v\n", err)

		fmt.Println("\n💡 Try fixing the config.yaml file and run again!")
		os.Exit(1)
	}

	fmt.Println("✅ Configuration is valid!")
	fmt.Printf("🎯 Server '%s' will run on %s:%d (log level: %s)\n",
		config.Name, config.Host, config.Port, config.LogLevel)
}

// LoadConfigFromFile demonstrates loading and parsing a YAML configuration file
func LoadConfigFromFile(filename string) (*ServerConfig, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config ServerConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	return &config, nil
}
