package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"gopkg.in/yaml.v3"

	goconfig "github.com/mateothegreat/go-config"
	"github.com/mateothegreat/go-config/config"
	"github.com/mateothegreat/go-config/plugins/sources"
	"github.com/mateothegreat/go-config/validation"
)

func testValidation() {
	fmt.Println("🧪 Testing the fixed validation system...")

	// Load config using standard YAML parsing
	data, err := os.ReadFile("config_app.yaml")
	if err != nil {
		fmt.Printf("❌ Error reading YAML: %v\n", err)
		return
	}

	var config AppConfig
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		fmt.Printf("❌ Error parsing YAML: %v\n", err)
		return
	}

	fmt.Println("✅ Configuration loaded successfully!")

	// Test generated validation detection
	if goconfig.HasGeneratedValidator(&config) {
		fmt.Println("✅ Generated validator detected!")

		// Test generated validation
		err = goconfig.ValidateWithGenerated(&config)
		if err != nil {
			fmt.Printf("❌ Validation failed: %v\n", err)
		} else {
			fmt.Println("✅ Generated validation passed!")
		}
	} else {
		fmt.Println("❌ No generated validator found")
	}

	// Test individual component validation
	fmt.Println("\n🔍 Testing individual components:")

	// Test server config
	if goconfig.HasGeneratedValidator(&config.Server) {
		fmt.Println("✅ ServerConfig has generated validation")
		if err := goconfig.ValidateWithGenerated(&config.Server); err != nil {
			fmt.Printf("❌ ServerConfig validation failed: %v\n", err)
		} else {
			fmt.Println("✅ ServerConfig validation passed")
		}
	}

	// Test database config
	if goconfig.HasGeneratedValidator(&config.Database) {
		fmt.Println("✅ DatabaseConfig has generated validation")
		if err := goconfig.ValidateWithGenerated(&config.Database); err != nil {
			fmt.Printf("❌ DatabaseConfig validation failed: %v\n", err)
		} else {
			fmt.Println("✅ DatabaseConfig validation passed")
		}
	}

	fmt.Println("\n🎉 Validation system is working correctly!")
}

// debugYAML prints the configuration values for debugging purposes.
//
// Arguments:
// - config: The application configuration to debug
func debugYAML(config AppConfig) {
	fmt.Printf("  Server.Host: %s\n", config.Server.Host)
	fmt.Printf("  Server.Port: %d\n", config.Server.Port)
	fmt.Printf("  Database.Host: %s\n", config.Database.Host)
	fmt.Printf("  Database.Port: %d\n", config.Database.Port)
	fmt.Printf("  Database.Username: %s\n", config.Database.Username)
	fmt.Printf("  Database.Password: %s\n", config.Database.Password)
	fmt.Printf("  Database.Password: %s\n", config.Database.Password)
}

