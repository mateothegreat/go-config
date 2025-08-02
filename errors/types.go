// Package errors provides shared error types for configuration and validation
package errors

import (
	"fmt"
	"strings"
)

// ValidationError represents a single validation error with rich context
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Value   any    `json:"value,omitempty"`
	Code    string `json:"code,omitempty"`
}

// Error implements the error interface
func (ve ValidationError) Error() string {
	return fmt.Sprintf("field '%s': %s", ve.Field, ve.Message)
}

// ValidationErrors represents a collection of validation errors
type ValidationErrors []ValidationError

// Error implements the error interface
func (ve ValidationErrors) Error() string {
	if len(ve) == 0 {
		return ""
	}

	var messages []string
	for _, err := range ve {
		messages = append(messages, err.Error())
	}
	return fmt.Sprintf("validation failed: %s", strings.Join(messages, "; "))
}

// Errors returns the slice of errors for iteration
func (ve ValidationErrors) Errors() []ValidationError {
	return []ValidationError(ve)
}

// ErrorStrings returns error messages as strings
func (ve ValidationErrors) ErrorStrings() []string {
	var messages []string
	for _, err := range ve {
		messages = append(messages, err.Error())
	}
	return messages
}

// HasErrors returns true if there are any validation errors
func (ve ValidationErrors) HasErrors() bool {
	return len(ve) > 0
}

// Add appends a new validation error
func (ve *ValidationErrors) Add(field, message string) {
	*ve = append(*ve, ValidationError{Field: field, Message: message})
}

// AddWithValue appends a validation error with the invalid value
func (ve *ValidationErrors) AddWithValue(field, message string, value any) {
	*ve = append(*ve, ValidationError{Field: field, Message: message, Value: value})
}

// AddWithCode appends a validation error with an error code
func (ve *ValidationErrors) AddWithCode(field, message, code string) {
	*ve = append(*ve, ValidationError{Field: field, Message: message, Code: code})
}

// ConfigurationError represents errors during configuration loading
type ConfigurationError struct {
	Source  string `json:"source"`
	Message string `json:"message"`
	Cause   error  `json:"cause,omitempty"`
}

// Error implements the error interface
func (ce ConfigurationError) Error() string {
	if ce.Cause != nil {
		return fmt.Sprintf("configuration error from %s: %s (caused by: %v)", ce.Source, ce.Message, ce.Cause)
	}
	return fmt.Sprintf("configuration error from %s: %s", ce.Source, ce.Message)
}

// Unwrap returns the underlying error
func (ce ConfigurationError) Unwrap() error {
	return ce.Cause
}

// MultiError represents multiple errors from different sources
type MultiError struct {
	Errors map[string]error `json:"errors"`
}

// NewMultiError creates a new multi-error
func NewMultiError() *MultiError {
	return &MultiError{
		Errors: make(map[string]error),
	}
}

// Add adds an error for a specific source
func (me *MultiError) Add(source string, err error) {
	if me.Errors == nil {
		me.Errors = make(map[string]error)
	}
	me.Errors[source] = err
}

// HasErrors returns true if there are any errors
func (me *MultiError) HasErrors() bool {
	return len(me.Errors) > 0
}

// Error implements the error interface
func (me *MultiError) Error() string {
	if !me.HasErrors() {
		return ""
	}

	var messages []string
	for source, err := range me.Errors {
		messages = append(messages, fmt.Sprintf("%s: %v", source, err))
	}
	return fmt.Sprintf("multiple errors: %s", strings.Join(messages, "; "))
}

// Get returns the error for a specific source
func (me *MultiError) Get(source string) error {
	return me.Errors[source]
}

// Sources returns all error sources
func (me *MultiError) Sources() []string {
	sources := make([]string, 0, len(me.Errors))
	for source := range me.Errors {
		sources = append(sources, source)
	}
	return sources
}

// GetSourceErrors returns all source errors
func (me *MultiError) GetSourceErrors() map[string]error {
	return me.Errors
}

// AsValidationErrors attempts to convert an error to ValidationErrors
func AsValidationErrors(err error) (ValidationErrors, bool) {
	if validationErr, ok := err.(ValidationErrors); ok {
		return validationErr, true
	}
	return nil, false
}

