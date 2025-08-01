package config

import (
	"context"

	"github.com/mateothegreat/go-config/plugins"
	"github.com/mateothegreat/go-config/plugins/sources"
	"github.com/mateothegreat/go-config/validation"
	"github.com/mateothegreat/go-multilog/multilog"
)

// FluentBuilder provides a fluent interface for configuration building
type FluentBuilder struct {
	loader        Loader
	config        LoaderConfig
	validator     validation.Validator
	pluginLoaders []PluginLoader
}

// PluginLoader represents a function that creates a plugin
type PluginLoader func() (plugins.Plugin, error)

// NewBuilder creates a new fluent configuration builder
func NewBuilder() Builder {
	config := DefaultLoaderConfig()
	return &FluentBuilder{
		config:        config,
		pluginLoaders: make([]PluginLoader, 0),
	}
}

// NewBuilderWithConfig creates a builder with custom configuration
func NewBuilderWithConfig(config LoaderConfig) Builder {
	return &FluentBuilder{
		config:        config,
		pluginLoaders: make([]PluginLoader, 0),
	}
}

// WithDefaults sets default values
func (fb *FluentBuilder) WithDefaults(defaults any) Builder {
	// Store defaults for later use in Build
	fb.config.HydrationStrategy = HydrationAuto
	return fb
}

// WithValidator sets a custom validator
func (fb *FluentBuilder) WithValidator(validator validation.Validator) Builder {
	fb.validator = validator
	return fb
}

// WithValidationStrategy sets the validation strategy
func (fb *FluentBuilder) WithValidationStrategy(strategy validation.Strategy) Builder {
	fb.config.ValidationStrategy = strategy
	return fb
}

// Build loads configuration into the target struct
func (fb *FluentBuilder) Build(target any) error {
	// Create loader with target
	fb.loader = NewConfigLoaderWithConfig(target, fb.config)

	// Add all plugins
	for _, pluginLoader := range fb.pluginLoaders {
		if plugin, err := pluginLoader(); err == nil {
			fb.loader.Use(plugin)
		}
		// Note: Errors are handled by individual plugin loaders
	}

	// Set validator if provided
	if fb.validator != nil {
		fb.loader.SetValidator(fb.validator)
	}
	// Note: If no validator is set, the loader will auto-detect the appropriate validator

	return fb.loader.Load(context.Background())
}

// Plugin creation helpers

// FromEnv creates a plugin loader for environment variables
func FromEnv(opts sources.EnvOpts) PluginLoader {
	return func() (plugins.Plugin, error) {
		factory := &sources.EnvFactory{}
		return factory.Create(opts)
	}
}

// FromYAML creates a plugin loader for YAML files
func FromYAML(opts sources.YAMLOpts) PluginLoader {
	return func() (plugins.Plugin, error) {
		factory := &sources.YAMLFactory{}
		return factory.Create(opts)
	}
}

// FromPlugin creates a plugin loader from a plugin instance
func FromPlugin(plugin plugins.Plugin) PluginLoader {
	return func() (plugins.Plugin, error) {
		return plugin, nil
	}
}

// PluginBuilder provides a fluent interface for loading config with multiple plugins
type PluginBuilder struct {
	pluginLoaders []PluginLoader
	config        LoaderConfig
	defaults      any
	validator     validation.Validator
}

// LoadWithPlugins creates a new config builder that accepts plugin constructor functions
func LoadWithPlugins(pluginLoaders ...PluginLoader) *PluginBuilder {
	return &PluginBuilder{
		pluginLoaders: pluginLoaders,
		config:        DefaultLoaderConfig(),
	}
}

// WithDefaults sets the defaults for the config
func (pb *PluginBuilder) WithDefaults(def any) *PluginBuilder {
	pb.defaults = def
	return pb
}

// WithValidator sets the validator for the config
func (pb *PluginBuilder) WithValidator(v validation.Validator) *PluginBuilder {
	pb.validator = v
	return pb
}

// WithValidationStrategy sets the validation strategy
func (pb *PluginBuilder) WithValidationStrategy(strategy validation.Strategy) *PluginBuilder {
	pb.config.ValidationStrategy = strategy
	return pb
}

