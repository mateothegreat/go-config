package sources

import (
	"context"
	"os"

	"gopkg.in/yaml.v3"
)

// YAMLOpts are the options for the YAML plugin
type YAMLOpts struct {
	Path string
}

// YAMLSource is the YAML file configuration source
type YAMLSource struct {
	Path string
}

// YAMLFactory creates YAML source plugins
type YAMLFactory struct{}

// Create creates a new YAML plugin instance
func (f *YAMLFactory) Create(opts any) (Plugin, error) {
	o := opts.(YAMLOpts)
	return &YAMLSource{Path: o.Path}, nil
}

// Metadata returns YAML plugin metadata
func (f *YAMLFactory) Metadata() PluginMetadata {
	return PluginMetadata{
		Name:        "yaml",
		Version:     "1.0.0",
		Description: "YAML file configuration source",
		Author:      "go-config",
		Type:        "source",
		Features:    []string{"file-based", "structured"},
	}
}

// Name returns the plugin name
func (y *YAMLSource) Name() string {
	return "yaml"
}

// Load loads configuration from a YAML file
func (y *YAMLSource) Load(ctx context.Context) (map[string]any, error) {
	raw, err := os.ReadFile(y.Path)
	if err != nil {
		return nil, err
	}
	var data map[string]any
	err = yaml.Unmarshal(raw, &data)
	return data, err
}
