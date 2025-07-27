// Package config provides configuration loading and hydration
package config

import (
	"context"

	"github.com/mateothegreat/go-config/errors"
	"github.com/mateothegreat/go-config/plugins"
	"github.com/mateothegreat/go-config/validate"
)

// Loader provides the main interface for configuration loading
type Loader interface {
	// Use adds a configuration source plugin
	Use(plugin plugins.Plugin) Loader
	// SetDefaults sets default values
	SetDefaults(defaults any) Loader
	// SetValidator sets a custom validator
	SetValidator(validator validate.Validator) Loader
	// Load processes all sources and validates the result
	Load(ctx context.Context) error
	// Sources returns the names of sources used
	Sources() []string
	// Inspect returns merged configuration data
	Inspect() map[string]any
	// Errors returns all accumulated errors
	Errors() []error
}

// Builder provides a fluent interface for configuration building
type Builder interface {
	// WithDefaults sets default values
	WithDefaults(defaults any) Builder
	// WithValidator sets a custom validator
	WithValidator(validator validate.Validator) Builder
	// WithValidationStrategy sets the validation strategy
	WithValidationStrategy(strategy validate.ValidationStrategy) Builder
	// Build loads configuration into the target struct
	Build(target any) error
}

// HydrationStrategy determines how configuration data is mapped to structs
type HydrationStrategy int

const (
	// HydrationAuto automatically detects the best hydration method
	HydrationAuto HydrationStrategy = iota
	// HydrationReflection uses reflection-based field mapping
	HydrationReflection
	// HydrationTags uses struct tags for field mapping
	HydrationTags
)

// LoaderConfig configures loader behavior
type LoaderConfig struct {
	HydrationStrategy     HydrationStrategy
	ValidationStrategy    validate.ValidationStrategy
	FailOnSourceError     bool
	FailOnValidationError bool
	AllowUnknownFields    bool
	CaseSensitive         bool
}

// DefaultLoaderConfig returns the default loader configuration
func DefaultLoaderConfig() LoaderConfig {
	return LoaderConfig{
		HydrationStrategy:     HydrationAuto,
		ValidationStrategy:    validate.StrategyAuto,
		FailOnSourceError:     true,
		FailOnValidationError: true,
		AllowUnknownFields:    true,
		CaseSensitive:         false,
	}
}

// Hydrator handles the conversion of raw configuration data to struct fields
type Hydrator interface {
	// Hydrate converts raw data into the target struct
	Hydrate(data map[string]any, target any) error
	// GetValidKeys returns the valid configuration keys for a target struct
	GetValidKeys(target any) map[string]bool
}

// SourceManager manages configuration source plugins
type SourceManager interface {
	// AddSource adds a configuration source
	AddSource(source plugins.Plugin)
	// LoadAll loads configuration from all sources
	LoadAll(ctx context.Context) (map[string]any, *errors.MultiError)
	// GetSources returns all registered sources
	GetSources() []plugins.Plugin
	// GetSourceNames returns the names of all sources
	GetSourceNames() []string
}
