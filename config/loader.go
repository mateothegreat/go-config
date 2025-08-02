package config

import (
	"context"
	"fmt"
	"sync"

	"github.com/mateothegreat/go-config/errors"
	"github.com/mateothegreat/go-config/plugins"
	"github.com/mateothegreat/go-config/validation"
)

// ConfigLoader provides a fluent interface for loading and validating configuration
type ConfigLoader struct {
	mu               sync.RWMutex
	target           any
	config           LoaderConfig
	sourceManager    SourceManager
	hydrator         Hydrator
	validator        validation.Validator
	defaultValidator validation.Validator
	loadError        *errors.ConfigLoadError
	merged           map[string]any
}

// NewConfigLoader creates a loader for a target struct
func NewConfigLoader(target any) Loader {
	config := DefaultLoaderConfig()
	return &ConfigLoader{
		target:           target,
		config:           config,
		sourceManager:    NewSourceManager(),
		hydrator:         NewHydrator(config.HydrationStrategy),
		defaultValidator: validation.NewValidator(),
		loadError:        errors.NewConfigLoadError(),
		merged:           make(map[string]any),
	}
}

// NewConfigLoaderWithConfig creates a loader with custom configuration
func NewConfigLoaderWithConfig(target any, config LoaderConfig) Loader {
	return &ConfigLoader{
		target:           target,
		config:           config,
		sourceManager:    NewSourceManager(),
		hydrator:         NewHydrator(config.HydrationStrategy),
		defaultValidator: validation.NewValidator(),
		loadError:        errors.NewConfigLoadError(),
		merged:           make(map[string]any),
	}
}

// Use adds a configuration source plugin
func (cl *ConfigLoader) Use(plugin plugins.Plugin) Loader {
	cl.mu.Lock()
	defer cl.mu.Unlock()
	cl.sourceManager.AddSource(plugin)
	return cl
}

// SetDefaults sets default values
func (cl *ConfigLoader) SetDefaults(defaults any) Loader {
	cl.mu.Lock()
	defer cl.mu.Unlock()

	// Convert defaults to map and merge
	if defaults != nil {
		defaultMap := make(map[string]any)
		if err := cl.hydrator.Hydrate(defaultMap, defaults); err == nil {
			// Reverse hydration - convert struct to map
			if defaultData, err := structToMap(defaults); err == nil {
				for k, v := range defaultData {
					if _, exists := cl.merged[k]; !exists {
						cl.merged[k] = v
					}
				}
			}
		}
	}
	return cl
}

// SetValidator sets a custom validator
func (cl *ConfigLoader) SetValidator(validator validation.Validator) Loader {
	cl.mu.Lock()
	defer cl.mu.Unlock()
	cl.validator = validator
	return cl
}

// Load processes all configuration sources and validates the result
func (cl *ConfigLoader) Load(ctx context.Context) error {
	cl.mu.Lock()
	defer cl.mu.Unlock()

	// Reset error tracker
	cl.loadError = errors.NewConfigLoadError()

	// Load from all sources
	sourceData, sourceErrors := cl.sourceManager.LoadAll(ctx)
	if sourceErrors != nil && sourceErrors.HasErrors() {
		// Add source errors
		for source, err := range sourceErrors.GetSourceErrors() {
			cl.loadError.AddSourceError(source, err)
		}

		if cl.config.FailOnSourceError {
			return cl.loadError
		}
	}

	// Merge source data
	validKeys := cl.hydrator.GetValidKeys(cl.target)
	for k, v := range sourceData {
		if validKeys[k] || cl.config.AllowUnknownFields {
			cl.merged[k] = v
		}
	}

	// Hydrate target struct
	if err := cl.hydrator.Hydrate(cl.merged, cl.target); err != nil {
		cl.loadError.AddHydrationError(err.Error())
		if cl.config.FailOnValidationError {
			return cl.loadError
		}
	}

	// Validate
	validator := cl.getValidator()
	if err := validator.Validate(cl.target); err != nil {
		// Add validation errors
		if validationErrs, ok := errors.AsValidationErrors(err); ok {
			for _, validationErr := range validationErrs {
				cl.loadError.AddHydrationError(validationErr.Error())
			}
		} else {
			// Handle other validation error types
			cl.loadError.AddHydrationError(err.Error())
		}

		if cl.config.FailOnValidationError {
			return cl.loadError
		}
	}

	// Return error only if we actually have errors
	if cl.loadError.HasErrors() {
		return cl.loadError
	}

	return nil
}

