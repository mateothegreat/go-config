package errors

import "fmt"

// FluentError provides a fluent interface for building validation errors
type FluentError struct {
	errors []ValidationError
}

// NewError creates a new fluent error builder
func NewError() *FluentError {
	return &FluentError{
		errors: make([]ValidationError, 0),
	}
}

// Field sets the field name for the next error
func (fe *FluentError) Field(name string) *FieldError {
	return &FieldError{
		builder:   fe,
		fieldName: name,
	}
}

// HasErrors returns true if any errors have been added
func (fe *FluentError) HasErrors() bool {
	return len(fe.errors) > 0
}

// Errors returns the collected validation errors
func (fe *FluentError) Errors() []ValidationError {
	return fe.errors
}

// ValidationErrors returns the errors as ValidationErrors type
func (fe *FluentError) ValidationErrors() ValidationErrors {
	return ValidationErrors(fe.errors)
}

// Error implements the error interface
func (fe *FluentError) Error() string {
	return ValidationErrors(fe.errors).Error()
}

// FieldError provides fluent methods for building field-specific errors
type FieldError struct {
	builder   *FluentError
	fieldName string
}

// Required adds a required field error
func (fie *FieldError) Required() *FluentError {
	fie.builder.errors = append(fie.builder.errors, ValidationError{
		Field:   fie.fieldName,
		Message: "is required",
		Code:    "required",
	})
	return fie.builder
}

// Format adds a custom format error
func (fie *FieldError) Format(message string) *FluentError {
	fie.builder.errors = append(fie.builder.errors, ValidationError{
		Field:   fie.fieldName,
		Message: message,
		Code:    "format",
	})
	return fie.builder
}

// MinLength adds a minimum length error
func (fie *FieldError) MinLength(min int) *FluentError {
	fie.builder.errors = append(fie.builder.errors, ValidationError{
		Field:   fie.fieldName,
		Message: fmt.Sprintf("must be at least %d characters long", min),
		Code:    "minlen",
	})
	return fie.builder
}

// MaxLength adds a maximum length error
func (fie *FieldError) MaxLength(max int) *FluentError {
	fie.builder.errors = append(fie.builder.errors, ValidationError{
		Field:   fie.fieldName,
		Message: fmt.Sprintf("must be at most %d characters long", max),
		Code:    "maxlen",
	})
	return fie.builder
}

// Min adds a minimum value error
func (fie *FieldError) Min(min interface{}) *FluentError {
	fie.builder.errors = append(fie.builder.errors, ValidationError{
		Field:   fie.fieldName,
		Message: fmt.Sprintf("must be at least %v", min),
		Code:    "min",
	})
	return fie.builder
}

// Max adds a maximum value error
func (fie *FieldError) Max(max interface{}) *FluentError {
	fie.builder.errors = append(fie.builder.errors, ValidationError{
		Field:   fie.fieldName,
		Message: fmt.Sprintf("must be at most %v", max),
		Code:    "max",
	})
	return fie.builder
}

// Range adds a range validation error
func (fie *FieldError) Range(min, max interface{}) *FluentError {
	fie.builder.errors = append(fie.builder.errors, ValidationError{
		Field:   fie.fieldName,
		Message: fmt.Sprintf("must be between %v and %v", min, max),
		Code:    "range",
	})
	return fie.builder
}

// Email adds an email validation error
func (fie *FieldError) Email() *FluentError {
	fie.builder.errors = append(fie.builder.errors, ValidationError{
		Field:   fie.fieldName,
		Message: "must be a valid email address",
		Code:    "email",
	})
	return fie.builder
}

// URL adds a URL validation error
func (fie *FieldError) URL() *FluentError {
	fie.builder.errors = append(fie.builder.errors, ValidationError{
		Field:   fie.fieldName,
		Message: "must be a valid URL",
		Code:    "url",
	})
	return fie.builder
}

// OneOf adds a one-of validation error
func (fie *FieldError) OneOf(options ...string) *FluentError {
	fie.builder.errors = append(fie.builder.errors, ValidationError{
		Field:   fie.fieldName,
		Message: fmt.Sprintf("must be one of: %v", options),
		Code:    "oneof",
	})
	return fie.builder
}

// Custom adds a custom validation error with code
func (fie *FieldError) Custom(message, code string) *FluentError {
	fie.builder.errors = append(fie.builder.errors, ValidationError{
		Field:   fie.fieldName,
		Message: message,
		Code:    code,
	})
	return fie.builder
}

// WithValue adds the invalid value to the error
func (fie *FieldError) WithValue(value interface{}) *FieldError {
	if len(fie.builder.errors) > 0 {
		lastIdx := len(fie.builder.errors) - 1
		fie.builder.errors[lastIdx].Value = value
	}
	return fie
}
