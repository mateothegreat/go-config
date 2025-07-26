package plugins

import (
	"context"
	"os"

	"gopkg.in/yaml.v3"
)

type YAMLOpts struct {
	Path string
}

type YAMLSource struct {
	Path string
}

func YAML(opts any) (Plugin, error) {
	o := opts.(YAMLOpts)
	return &YAMLSource{Path: o.Path}, nil
}

func (y *YAMLSource) Name() string { return "yaml" }

func (y *YAMLSource) Load(ctx context.Context) (map[string]any, error) {
	raw, err := os.ReadFile(y.Path)
	if err != nil {
		return nil, err
	}
	var data map[string]any
	err = yaml.Unmarshal(raw, &data)
	return data, err
}
