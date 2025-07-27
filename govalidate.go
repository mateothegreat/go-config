// Package goconfig provides fast, fluent, and flexible validation for Go configs.
//
// go-validate is a powerful code generation tool built on three foundational pillars:
// - ⚡ Performance: Generates zero-reflection, allocation-conscious code
// - ✨ Ergonomics: Offers intuitive interfaces and seamless config integration
// - 🔗 Fluency: Enables chainable APIs for expressive validation and error handling
//
// Unlike typical reflection-based libraries, go-validate generates optimized validation
// code directly from your Go structs—ensuring lightning-fast execution and cold-start performance.
package goconfig

import (
	"github.com/mateothegreat/go-config/internal/errors"
	"github.com/mateothegreat/go-config/internal/generator"
	"github.com/mateothegreat/go-config/internal/plugins"
	"github.com/mateothegreat/go-config/internal/scanner"
)

// Core types re-exported for convenience

// ValidationError represents a single validation error with rich context
type ValidationError = errors.ValidationError

// FluentError provides a fluent interface for building validation errors
type FluentError = errors.FluentError

// MultiError represents multiple validation errors
type MultiError = errors.MultiError

// Generator options
type GeneratorOption = generator.GeneratorOption

// Generator configuration
type GeneratorConfig = generator.GeneratorConfig

// Struct and field information
type StructInfo = scanner.StructInfo
type FieldInfo = scanner.FieldInfo

// Plugin system
type ValidatorPlugin = plugins.ValidatorPlugin
type PluginMetadata = plugins.PluginMetadata

// Core functions

// NewGenerator creates a new code generator with the given options
func NewGenerator(options ...GeneratorOption) *generator.Generator {
	return generator.NewGenerator(options...)
}

// Generator option functions
var (
	WithInputDir  = generator.WithInputDir
	WithOutputDir = generator.WithOutputDir
	WithPackage   = generator.WithPackage
	WithStructs   = generator.WithStructs
	WithDryRun    = generator.WithDryRun
	WithMulti     = generator.WithMulti
	WithCache     = generator.WithCache
	WithVerbose   = generator.WithVerbose
)

// Scanner functions

// NewASTScanner creates a new AST scanner
func NewASTScanner(verbose bool) *scanner.ASTScanner {
	return scanner.NewASTScanner(verbose)
}

// Plugin functions

// RegisterValidator registers a custom validator plugin
func RegisterValidator(name string, plugin ValidatorPlugin) {
	plugins.RegisterValidator(name, plugin)
}

// RegisterValidatorWithMetadata registers a plugin with metadata
func RegisterValidatorWithMetadata(name string, plugin ValidatorPlugin, metadata PluginMetadata) {
	plugins.RegisterValidatorWithMetadata(name, plugin, metadata)
}

// GetValidator retrieves a validator plugin by name
func GetValidator(name string) (ValidatorPlugin, bool) {
	return plugins.GetValidator(name)
}

// ListValidators returns all registered validator names
func ListValidators() []string {
	return plugins.ListValidators()
}

// Error building functions

// NewValidationError creates a new validation error
func NewValidationError(field, message string) ValidationError {
	return ValidationError{
		Field:   field,
		Message: message,
	}
}

// AsValidationErrors attempts to convert an error to ValidationErrors
func AsValidationErrors(err error) (ValidationErrors, bool) {
	if validationErr, ok := err.(ValidationErrors); ok {
		return validationErr, true
	}
	return nil, false
}

// Convenience constants for common validation rules
const (
	Required     = "required"
	Min          = "min"
	Max          = "max"
	MinLen       = "minlen"
	MaxLen       = "maxlen"
	Len          = "len"
	Email        = "email"
	URL          = "url"
	Alpha        = "alpha"
	AlphaNumeric = "alphanumeric"
	Numeric      = "numeric"
	Regex        = "regex"
	OneOf        = "oneof"
	Range        = "range"
	IP           = "ip"
	UUID         = "uuid"
	CreditCard   = "creditcard"
	Phone        = "phone"
)