// Sources returns the names of sources used
func (cl *ConfigLoader) Sources() []string {
	cl.mu.RLock()
	defer cl.mu.RUnlock()
	return cl.sourceManager.GetSourceNames()
}

// Inspect returns merged configuration data
func (cl *ConfigLoader) Inspect() map[string]any {
	cl.mu.RLock()
	defer cl.mu.RUnlock()
	// Return a copy to prevent external modification
	result := make(map[string]any)
	for k, v := range cl.merged {
		result[k] = v
	}
	return result
}

// Errors returns the current load error if any
func (cl *ConfigLoader) Errors() error {
	cl.mu.RLock()
	defer cl.mu.RUnlock()
	if cl.loadError != nil && cl.loadError.HasErrors() {
		return cl.loadError
	}
	return nil
}

// getValidator returns the appropriate validator with auto-detection
func (cl *ConfigLoader) getValidator() validation.Validator {
	if cl.validator != nil {
		return cl.validator
	}

	// Auto-detect validation strategy based on the target struct
	detector := validation.NewValidationDetector(validation.ValidatorConfig{
		Strategy: cl.config.ValidationStrategy,
	})

	info := detector.GetValidationInfo(cl.target)

	// If generated validation is available, prioritize it
	if info.HasGeneratedCode && (cl.config.ValidationStrategy == validation.StrategyAuto || cl.config.ValidationStrategy == validation.StrategyGenerated) {
		return &GeneratedValidatorAdapter{target: cl.target}
	}

	// Configure unified validator with detected strategy
	config := validation.DefaultValidatorConfig()
	config.Strategy = info.RecommendedStrategy
	return validation.NewUnifiedValidator(config)
}

// GeneratedValidatorAdapter adapts generated validation to the Validator interface
type GeneratedValidatorAdapter struct {
	target any
}

// Validate implements the Validator interface using generated validation
func (gva *GeneratedValidatorAdapter) Validate(data any) error {
	// Check if the target implements GeneratedValidator
	if gv, ok := gva.target.(validation.GeneratedValidator); ok {
		return gv.Validate()
	}

	// Fall back to registry lookup
	return validation.ValidateWithGenerated(gva.target)
}

// SourceManager implementation

type sourceManager struct {
	mu      sync.RWMutex
	sources []plugins.Plugin
}

// NewSourceManager creates a new source manager
func NewSourceManager() SourceManager {
	return &sourceManager{
		sources: make([]plugins.Plugin, 0),
	}
}

func (sm *sourceManager) AddSource(source plugins.Plugin) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.sources = append(sm.sources, source)
}

func (sm *sourceManager) LoadAll(ctx context.Context) (map[string]any, *errors.MultiError) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	merged := make(map[string]any)
	multiErr := errors.NewMultiError()

	for _, source := range sm.sources {
		if source == nil {
			multiErr.Add("unknown", fmt.Errorf("nil configuration source"))
			continue
		}

		data, err := source.Load(ctx)
		if err != nil {
			multiErr.Add(source.Name(), err)
			continue
		}

		// Merge data (later sources override earlier ones)
		for k, v := range data {
			merged[k] = v
		}
	}

	if multiErr.HasErrors() {
		return merged, multiErr
	}
	return merged, nil
}

func (sm *sourceManager) GetSources() []plugins.Plugin {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	// Return a copy
	result := make([]plugins.Plugin, len(sm.sources))
	copy(result, sm.sources)
	return result
}

func (sm *sourceManager) GetSourceNames() []string {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	names := make([]string, 0, len(sm.sources))
	for _, source := range sm.sources {
		if source != nil {
			names = append(names, source.Name())
		}
	}
	return names
}