// AsConfigurationError attempts to convert an error to ConfigurationError
func AsConfigurationError(err error) (ConfigurationError, bool) {
	if configErr, ok := err.(ConfigurationError); ok {
		return configErr, true
	}
	return ConfigurationError{}, false
}

// ConfigLoadError represents a comprehensive configuration loading error
type ConfigLoadError struct {
	SourceErrors    map[string]error `json:"source_errors,omitempty"`
	HydrationErrors *[]string        `json:"hydration_errors,omitempty"`
}

// NewConfigLoadError creates a new configuration load error
func NewConfigLoadError() *ConfigLoadError {
	return &ConfigLoadError{
		SourceErrors:    make(map[string]error),
		HydrationErrors: &[]string{},
	}
}

// AddSourceError adds a source-specific error
func (cle *ConfigLoadError) AddSourceError(source string, err error) {
	if cle.SourceErrors == nil {
		cle.SourceErrors = make(map[string]error)
	}
	cle.SourceErrors[source] = err
}

// AddHydrationError adds a hydration error
func (cle *ConfigLoadError) AddHydrationError(message string) {
	*cle.HydrationErrors = append(*cle.HydrationErrors, message)
}

// HasErrors returns true if there are any errors
func (cle *ConfigLoadError) HasErrors() bool {
	return len(cle.SourceErrors) > 0 || len(*cle.HydrationErrors) > 0
}

// Error implements the error interface with clear, readable messages
func (cle *ConfigLoadError) Error() string {
	if !cle.HasErrors() {
		return ""
	}

	var parts []string

	// Add source errors
	if len(cle.SourceErrors) > 0 {
		var sourceMessages []string
		for source, err := range cle.SourceErrors {
			sourceMessages = append(sourceMessages, fmt.Sprintf("%s: %v", source, err))
		}
		parts = append(parts, fmt.Sprintf("source errors: %s", strings.Join(sourceMessages, ", ")))
	}

	// Add hydration errors
	if len(*cle.HydrationErrors) > 0 {
		parts = append(parts, fmt.Sprintf("hydration errors: %s", strings.Join(*cle.HydrationErrors, ", ")))
	}

	return fmt.Sprintf("configuration loading failed: %s", strings.Join(parts, "; "))
}

// GetSuggestions returns helpful suggestions for fixing the errors
func (cle *ConfigLoadError) GetSuggestions() []string {
	var suggestions []string

	// Source error suggestions
	for source, err := range cle.SourceErrors {
		errMsg := strings.ToLower(err.Error())
		switch {
		case strings.Contains(errMsg, "not found") || strings.Contains(errMsg, "no such file"):
			suggestions = append(suggestions, fmt.Sprintf("Ensure the %s configuration file exists and is accessible", source))
		case strings.Contains(errMsg, "permission"):
			suggestions = append(suggestions, fmt.Sprintf("Check file permissions for the %s configuration source", source))
		case strings.Contains(errMsg, "parse") || strings.Contains(errMsg, "unmarshal"):
			suggestions = append(suggestions, fmt.Sprintf("Fix syntax errors in the %s configuration format", source))
		default:
			suggestions = append(suggestions, fmt.Sprintf("Review the %s configuration source", source))
		}
	}

	// Hydration error suggestions
	for _, hydrationErr := range *cle.HydrationErrors {
		msg := strings.ToLower(hydrationErr)
		switch {
		case strings.Contains(msg, "required"):
			suggestions = append(suggestions, fmt.Sprintf("Set the required '%s' field in your configuration", hydrationErr))
		case strings.Contains(msg, "format") || strings.Contains(msg, "pattern"):
			suggestions = append(suggestions, fmt.Sprintf("Check the format of the '%s' field", hydrationErr))
		case strings.Contains(msg, "range") || strings.Contains(msg, "min") || strings.Contains(msg, "max"):
			suggestions = append(suggestions, fmt.Sprintf("Ensure the '%s' field value is within the allowed range", hydrationErr))
		case strings.Contains(msg, "email"):
			suggestions = append(suggestions, fmt.Sprintf("Use a valid email format for the '%s' field", hydrationErr))
		default:
			suggestions = append(suggestions, fmt.Sprintf("Review the '%s' field configuration", hydrationErr))
		}
	}

	return suggestions
}
