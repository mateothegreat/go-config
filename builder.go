package goconfig

import (
	"context"
	"log"

	"github.com/mateothegreat/go-config/plugins"
)

// BuilderSource is a builder for a source plugin.
type BuilderSource struct {
	Name string
	Opts any
}

// BuilderConfig is a builder for a config.
type BuilderConfig struct {
	sources   []BuilderSource
	defaults  any
	validator Validator
	target    any
}

// Source creates a new source builder.
func Source(name string, opts any) BuilderSource {
	return BuilderSource{Name: name, Opts: opts}
}

// PluginBuilder is a builder for loading config with plugin instances.
//
// Arguments:
//   - pluginInstances: The plugin instances to use for the config.
//
// Returns:
//   - The config builder.
type PluginBuilder struct {
	plugins   []plugins.Plugin
	defaults  any
	validator Validator
}

// LoadablePlugin represents a function that creates a plugin
type LoadablePlugin func() (plugins.Plugin, error)

// LoadWithPlugins creates a new config builder that accepts plugin constructor functions.
// It accepts functions that return (plugins.Plugin, error) like the plugin constructors.
//
// Example:
//   err := LoadWithPlugins(
//       func() (plugins.Plugin, error) { return plugins.Env(plugins.EnvOpts{}) },
//       func() (plugins.Plugin, error) { return plugins.YAML(plugins.YAMLOpts{Path: "config.yaml"}) },
//   ).Build(&config)
//
// Or use the helper functions:
//   err := LoadWithPlugins(
//       Env(plugins.EnvOpts{}),
//       YAML(plugins.YAMLOpts{Path: "config.yaml"}),
//   ).Build(&config)
//
// Arguments:
//   - pluginLoaders: The plugin constructor functions to use for the config.
//
// Returns:
//   - The config builder.
func LoadWithPlugins(pluginLoaders ...LoadablePlugin) *PluginBuilder {
	pluginList := make([]plugins.Plugin, 0, len(pluginLoaders))
	for _, loader := range pluginLoaders {
		if plugin, err := loader(); err != nil {
			log.Printf("Failed to load plugin: %v", err)
		} else {
			pluginList = append(pluginList, plugin)
		}
	}
	return &PluginBuilder{
		plugins: pluginList,
	}
}

// Helper functions for cleaner syntax

// Env creates a LoadablePlugin for environment variables
func Env(opts plugins.EnvOpts) LoadablePlugin {
	return func() (plugins.Plugin, error) {
		return plugins.Env(opts)
	}
}

// YAML creates a LoadablePlugin for YAML files
func YAML(opts plugins.YAMLOpts) LoadablePlugin {
	return func() (plugins.Plugin, error) {
		return plugins.YAML(opts)
	}
}

// Cobra creates a LoadablePlugin for Cobra commands
func Cobra(opts plugins.CobraOpts) LoadablePlugin {
	return func() (plugins.Plugin, error) {
		return plugins.Cobra(opts)
	}
}

// WithDefaults sets the defaults for the config.
//
// Arguments:
//   - def: The defaults to use for the config.
//
// Returns:
//   - The config builder.
func (pb *PluginBuilder) WithDefaults(def any) *PluginBuilder {
	pb.defaults = def
	return pb
}

// WithValidator sets the validator for the config.
//
// Arguments:
//   - v: The validator to use for the config.
//
// Returns:
//   - The config builder.
func (pb *PluginBuilder) WithValidator(v Validator) *PluginBuilder {
	pb.validator = v
	return pb
}

// WithStructValidator sets up struct tag-based validation with optional custom validators.
//
// Arguments:
//   - customValidators: The custom validators to use for the config.
//
// Returns:
//   - The config builder.
func (pb *PluginBuilder) WithStructValidator(customValidators ...func(*StructValidator)) *PluginBuilder {
	validator := NewStructValidator()
	for _, fn := range customValidators {
		fn(validator)
	}
	pb.validator = validator
	return pb
}

// Build loads the configuration into the target struct.
//
// Arguments:
//   - target: The target to build the config into.
//
// Returns:
//   - The error, if any.
func (pb *PluginBuilder) Build(target any) error {
	loader := NewConfigLoader(target)

	if pb.defaults != nil {
		loader.SetDefaults(pb.defaults)
	}

	if pb.validator != nil {
		loader.SetValidator(pb.validator)
	}

	for _, plugin := range pb.plugins {
		loader.Use(plugin)
	}

	return loader.Load(context.Background())
}

// LoadConfig creates a new config builder.
//
// Arguments:
//   - builders: The builders to use for the config.
//
// Returns:
//   - The config builder.
func LoadConfig(builders ...BuilderSource) *BuilderConfig {

	RegisterPlugin("env", plugins.Env)
	RegisterPlugin("yaml", plugins.YAML)
	RegisterPlugin("cobra", plugins.Cobra)

	return &BuilderConfig{sources: builders}
}

// WithDefaults sets the defaults for the config.
//
// Arguments:
//
//   - def: The defaults to use for the config.
//
// Returns:
//   - The config builder.
func (cb *BuilderConfig) WithDefaults(def any) *BuilderConfig {
	cb.defaults = def
	return cb
}

// WithValidator sets the validator for the config.
//
// Arguments:
//   - v: The validator to use for the config.
//
// Returns:
//   - The config builder.
func (cb *BuilderConfig) WithValidator(v Validator) *BuilderConfig {
	cb.validator = v
	return cb
}

// Build builds the config.
//
// Arguments:
//   - target: The target to build the config into.
//
// Returns:
//   - The error, if any.
func (cb *BuilderConfig) Build(target any) error {
	cb.target = target

	loader := NewConfigLoader(target)
	loader.defaults = cb.defaults
	loader.SetValidator(cb.validator)

	for _, sb := range cb.sources {
		f, err := LoadPlugin(sb.Name, sb.Opts)
		if err != nil {
			return err
		}
		loader.Use(f)
	}

	return loader.Load(context.Background())
}
