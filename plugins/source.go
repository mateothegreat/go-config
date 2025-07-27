package plugins

import "context"

// LegacyPlugin is a legacy plugin interface (deprecated - use Plugin from types.go)
type LegacyPlugin interface {
	Load(ctx context.Context) (map[string]any, error)
	Name() string
}

// LegacyReloadablePlugin is a legacy reloadable plugin interface
type LegacyReloadablePlugin interface {
	LegacyPlugin
	Watch(ctx context.Context, onChange func())
}
