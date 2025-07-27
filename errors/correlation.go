package errors

import (
	"fmt"
	"strings"
)

// ErrorCorrelator provides enhanced error correlation between configuration and validation
type ErrorCorrelator struct {
	sourceErrors     map[string]error
	validationErrors ValidationErrors
	correlations     []ErrorCorrelation
}

// ErrorCorrelation represents a correlation between source and validation errors
type ErrorCorrelation struct {
	SourceName      string `json:"source_name"`
	FieldName       string `json:"field_name"`
	SourceError     string `json:"source_error,omitempty"`
	ValidationError string `json:"validation_error,omitempty"`
	Suggestion      string `json:"suggestion,omitempty"`
	Category        string `json:"category"`
}

// NewErrorCorrelator creates a new error correlator
func NewErrorCorrelator() *ErrorCorrelator {
	return &ErrorCorrelator{
		sourceErrors:     make(map[string]error),
		validationErrors: make(ValidationErrors, 0),
		correlations:     make([]ErrorCorrelation, 0),
	}
}

// AddSourceError adds a source-specific error
func (ec *ErrorCorrelator) AddSourceError(source string, err error) {
	ec.sourceErrors[source] = err
}

// AddValidationErrors adds validation errors
func (ec *ErrorCorrelator) AddValidationErrors(errors ValidationErrors) {
	ec.validationErrors = append(ec.validationErrors, errors...)
}

// Correlate analyzes and correlates errors to provide better diagnostics
func (ec *ErrorCorrelator) Correlate() []ErrorCorrelation {
	ec.correlations = make([]ErrorCorrelation, 0)

	// Correlate validation errors with potential source issues
	for _, validationErr := range ec.validationErrors {
		correlation := ec.analyzeValidationError(validationErr)
		if correlation != nil {
			ec.correlations = append(ec.correlations, *correlation)
		}
	}

	// Add source-only errors
	for source, err := range ec.sourceErrors {
		correlation := ErrorCorrelation{
			SourceName:  source,
			SourceError: err.Error(),
			Category:    "source_error",
			Suggestion:  ec.generateSourceSuggestion(source, err),
		}
		ec.correlations = append(ec.correlations, correlation)
	}

	return ec.correlations
}

// analyzeValidationError analyzes a validation error and creates correlations
func (ec *ErrorCorrelator) analyzeValidationError(validationErr ValidationError) *ErrorCorrelation {
	correlation := &ErrorCorrelation{
		FieldName:       validationErr.Field,
		ValidationError: validationErr.Message,
		Category:        ec.categorizeValidationError(validationErr),
	}

	// Try to find related source errors
	for source, sourceErr := range ec.sourceErrors {
		if ec.isRelated(validationErr, sourceErr) {
			correlation.SourceName = source
			correlation.SourceError = sourceErr.Error()
			break
		}
	}

	correlation.Suggestion = ec.generateValidationSuggestion(validationErr, correlation.SourceName)
	return correlation
}

// categorizeValidationError categorizes validation errors
func (ec *ErrorCorrelator) categorizeValidationError(validationErr ValidationError) string {
	message := strings.ToLower(validationErr.Message)

	if strings.Contains(message, "required") {
		return "missing_required"
	}
	if strings.Contains(message, "format") || strings.Contains(message, "pattern") {
		return "format_error"
	}
	if strings.Contains(message, "range") || strings.Contains(message, "min") || strings.Contains(message, "max") {
		return "range_error"
	}
	if strings.Contains(message, "length") {
		return "length_error"
	}
	if strings.Contains(message, "email") {
		return "email_format"
	}
	if strings.Contains(message, "url") {
		return "url_format"
	}

	return "validation_error"
}

// isRelated checks if a validation error is related to a source error
func (ec *ErrorCorrelator) isRelated(validationErr ValidationError, sourceErr error) bool {
	sourceMsg := strings.ToLower(sourceErr.Error())
	fieldName := strings.ToLower(validationErr.Field)

	// Check if source error mentions the field
	if strings.Contains(sourceMsg, fieldName) {
		return true
	}

	// Check for common patterns
	if strings.Contains(sourceMsg, "not found") && strings.Contains(validationErr.Message, "required") {
		return true
	}

	if strings.Contains(sourceMsg, "parse") && strings.Contains(validationErr.Message, "format") {
		return true
	}

	return false
}

