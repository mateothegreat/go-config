package plugins

import (
	"fmt"
	"sync"
)

// Registry manages plugin registration and discovery
type Registry struct {
	mu                 sync.RWMutex
	sourceFactories    map[string]PluginFactory
	validatorFactories map[string]ValidatorFactory
	sourcePlugins      map[string]Plugin
	validatorPlugins   map[string]ValidatorPlugin
}

// NewRegistry creates a new plugin registry
func NewRegistry() *Registry {
	return &Registry{
		sourceFactories:    make(map[string]PluginFactory),
		validatorFactories: make(map[string]ValidatorFactory),
		sourcePlugins:      make(map[string]Plugin),
		validatorPlugins:   make(map[string]ValidatorPlugin),
	}
}

// Global registry instance
var globalRegistry = NewRegistry()

// RegisterSourceFactory registers a configuration source factory
func RegisterSourceFactory(name string, factory PluginFactory) {
	globalRegistry.RegisterSourceFactory(name, factory)
}

// RegisterValidatorFactory registers a validator factory
func RegisterValidatorFactory(name string, factory ValidatorFactory) {
	globalRegistry.RegisterValidatorFactory(name, factory)
}

// RegisterSourcePlugin registers a source plugin instance
func RegisterSourcePlugin(name string, plugin Plugin) {
	globalRegistry.RegisterSourcePlugin(name, plugin)
}

// RegisterValidatorPlugin registers a validator plugin instance
func RegisterValidatorPlugin(name string, plugin ValidatorPlugin) {
	globalRegistry.RegisterValidatorPlugin(name, plugin)
}

// CreateSourcePlugin creates a source plugin using a registered factory
func CreateSourcePlugin(name string, opts any) (Plugin, error) {
	return globalRegistry.CreateSourcePlugin(name, opts)
}

// CreateValidatorPlugin creates a validator plugin using a registered factory
func CreateValidatorPlugin(name string) (ValidatorPlugin, error) {
	return globalRegistry.CreateValidatorPlugin(name)
}

// GetSourcePlugin retrieves a registered source plugin
func GetSourcePlugin(name string) (Plugin, bool) {
	return globalRegistry.GetSourcePlugin(name)
}

// GetValidatorPlugin retrieves a registered validator plugin
func GetValidatorPlugin(name string) (ValidatorPlugin, bool) {
	return globalRegistry.GetValidatorPlugin(name)
}

// ListSourcePlugins returns all registered source plugin names
func ListSourcePlugins() []string {
	return globalRegistry.ListSourcePlugins()
}

// ListValidatorPlugins returns all registered validator plugin names
func ListValidatorPlugins() []string {
	return globalRegistry.ListValidatorPlugins()
}

// Registry methods

func (r *Registry) RegisterSourceFactory(name string, factory PluginFactory) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.sourceFactories[name] = factory
}

func (r *Registry) RegisterValidatorFactory(name string, factory ValidatorFactory) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.validatorFactories[name] = factory
}

func (r *Registry) RegisterSourcePlugin(name string, plugin Plugin) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.sourcePlugins[name] = plugin
}

func (r *Registry) RegisterValidatorPlugin(name string, plugin ValidatorPlugin) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.validatorPlugins[name] = plugin
}

func (r *Registry) CreateSourcePlugin(name string, opts any) (Plugin, error) {
	r.mu.RLock()
	factory, exists := r.sourceFactories[name]
	r.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("source plugin factory '%s' not found", name)
	}

	return factory.Create(opts)
}

func (r *Registry) CreateValidatorPlugin(name string) (ValidatorPlugin, error) {
	r.mu.RLock()
	factory, exists := r.validatorFactories[name]
	r.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("validator plugin factory '%s' not found", name)
	}

	return factory.Create(), nil
}

func (r *Registry) GetSourcePlugin(name string) (Plugin, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	plugin, exists := r.sourcePlugins[name]
	return plugin, exists
}

func (r *Registry) GetValidatorPlugin(name string) (ValidatorPlugin, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	plugin, exists := r.validatorPlugins[name]
	return plugin, exists
}

func (r *Registry) ListSourcePlugins() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var names []string
	for name := range r.sourceFactories {
		names = append(names, name)
	}
	for name := range r.sourcePlugins {
		names = append(names, name)
	}
	return names
}

func (r *Registry) ListValidatorPlugins() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var names []string
	for name := range r.validatorFactories {
		names = append(names, name)
	}
	for name := range r.validatorPlugins {
		names = append(names, name)
	}
	return names
}

// Clear removes all registered plugins and factories
func (r *Registry) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.sourceFactories = make(map[string]PluginFactory)
	r.validatorFactories = make(map[string]ValidatorFactory)
	r.sourcePlugins = make(map[string]Plugin)
	r.validatorPlugins = make(map[string]ValidatorPlugin)
}

// GetSourceMetadata returns metadata for a source plugin
func (r *Registry) GetSourceMetadata(name string) (PluginMetadata, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if factory, exists := r.sourceFactories[name]; exists {
		return factory.Metadata(), true
	}
	return PluginMetadata{}, false
}

// GetValidatorMetadata returns metadata for a validator plugin
func (r *Registry) GetValidatorMetadata(name string) (PluginMetadata, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if factory, exists := r.validatorFactories[name]; exists {
		return factory.Metadata(), true
	}
	return PluginMetadata{}, false
}
