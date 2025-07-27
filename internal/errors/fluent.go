package errors

import (
	"fmt"
	"strings"
)

// ValidationError represents a single validation error with rich context
type ValidationError struct {
	Field   string
	Value   interface{}
	Tag     string
	Message string
	Code    string
}

// Error implements the error interface
func (ve ValidationError) Error() string {
	if ve.Message != "" {
		return ve.Message
	}
	return fmt.Sprintf("validation failed for field '%s'", ve.Field)
}

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

// Field adds a field validation error
func (fe *FluentError) Field(fieldName string) *FieldErrorBuilder {
	return &FieldErrorBuilder{
		parent:    fe,
		fieldName: fieldName,
	}
}

// Add adds a validation error directly
func (fe *FluentError) Add(err ValidationError) *FluentError {
	fe.errors = append(fe.errors, err)
	return fe
}

// HasErrors returns true if there are any validation errors
func (fe *FluentError) HasErrors() bool {
	return len(fe.errors) > 0
}

// Errors returns all validation errors
func (fe *FluentError) Errors() []ValidationError {
	return fe.errors
}

// Error implements the error interface
func (fe *FluentError) Error() string {
	if len(fe.errors) == 0 {
		return ""
	}
	
	if len(fe.errors) == 1 {
		return fe.errors[0].Error()
	}
	
	var messages []string
	for _, err := range fe.errors {
		messages = append(messages, err.Error())
	}
	return fmt.Sprintf("validation failed: %s", strings.Join(messages, "; "))
}

// FieldErrorBuilder provides a fluent interface for building field-specific errors
type FieldErrorBuilder struct {
	parent    *FluentError
	fieldName string
	value     interface{}
	tag       string
	code      string
}

// Value sets the field value for context
func (feb *FieldErrorBuilder) Value(value interface{}) *FieldErrorBuilder {
	feb.value = value
	return feb
}

// Tag sets the validation tag that failed
func (feb *FieldErrorBuilder) Tag(tag string) *FieldErrorBuilder {
	feb.tag = tag
	return feb
}

// Code sets the error code
func (feb *FieldErrorBuilder) Code(code string) *FieldErrorBuilder {
	feb.code = code
	return feb
}

// Format sets a custom error message and returns to the parent FluentError
func (feb *FieldErrorBuilder) Format(message string, args ...interface{}) *FluentError {
	err := ValidationError{
		Field:   feb.fieldName,
		Value:   feb.value,
		Tag:     feb.tag,
		Code:    feb.code,
		Message: fmt.Sprintf(message, args...),
	}
	feb.parent.errors = append(feb.parent.errors, err)
	return feb.parent
}

// Required creates a "required field" error
func (feb *FieldErrorBuilder) Required() *FluentError {
	return feb.Format("field '%s' is required", feb.fieldName)
}

// MinLength creates a minimum length validation error
func (feb *FieldErrorBuilder) MinLength(min int) *FluentError {
	return feb.Format("field '%s' must be at least %d characters long", feb.fieldName, min)
}

// MaxLength creates a maximum length validation error
func (feb *FieldErrorBuilder) MaxLength(max int) *FluentError {
	return feb.Format("field '%s' must be at most %d characters long", feb.fieldName, max)
}

// Range creates a range validation error
func (feb *FieldErrorBuilder) Range(min, max interface{}) *FluentError {
	return feb.Format("field '%s' must be between %v and %v", feb.fieldName, min, max)
}

// MultiError represents multiple validation errors
type MultiError struct {
	errors map[string]error
}

// NewMultiError creates a new multi-error from a map of field errors
func NewMultiError(fieldErrors map[string]error) *MultiError {
	return &MultiError{
		errors: fieldErrors,
	}
}

// Error implements the error interface
func (me *MultiError) Error() string {
	if len(me.errors) == 0 {
		return ""
	}
	
	var messages []string
	for field, err := range me.errors {
		messages = append(messages, fmt.Sprintf("%s: %s", field, err.Error()))
	}
	return fmt.Sprintf("validation failed: %s", strings.Join(messages, "; "))
}

// Errors returns the map of field errors
func (me *MultiError) Errors() map[string]error {
	return me.errors
}

// HasErrors returns true if there are any errors
func (me *MultiError) HasErrors() bool {
	return len(me.errors) > 0
}

// Add adds an error for a specific field
func (me *MultiError) Add(field string, err error) {
	if me.errors == nil {
		me.errors = make(map[string]error)
	}
	me.errors[field] = err
}

// Get returns the error for a specific field
func (me *MultiError) Get(field string) error {
	return me.errors[field]
}