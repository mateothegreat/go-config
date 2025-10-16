package validation

import (
	"fmt"
	"strings"
)

// ValidationErrors represents validation errors using bit fields for zero-allocation tracking
type ValidationErrors uint64

// Error constants using bit fields for maximum performance
const (
	ErrNone ValidationErrors = 0

	// Common validation errors
	ErrRequired ValidationErrors = 1 << iota
	ErrEmail
	ErrURL
	ErrMinLength
	ErrMaxLength
	ErrLength
	ErrMin
	ErrMax
	ErrAlpha
	ErrAlphaNum
	ErrNumeric
	ErrBase64
	ErrUUID
	ErrIP
	ErrIPv4
	ErrIPv6
	ErrHostname
	ErrCIDR
	ErrMAC
	ErrOneOf

	// Additional validation error types
	ErrConditional
	ErrCrossfieldValidation
	ErrCustom
	ErrFormat
	ErrRange
)

// ErrorDetails contains human-readable error information
type ErrorDetails struct {
	Code    ValidationErrors
	Field   string
	Message string
	Value   interface{}
}

// Error messages for each validation type
var errorMessages = map[ValidationErrors]string{
	ErrRequired:   "field is required",
	ErrEmail:      "must be a valid email address",
	ErrURL:        "must be a valid URL",
	ErrMinLength:  "length is too short",
	ErrMaxLength:  "length is too long",
	ErrLength:     "length is incorrect",
	ErrMin:        "value is too small",
	ErrMax:        "value is too large",
	ErrAlpha:      "must contain only letters",
	ErrAlphaNum:   "must contain only letters and numbers",
	ErrNumeric:    "must contain only numbers",
	ErrBase64:     "must be valid base64",
	ErrUUID:       "must be a valid UUID",
	ErrIP:         "must be a valid IP address",
	ErrIPv4:       "must be a valid IPv4 address",
	ErrIPv6:       "must be a valid IPv6 address",
	ErrHostname:   "must be a valid hostname",
	ErrCIDR:       "must be a valid CIDR notation",
	ErrMAC:        "must be a valid MAC address",
	ErrOneOf:      "must be one of the allowed values",
}

// HasErrors returns true if any validation errors are present
func (e ValidationErrors) HasErrors() bool {
	return e != ErrNone
}

// HasError returns true if the specific error is present
func (e ValidationErrors) HasError(err ValidationErrors) bool {
	return e&err != 0
}

// Add adds an error to the set
func (e *ValidationErrors) Add(err ValidationErrors) {
	*e |= err
}

// Remove removes an error from the set
func (e *ValidationErrors) Remove(err ValidationErrors) {
	*e &^= err
}

// Count returns the number of errors present
func (e ValidationErrors) Count() int {
	count := 0
	for i := uint(0); i < 64; i++ {
		if e&(1<<i) != 0 {
			count++
		}
	}
	return count
}

// Error implements the error interface
func (e ValidationErrors) Error() string {
	if e == ErrNone {
		return ""
	}

	var errors []string
	for code, message := range errorMessages {
		if e&code != 0 {
			errors = append(errors, message)
		}
	}

	if len(errors) == 1 {
		return errors[0]
	}
	return fmt.Sprintf("validation failed: %s", strings.Join(errors, ", "))
}

// String returns a string representation of the errors
func (e ValidationErrors) String() string {
	return e.Error()
}

// ErrorList returns a slice of individual errors
func (e ValidationErrors) ErrorList() []ValidationErrors {
	var errors []ValidationErrors
	for i := uint(0); i < 64; i++ {
		code := ValidationErrors(1 << i)
		if e&code != 0 {
			errors = append(errors, code)
		}
	}
	return errors
}

// ValidationResult contains validation results with context
type ValidationResult struct {
	Errors ValidationErrors
	Fields map[string]ValidationErrors
}

// NewValidationResult creates a new validation result
func NewValidationResult() *ValidationResult {
	return &ValidationResult{
		Errors: ErrNone,
		Fields: make(map[string]ValidationErrors),
	}
}

// AddFieldError adds an error for a specific field
func (r *ValidationResult) AddFieldError(field string, err ValidationErrors) {
	r.Errors |= err
	if r.Fields[field] == 0 {
		r.Fields[field] = err
	} else {
		r.Fields[field] |= err
	}
}

// HasErrors returns true if any validation errors are present
func (r *ValidationResult) HasErrors() bool {
	return r.Errors.HasErrors()
}

// Error implements the error interface
func (r *ValidationResult) Error() string {
	if !r.HasErrors() {
		return ""
	}

	var fieldErrors []string
	for field, errs := range r.Fields {
		if errs.HasErrors() {
			fieldErrors = append(fieldErrors, fmt.Sprintf("%s: %s", field, errs.Error()))
		}
	}

	if len(fieldErrors) == 0 {
		return r.Errors.Error()
	}

	return strings.Join(fieldErrors, "; ")
}

// GetFieldErrors returns errors for a specific field
func (r *ValidationResult) GetFieldErrors(field string) ValidationErrors {
	return r.Fields[field]
}