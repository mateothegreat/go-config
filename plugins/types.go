// Package plugins provides unified plugin interfaces for configuration sources and validators
package plugins

import "context"

// Plugin represents a configuration source plugin
type Plugin interface {
	// Name returns the plugin name
	Name() string
	// Load loads configuration data from the source
	Load(ctx context.Context) (map[string]any, error)
}

// ReloadablePlugin is a plugin that can watch for changes
type ReloadablePlugin interface {
	Plugin
	// Watch monitors for configuration changes
	Watch(ctx context.Context, onChange func()) error
}

// ValidatorPlugin represents a custom validation plugin
type ValidatorPlugin interface {
	// Name returns the validator name
	Name() string
	// Validate validates a field value
	Validate(field interface{}) error
	// SupportedTypes returns the types this validator can handle
	SupportedTypes() []string
}

// UnifiedPlugin combines configuration source and validation capabilities
type UnifiedPlugin interface {
	Plugin
	ValidatorPlugin
}

// PluginMetadata contains information about a plugin
type PluginMetadata struct {
	Name        string   `json:"name"`
	Version     string   `json:"version"`
	Description string   `json:"description"`
	Author      string   `json:"author"`
	Type        string   `json:"type"` // "source", "validator", or "unified"
	Features    []string `json:"features,omitempty"`
}

// PluginFactory creates plugin instances
type PluginFactory interface {
	// Create creates a new plugin instance with the given options
	Create(opts any) (Plugin, error)
	// Metadata returns plugin metadata
	Metadata() PluginMetadata
}

// ValidatorFactory creates validator plugin instances
type ValidatorFactory interface {
	// Create creates a new validator plugin instance
	Create() ValidatorPlugin
	// Metadata returns validator metadata
	Metadata() PluginMetadata
}
