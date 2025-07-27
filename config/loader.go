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
	errorCorrelator  *errors.ErrorCorrelator
	merged           map[string]any
	errors           []error
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
		errorCorrelator:  errors.NewErrorCorrelator(),
		merged:           make(map[string]any),
		errors:           make([]error, 0),
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
		errorCorrelator:  errors.NewErrorCorrelator(),
		merged:           make(map[string]any),
		errors:           make([]error, 0),
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

	// Reset errors and correlator
	cl.errors = make([]error, 0)
	cl.errorCorrelator = errors.NewErrorCorrelator()

	// Load from all sources
	sourceData, sourceErrors := cl.sourceManager.LoadAll(ctx)
	if sourceErrors != nil && sourceErrors.HasErrors() {
		// Add source errors to correlator
		for source, err := range sourceErrors.GetSourceErrors() {
			cl.errorCorrelator.AddSourceError(source, err)
		}

		if cl.config.FailOnSourceError {
			return cl.createCorrelatedError()
		}
		cl.errors = append(cl.errors, sourceErrors)
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
		cl.errors = append(cl.errors, err)
		if cl.config.FailOnValidationError {
			return cl.createCorrelatedError()
		}
	}

	// Validate
	validator := cl.getValidator()
	if err := validator.Validate(cl.target); err != nil {
		// Add validation errors to correlator
		if validationErrs, ok := errors.AsValidationErrors(err); ok {
			cl.errorCorrelator.AddValidationErrors(validationErrs)
		}

		cl.errors = append(cl.errors, err)
		if cl.config.FailOnValidationError {
			return cl.createCorrelatedError()
		}
	}

	// Return correlated error if we have any errors but aren't failing fast
	if len(cl.errors) > 0 {
		return cl.createCorrelatedError()
	}

	return nil
}

// createCorrelatedError creates a correlated error from collected errors
func (cl *ConfigLoader) createCorrelatedError() error {
	correlations := cl.errorCorrelator.Correlate()
	summary := cl.errorCorrelator.GetCorrelationSummary()
	return errors.NewCorrelatedError(correlations, summary)
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

// Errors returns all accumulated errors
func (cl *ConfigLoader) Errors() []error {
	cl.mu.RLock()
	defer cl.mu.RUnlock()
	// Return a copy
	result := make([]error, len(cl.errors))
	copy(result, cl.errors)
	return result
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
