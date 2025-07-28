package main

import (
	"context"
	"fmt"
	"log"

	"github.com/mateothegreat/go-config/config"
	"github.com/mateothegreat/go-config/plugins"
	"github.com/mateothegreat/go-config/plugins/sources"
	"github.com/mateothegreat/go-config/validation"
)

func main() {
	fmt.Println("🔧 CLI Usage Example - Unified Architecture")
	fmt.Println("===========================================")
	fmt.Println()

	// Show CLI commands for code generation
	showCLICommands()

	// Method 1: Load configuration using unified builder
	fmt.Println("📋 Method 1: Unified Configuration Loading")
	config1 := &Config{}

	err := config.LoadWithPlugins(
		config.FromYAML(sources.YAMLOpts{Path: "config.yaml"}),
		config.FromEnv(sources.EnvOpts{Prefix: "APP"}),
	).WithValidationStrategy(validation.StrategyAuto).Build(config1)
	if err != nil {
		fmt.Println("❌ Configuration loading/validation failed!")
		handleValidationError(err)
		return
	}

	fmt.Println("✅ Configuration loaded and validated successfully!")

	// Method 2: Individual component validation
	fmt.Println("\n📋 Method 2: Individual Component Validation")
	config2 := &Config{}

	loader := config.NewConfigLoader(config2)

	// Add YAML source
	yamlPlugin, err := plugins.CreateSourcePlugin("yaml", sources.YAMLOpts{Path: "config.yaml"})
	if err != nil {
		log.Fatalf("Failed to create YAML plugin: %v", err)
	}
	loader.Use(yamlPlugin)

	// Set defaults for missing configuration
	defaults := &Config{
		Server: ServerConfig{
			Host:        "localhost",
			Environment: "dev",
			Debug:       true,
			Metrics:     false,
		},
		Database: DatabaseConfig{
			Host:           "localhost",
			Port:           5432,
			MaxConnections: 10,
			Timeout:        30,
			SSLMode:        "disable",
		},
		API: APIConfig{
			Timeout:          30,
			RateLimit:        1000,
			RateLimitEnabled: true,
			TLSVerify:        true,
		},
	}
	loader.SetDefaults(defaults)

	if err := loader.Load(context.Background()); err != nil {
		fmt.Println("❌ Configuration loading/validation failed!")
		handleValidationError(err)
		return
	}

	// Test validation strategies for each component
	fmt.Println("🔍 Testing validation strategies...")

	// Check server config validation
	if validation.HasGeneratedValidator(&config2.Server) {
		fmt.Println("✅ ServerConfig: Using generated validation")
		if err := validation.ValidateWithGenerated(&config2.Server); err != nil {
			fmt.Printf("❌ Server validation failed: %v\n", err)
		} else {
			fmt.Println("✅ Server validation passed")
		}
	} else {
		fmt.Println("ℹ️  ServerConfig: Using reflection-based validation")
		validator := validation.NewValidator()
		if err := validator.Validate(&config2.Server); err != nil {
			fmt.Printf("❌ Server validation failed: %v\n", err)
		} else {
			fmt.Println("✅ Server validation passed")
		}
	}

	// Check database config validation
	if validation.HasGeneratedValidator(&config2.Database) {
		fmt.Println("✅ DatabaseConfig: Using generated validation")
		if err := validation.ValidateWithGenerated(&config2.Database); err != nil {
			fmt.Printf("❌ Database validation failed: %v\n", err)
		} else {
			fmt.Println("✅ Database validation passed")
		}
	} else {
		fmt.Println("ℹ️  DatabaseConfig: Using reflection-based validation")
		validator := validation.NewValidator()
		if err := validator.Validate(&config2.Database); err != nil {
			fmt.Printf("❌ Database validation failed: %v\n", err)
		} else {
			fmt.Println("✅ Database validation passed")
		}
	}

	// Check API config validation
	if validation.HasGeneratedValidator(&config2.API) {
		fmt.Println("✅ APIConfig: Using generated validation")
		if err := validation.ValidateWithGenerated(&config2.API); err != nil {
			fmt.Printf("❌ API validation failed: %v\n", err)
		} else {
			fmt.Println("✅ API validation passed")
		}
	} else {
		fmt.Println("ℹ️  APIConfig: Using reflection-based validation")
		validator := validation.NewValidator()
		if err := validator.Validate(&config2.API); err != nil {
			fmt.Printf("❌ API validation failed: %v\n", err)
		} else {
			fmt.Println("✅ API validation passed")
		}
	}

	fmt.Println()
	fmt.Println("🎉 All configurations validated successfully!")
	fmt.Printf("🚀 Server '%s' ready on %s:%d (env: %s)\n",
		config2.Server.Name, config2.Server.Host, config2.Server.Port, config2.Server.Environment)

	// Show configuration sources and final merged data
	fmt.Println("\n🔍 Configuration sources used:")
	for _, source := range loader.Sources() {
		fmt.Printf("  - %s\n", source)
	}

	fmt.Println("\n📊 Final merged configuration keys:")
	data := loader.Inspect()
	for key := range data {
		fmt.Printf("  - %s\n", key)
	}
}

// Config represents the complete application configuration
type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	API      APIConfig      `yaml:"api"`
}

// handleValidationError provides comprehensive error handling
func handleValidationError(err error) {
	// Check if it's a correlated error with suggestions
	if correlatedErr, ok := err.(*validation.CorrelatedError); ok {
		fmt.Printf("Error: %s\n", correlatedErr.Error())

		fmt.Println("\n💡 Suggestions:")
		for _, suggestion := range correlatedErr.GetSuggestions() {
			fmt.Printf("  - %s\n", suggestion)
		}
		return
	}

	// Check if it's validation errors
	if validationErrs, ok := validation.AsValidationErrors(err); ok {
		fmt.Println("Validation errors found:")
		for _, validationErr := range validationErrs.Errors() {
			fmt.Printf("  - %s\n", validationErr.Error())
		}
		return
	}

	// Generic error
	fmt.Printf("Error: %v\n", err)
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
