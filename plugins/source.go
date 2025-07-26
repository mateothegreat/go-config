package plugins

import "context"

// Plugin is a plugin.
type Plugin interface {
	Load(ctx context.Context) (map[string]any, error)
	Name() string
}

// ReloadablePlugin is a plugin that can be reloaded.
type ReloadablePlugin interface {
	Plugin
	Watch(ctx context.Context, onChange func())
}