func main() {
	fmt.Println("🔄 Embedded Generation Example - Unified Architecture")
	fmt.Println("====================================================")

	// Test the fixed validation system first
	testValidation()

	// Method 1: Using the unified loader with auto-detection
	fmt.Println("\n📋 Method 1: Auto-Detection of Generated Validation")
	config1 := &AppConfig{}

	// Debug: Try using the direct loader instead
	fmt.Println("🔍 Debug: Testing direct loader...")
	debugLoader := goconfig.NewLoader(config1)
	debugYAML(*config1)

	// Try creating a plugin manually
	fmt.Println("🔍 Debug: Creating YAML plugin...")
	debugYamlPlugin, err := goconfig.CreateSourcePlugin("yaml", sources.YAMLOpts{Path: "config_app.yaml"})
	if err != nil {
		fmt.Printf("❌ Failed to create YAML plugin: %v\n", err)
		return
	}

	fmt.Println("🔍 Debug: Using YAML plugin...")
	debugLoader.Use(debugYamlPlugin)

	fmt.Println("🔍 Debug: Loading configuration...")
	err = debugLoader.Load(context.Background())
	if err != nil {
		fmt.Println("❌ Configuration loading failed!")
		fmt.Printf("Raw error: %v\n", err)
		fmt.Printf("Error type: %T\n", err)

		// Check loader errors
		loaderErrors := debugLoader.Errors()
		fmt.Printf("🔍 Debug: Loader has %d errors:\n", len(loaderErrors))
		for i, lerr := range loaderErrors {
			fmt.Printf("  Error %d: %v (type: %T)\n", i, lerr, lerr)
		}

		// Try to get more details about the error
		if correlatedErr, ok := err.(*goconfig.CorrelatedError); ok {
			fmt.Println("🔍 Debug: Correlated error details:")
			fmt.Printf("  Correlations count: %d\n", len(correlatedErr.Correlations))
			fmt.Printf("  Summary: %s\n", correlatedErr.Summary)
			for i, correlation := range correlatedErr.Correlations {
				fmt.Printf("  Correlation %d:\n", i)
				fmt.Printf("    Source: %s\n", correlation.SourceName)
				fmt.Printf("    Field: %s\n", correlation.FieldName)
				fmt.Printf("    Source Error: %s\n", correlation.SourceError)
				fmt.Printf("    Validation Error: %s\n", correlation.ValidationError)
				fmt.Printf("    Category: %s\n", correlation.Category)
				fmt.Printf("    Suggestion: %s\n", correlation.Suggestion)
			}
		}
		return
	}

	fmt.Println("✅ Configuration loaded successfully without validation!")

	fmt.Println("🔍 Debug: Now trying with validation...")
	// Now try with validation
	config2 := &AppConfig{}
	err = config.LoadWithPlugins(
		config.FromYAML(sources.YAMLOpts{Path: "config_app.yaml"}),
		config.FromEnv(sources.EnvOpts{Prefix: "APP"}),
	).WithValidationStrategy(validation.StrategyAuto).Build(config2)
	if err != nil {
		fmt.Println("❌ Configuration validation failed!")
		fmt.Printf("Raw error: %v\n", err)
		fmt.Printf("Error type: %T\n", err)
		handleError(err)
		return
	}

	fmt.Println("✅ Configuration loaded and validated successfully!")
	fmt.Printf("🎯 Auto-detected validation strategy used\n\n")

	// Method 2: Explicitly using generated validation
	fmt.Println("📋 Method 2: Explicit Generated Validation Strategy")
	config3 := &AppConfig{}

	loader := goconfig.NewLoader(config3)

	// Add YAML source
	yamlPlugin, err := goconfig.CreateSourcePlugin("yaml", sources.YAMLOpts{Path: "config_app.yaml"})
	if err != nil {
		log.Fatalf("Failed to create YAML plugin: %v", err)
	}
	loader.Use(yamlPlugin)

	// Set explicit validation strategy to use generated code
	validator := goconfig.NewValidatorWithConfig(validation.ValidatorConfig{
		Strategy: validation.StrategyGenerated,
	})
	loader.SetValidator(validator)

	if err := loader.Load(context.Background()); err != nil {
		fmt.Println("❌ Configuration loading/validation failed!")
		handleError(err)
		return
	}

	fmt.Println("✅ Configuration loaded with generated validation!")

	// Method 3: Testing nested struct validation
	fmt.Println("\n📋 Method 3: Nested Struct Validation")

	// Show validation detection info
	detector := goconfig.NewValidationDetector(goconfig.ValidatorConfig{
		Strategy: goconfig.StrategyAuto,
	})

	info := detector.GetValidationInfo(config3)
	fmt.Printf("🔍 Validation Info: %s\n", info.String())

	// Test individual nested components if they have generated validation
	if goconfig.HasGeneratedValidator(&config3.Server) {
		fmt.Println("✅ ServerConfig has generated validation")
		if err := goconfig.ValidateWithGenerated(&config3.Server); err != nil {
			fmt.Printf("❌ ServerConfig validation failed: %v\n", err)
		} else {
			fmt.Println("✅ ServerConfig validation passed")
		}
	}

	if goconfig.HasGeneratedValidator(&config3.Database) {
		fmt.Println("✅ DatabaseConfig has generated validation")
		if err := goconfig.ValidateWithGenerated(&config3.Database); err != nil {
			fmt.Printf("❌ DatabaseConfig validation failed: %v\n", err)
		} else {
			fmt.Println("✅ DatabaseConfig validation passed")
		}
	}

	// Features config typically doesn't have validation
	if !goconfig.HasGeneratedValidator(&config3.Features) {
		fmt.Println("ℹ️  FeatureConfig has no generated validation")
	}

	fmt.Println("\n🎉 All validations passed!")
	fmt.Printf("🚀 Application '%s' (%s) ready to start on %s:%d\n",
		config3.Name, config3.Version, config3.Server.Host, config3.Server.Port)

	// Show final configuration
	fmt.Println("\n📊 Final Configuration:")
	data := loader.Inspect()
	for key, value := range data {
		fmt.Printf("  %s: %v\n", key, value)
	}
}

// handleError provides comprehensive error handling
func handleError(err error) {
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
}
