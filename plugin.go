package goconfig

import (
	"fmt"

	"github.com/mateothegreat/go-config/plugins"
)

var registry = map[string]func(opts any) (plugins.Plugin, error){}

// RegisterPlugin registers a plugin.
func RegisterPlugin(name string, factory func(opts any) (plugins.Plugin, error)) {
	registry[name] = factory
}

// LoadPlugin gets a source factory.
//
// Arguments:
//   - name: The name of the source.
//   - opts: The options for the source.
//
// Returns:
//   - The source factory.
func LoadPlugin(name string, opts any) (plugins.Plugin, error) {
	factory, ok := registry[name]
	if !ok {
		return nil, fmt.Errorf("source %q not registered", name)
	}
	return factory(opts)
}

// ListSources lists all registered sources.
//
// Returns:
//   - The list of sources.
func ListSources() []string {
	var keys []string
	for k := range registry {
		keys = append(keys, k)
	}
	return keys
}
