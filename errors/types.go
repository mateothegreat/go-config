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
