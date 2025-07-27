// Package sources provides built-in configuration source plugins
package sources

import (
	"context"
	"os"
	"strconv"
	"strings"
)

// EnvOpts are the options for the env plugin
type EnvOpts struct {
	Prefix string
}

// EnvSource is the environment variable configuration source
type EnvSource struct {
	Prefix string
}

// Using local interfaces to avoid import cycles - these will be bridged by adapters
type Plugin interface {
	Name() string
	Load(ctx context.Context) (map[string]any, error)
}

type PluginMetadata struct {
	Name        string   `json:"name"`
	Version     string   `json:"version"`
	Description string   `json:"description"`
	Author      string   `json:"author"`
	Type        string   `json:"type"`
	Features    []string `json:"features,omitempty"`
}

type PluginFactory interface {
	Create(opts any) (Plugin, error)
	Metadata() PluginMetadata
}

// EnvFactory creates environment variable source plugins
type EnvFactory struct{}

// Create creates a new env plugin instance
func (f *EnvFactory) Create(opts any) (Plugin, error) {
	o := opts.(EnvOpts)
	return &EnvSource{Prefix: o.Prefix}, nil
}

// Metadata returns env plugin metadata
func (f *EnvFactory) Metadata() PluginMetadata {
	return PluginMetadata{
		Name:        "env",
		Version:     "1.0.0",
		Description: "Environment variable configuration source",
		Author:      "go-config",
		Type:        "source",
		Features:    []string{"dynamic", "system-integration"},
	}
}

// Name returns the plugin name
func (e *EnvSource) Name() string {
	return "env"
}

// Load loads configuration from environment variables
func (e *EnvSource) Load(ctx context.Context) (map[string]any, error) {
	result := make(map[string]any)
	prefix := strings.ToUpper(e.Prefix)
	if prefix != "" && !strings.HasSuffix(prefix, "_") {
		prefix = prefix + "_"
	}

	for _, kv := range os.Environ() {
		parts := strings.SplitN(kv, "=", 2)
		if len(parts) == 2 {
			key := parts[0]
			value := parts[1]

			// Filter by prefix if specified
			if e.Prefix != "" {
				if !strings.HasPrefix(key, prefix) {
					continue
				}
				// Remove the prefix and underscore
				key = strings.TrimPrefix(key, prefix)
			}

			// Convert to lowercase to match config tags
			key = strings.ToLower(key)
			result[key] = inferType(value)
		}
	}
	return result, nil
}

// inferType attempts to infer the type of a string value
func inferType(value string) any {
	// Try bool first
	if boolVal, err := strconv.ParseBool(value); err == nil {
		return boolVal
	}

	// Try int
	if intVal, err := strconv.ParseInt(value, 10, 64); err == nil {
		return intVal
	}

	// Try float
	if floatVal, err := strconv.ParseFloat(value, 64); err == nil {
		return floatVal
	}

	// Default to string
	return value
}