// WithStructValidator sets up struct tag-based validation with optional custom validators
func (pb *PluginBuilder) WithStructValidator(customValidators ...func(*validation.UnifiedValidator)) *PluginBuilder {
	config := validation.DefaultValidatorConfig()
	validator := validation.NewUnifiedValidator(config)
	for _, fn := range customValidators {
		fn(validator)
	}
	pb.validator = validator
	return pb
}

// Build loads configuration into the target struct using auto-detection
func (pb *PluginBuilder) Build(target any) error {
	loader := NewConfigLoaderWithConfig(target, pb.config)

	// Set defaults
	if pb.defaults != nil {
		loader.SetDefaults(pb.defaults)
	}

	multilog.Debug("go-config.builder.Build", "building config loader", map[string]any{
		"defaults":  pb.defaults,
		"target":    target,
		"validator": pb.validator,
	})

	// Auto-detect validation strategy if no validator is set
	if pb.validator == nil {
		detector := validation.NewValidationDetector(validation.ValidatorConfig{
			Strategy: pb.config.ValidationStrategy,
		})

		info := detector.GetValidationInfo(target)

		if info.HasGeneratedCode {
			// Use generated validation - let the loader handle this automatically
			// The loader will auto-detect and create the appropriate adapter
			pb.validator = nil
		} else {
			// Use unified validator with auto-detection
			validatorConfig := validation.DefaultValidatorConfig()
			validatorConfig.Strategy = pb.config.ValidationStrategy
			pb.validator = validation.NewUnifiedValidator(validatorConfig)
		}
	}

	if pb.validator != nil {
		loader.SetValidator(pb.validator)
	}
	// If pb.validator is nil, let the loader auto-detect the appropriate validator

	// Add all plugins
	for _, pluginLoader := range pb.pluginLoaders {
		if plugin, err := pluginLoader(); err == nil {
			loader.Use(plugin)
			multilog.Debug("go-config.builder.Build", "loaded plugin", map[string]any{
				"name":   plugin.Name(),
				"plugin": plugin,
			})
		} else {
			multilog.Debug("go-config.builder.Build", "failed to add plugin", map[string]any{
				"name":  plugin.Name(),
				"error": err,
			})
		}
	}

	err := loader.Load(context.Background())
	if err != nil {
		multilog.Debug("go-config.builder.Build", "failed to load config", map[string]any{
			"error": err,
		})
		return err
	}

	multilog.Debug("go-config.builder.Build", "loaded config", map[string]any{
		"target": target,
	})

	return nil
}

// Legacy builder support for backwards compatibility

// BuilderSource represents a legacy builder source
type BuilderSource struct {
	Name string
	Opts any
}

// BuilderConfig represents a legacy builder config
type BuilderConfig struct {
	sources   []BuilderSource
	defaults  any
	validator validation.Validator
}

// Source creates a new source builder (legacy)
func Source(name string, opts any) BuilderSource {
	return BuilderSource{Name: name, Opts: opts}
}

// LoadConfig creates a new config builder (legacy)
func LoadConfig(builders ...BuilderSource) *BuilderConfig {
	return &BuilderConfig{sources: builders}
}

// WithDefaults sets the defaults for the config (legacy)
func (cb *BuilderConfig) WithDefaults(def any) *BuilderConfig {
	cb.defaults = def
	return cb
}

// WithValidator sets the validator for the config (legacy)
func (cb *BuilderConfig) WithValidator(v validation.Validator) *BuilderConfig {
	cb.validator = v
	return cb
}

// Build builds the config (legacy)
func (cb *BuilderConfig) Build(target any) error {
	loader := NewConfigLoader(target)

	if cb.defaults != nil {
		loader.SetDefaults(cb.defaults)
	}

	if cb.validator != nil {
		loader.SetValidator(cb.validator)
	}

	// Convert legacy sources to plugins
	for _, sb := range cb.sources {
		switch sb.Name {
		case "env":
			if opts, ok := sb.Opts.(sources.EnvOpts); ok {
				if plugin, err := FromEnv(opts)(); err == nil {
					loader.Use(plugin)
				}
			}
		case "yaml":
			if opts, ok := sb.Opts.(sources.YAMLOpts); ok {
				if plugin, err := FromYAML(opts)(); err == nil {
					loader.Use(plugin)
				}
			}
		}
	}

	return loader.Load(context.Background())
}