// generateValidationSuggestion generates suggestions for validation errors
func (ec *ErrorCorrelator) generateValidationSuggestion(validationErr ValidationError, sourceName string) string {
	category := ec.categorizeValidationError(validationErr)
	field := validationErr.Field

	switch category {
	case "missing_required":
		if sourceName != "" {
			return fmt.Sprintf("Set the '%s' field in your %s configuration source", field, sourceName)
		}
		return fmt.Sprintf("The '%s' field is required. Add it to your configuration", field)

	case "format_error":
		return fmt.Sprintf("Check the format of the '%s' field. Ensure it matches the expected pattern", field)

	case "range_error":
		return fmt.Sprintf("The '%s' field value is out of range. Check the min/max constraints", field)

	case "email_format":
		return fmt.Sprintf("The '%s' field should be a valid email address (e.g., user@example.com)", field)

	case "url_format":
		return fmt.Sprintf("The '%s' field should be a valid URL (e.g., https://example.com)", field)

	case "length_error":
		return fmt.Sprintf("The '%s' field length is invalid. Check the length constraints", field)

	default:
		return fmt.Sprintf("Review the '%s' field configuration and ensure it meets validation requirements", field)
	}
}

// generateSourceSuggestion generates suggestions for source errors
func (ec *ErrorCorrelator) generateSourceSuggestion(source string, err error) string {
	errMsg := strings.ToLower(err.Error())

	switch {
	case strings.Contains(errMsg, "not found") || strings.Contains(errMsg, "no such file"):
		return fmt.Sprintf("Ensure the %s configuration file exists and is accessible", source)

	case strings.Contains(errMsg, "permission"):
		return fmt.Sprintf("Check file permissions for the %s configuration source", source)

	case strings.Contains(errMsg, "parse") || strings.Contains(errMsg, "unmarshal"):
		return fmt.Sprintf("Fix syntax errors in the %s configuration format", source)

	case strings.Contains(errMsg, "connection") || strings.Contains(errMsg, "network"):
		return fmt.Sprintf("Check network connectivity for the %s configuration source", source)

	default:
		return fmt.Sprintf("Review the %s configuration source for issues", source)
	}
}

// GetCorrelationSummary returns a human-readable summary of correlations
func (ec *ErrorCorrelator) GetCorrelationSummary() string {
	if len(ec.correlations) == 0 {
		return "No errors found"
	}

	var parts []string
	categories := make(map[string]int)

	for _, correlation := range ec.correlations {
		categories[correlation.Category]++
	}

	for category, count := range categories {
		if count == 1 {
			parts = append(parts, fmt.Sprintf("1 %s", strings.ReplaceAll(category, "_", " ")))
		} else {
			parts = append(parts, fmt.Sprintf("%d %s errors", count, strings.ReplaceAll(category, "_", " ")))
		}
	}

	return fmt.Sprintf("Found %s", strings.Join(parts, ", "))
}

// CorrelatedError combines multiple error types with correlations
type CorrelatedError struct {
	Correlations []ErrorCorrelation `json:"correlations"`
	Summary      string             `json:"summary"`
}

// NewCorrelatedError creates a new correlated error
func NewCorrelatedError(correlations []ErrorCorrelation, summary string) *CorrelatedError {
	return &CorrelatedError{
		Correlations: correlations,
		Summary:      summary,
	}
}

// Error implements the error interface
func (ce *CorrelatedError) Error() string {
	if len(ce.Correlations) == 0 {
		return "configuration error"
	}

	var messages []string
	for _, correlation := range ce.Correlations {
		if correlation.ValidationError != "" && correlation.SourceError != "" {
			messages = append(messages, fmt.Sprintf("%s: %s (from %s: %s)",
				correlation.FieldName, correlation.ValidationError,
				correlation.SourceName, correlation.SourceError))
		} else if correlation.ValidationError != "" {
			messages = append(messages, fmt.Sprintf("%s: %s",
				correlation.FieldName, correlation.ValidationError))
		} else if correlation.SourceError != "" {
			messages = append(messages, fmt.Sprintf("%s: %s",
				correlation.SourceName, correlation.SourceError))
		}
	}

	return strings.Join(messages, "; ")
}

// GetSuggestions returns all suggestions from correlations
func (ce *CorrelatedError) GetSuggestions() []string {
	var suggestions []string
	for _, correlation := range ce.Correlations {
		if correlation.Suggestion != "" {
			suggestions = append(suggestions, correlation.Suggestion)
		}
	}
	return suggestions
}
