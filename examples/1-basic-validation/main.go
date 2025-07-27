package main

import (
	"context"
	"fmt"
	"log"

	goconfig "github.com/mateothegreat/go-config"
	"github.com/mateothegreat/go-config/plugins/sources"
)

func main() {
	fmt.Println("🚀 Basic Validation Example - Unified Architecture")
	fmt.Println("==================================================")

	// Method 1: Using the fluent builder API
	fmt.Println("📋 Method 1: Using Fluent Builder API")
	config1 := &ServerConfig{}

	err := goconfig.LoadWithPlugins(
		goconfig.FromYAML(sources.YAMLOpts{Path: "config.yaml"}),
		goconfig.FromEnv(sources.EnvOpts{Prefix: "SERVER"}),
	).WithValidationStrategy(goconfig.StrategyAuto).Build(config1)

	if err != nil {
		fmt.Println("❌ Validation failed!")
		handleValidationError(err)
		return
	}

	fmt.Println("✅ Configuration loaded and validated successfully!")
	fmt.Printf("🎯 Server '%s' will run on %s:%d (log level: %s)\n\n",
		config1.Name, config1.Host, config1.Port, config1.LogLevel)

	// Method 2: Using the loader API for more control
	fmt.Println("📋 Method 2: Using Loader API with Custom Validation")
	config2 := &ServerConfig{}

	loader := goconfig.NewLoader(config2)

	// Add YAML source
	yamlPlugin, err := goconfig.CreateSourcePlugin("yaml", sources.YAMLOpts{Path: "config.yaml"})
	if err != nil {
		log.Fatalf("Failed to create YAML plugin: %v", err)
	}
	loader.Use(yamlPlugin)

	// Add environment variable source (higher priority - overrides YAML)
	envPlugin, err := goconfig.CreateSourcePlugin("env", sources.EnvOpts{Prefix: "SERVER"})
	if err != nil {
		log.Fatalf("Failed to create ENV plugin: %v", err)
	}
	loader.Use(envPlugin)

	// Set defaults
	defaults := &ServerConfig{
		Host:     "localhost",
		Port:     8080,
		LogLevel: "info",
		Debug:    false,
	}
	loader.SetDefaults(defaults)

	// Load and validate
	if err := loader.Load(context.Background()); err != nil {
		fmt.Println("❌ Configuration loading/validation failed!")
		handleValidationError(err)
		return
	}

	fmt.Println("✅ Configuration loaded and validated successfully!")
	fmt.Printf("🎯 Server '%s' will run on %s:%d (log level: %s)\n",
		config2.Name, config2.Host, config2.Port, config2.LogLevel)

	// Show inspection data
	fmt.Println("\n🔍 Configuration sources used:")
	for _, source := range loader.Sources() {
		fmt.Printf("  - %s\n", source)
	}

	fmt.Println("\n📊 Final merged configuration:")
	data := loader.Inspect()
	for key, value := range data {
		fmt.Printf("  %s: %v\n", key, value)
	}
}

// handleValidationError provides detailed error handling
func handleValidationError(err error) {
	// Check if it's a correlated error with suggestions
	if correlatedErr, ok := err.(*goconfig.CorrelatedError); ok {
		fmt.Printf("Error: %s\n", correlatedErr.Error())

		fmt.Println("\n💡 Suggestions:")
		for _, suggestion := range correlatedErr.GetSuggestions() {
			fmt.Printf("  - %s\n", suggestion)
		}
		return
	}

	// Check if it's validation errors
	if validationErrs, ok := goconfig.AsValidationErrors(err); ok {
		fmt.Println("Validation errors found:")
		for _, validationErr := range validationErrs.Errors() {
			fmt.Printf("  - %s\n", validationErr.Error())
		}
		return
	}

	// Generic error
	fmt.Printf("Error: %v\n", err)
	fmt.Println("\n💡 Try fixing the config.yaml file and run again!")
}
