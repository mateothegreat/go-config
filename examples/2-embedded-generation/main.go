package main

import (
	"context"
	"fmt"
	"log"

	goconfig "github.com/mateothegreat/go-config"
	"github.com/mateothegreat/go-config/plugins/sources"
)

func main() {
	fmt.Println("🔄 Embedded Generation Example - Unified Architecture")
	fmt.Println("====================================================")

	// Method 1: Using the unified loader with auto-detection
	fmt.Println("📋 Method 1: Auto-Detection of Generated Validation")
	config1 := &AppConfig{}

	err := goconfig.LoadWithPlugins(
		goconfig.FromYAML(sources.YAMLOpts{Path: "config_app.yaml"}),
		goconfig.FromEnv(sources.EnvOpts{Prefix: "APP"}),
	).WithValidationStrategy(goconfig.StrategyAuto).Build(config1)

	if err != nil {
		fmt.Println("❌ Configuration loading/validation failed!")
		handleError(err)
		return
	}

	fmt.Println("✅ Configuration loaded and validated successfully!")
	fmt.Printf("🎯 Auto-detected validation strategy used\n\n")

	// Method 2: Explicitly using generated validation
	fmt.Println("📋 Method 2: Explicit Generated Validation Strategy")
	config2 := &AppConfig{}

	loader := goconfig.NewLoader(config2)

	// Add YAML source
	yamlPlugin, err := goconfig.CreateSourcePlugin("yaml", sources.YAMLOpts{Path: "config_app.yaml"})
	if err != nil {
		log.Fatalf("Failed to create YAML plugin: %v", err)
	}
	loader.Use(yamlPlugin)

	// Set explicit validation strategy to use generated code
	validator := goconfig.NewValidatorWithConfig(goconfig.ValidatorConfig{
		Strategy: goconfig.StrategyGenerated,
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

	info := detector.GetValidationInfo(config2)
	fmt.Printf("🔍 Validation Info: %s\n", info.String())

	// Test individual nested components if they have generated validation
	if goconfig.HasGeneratedValidator(&config2.Server) {
		fmt.Println("✅ ServerConfig has generated validation")
		if err := goconfig.ValidateWithGenerated(&config2.Server); err != nil {
			fmt.Printf("❌ ServerConfig validation failed: %v\n", err)
		} else {
			fmt.Println("✅ ServerConfig validation passed")
		}
	}

	if goconfig.HasGeneratedValidator(&config2.Database) {
		fmt.Println("✅ DatabaseConfig has generated validation")
		if err := goconfig.ValidateWithGenerated(&config2.Database); err != nil {
			fmt.Printf("❌ DatabaseConfig validation failed: %v\n", err)
		} else {
			fmt.Println("✅ DatabaseConfig validation passed")
		}
	}

	// Features config typically doesn't have validation
	if !goconfig.HasGeneratedValidator(&config2.Features) {
		fmt.Println("ℹ️  FeatureConfig has no generated validation")
	}

	fmt.Println("\n🎉 All validations passed!")
	fmt.Printf("🚀 Application '%s' (%s) ready to start on %s:%d\n",
		config2.Name, config2.Version, config2.Server.Host, config2.Server.Port)

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
