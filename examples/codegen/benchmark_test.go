package main

import (
	"testing"

	goconfig "github.com/mateothegreat/go-config"
)

// Test config for benchmarking
type BenchmarkConfig struct {
	Host        string  `validate:"required,regex=^[a-zA-Z0-9.-]+$"`
	Port        int     `validate:"required,range=1000:65535"`
	Protocol    string  `validate:"oneof=http|https"`
	DatabaseURL string  `validate:"required,url"`
	MaxConns    int     `validate:"min=1,max=100"`
	JWTSecret   string  `validate:"required,minlen=32"`
	AdminEmail  string  `validate:"required,email"`
	LogLevel    string  `validate:"oneof=debug|info|warn|error"`
	Version     string  `validate:"required,regex=^v[0-9]+\\.[0-9]+\\.[0-9]+$"`
	APIKey      string  `validate:"required,len=40,alphanumeric"`
	Rate        float64 `validate:"min=0.1,max=100.0"`
}

// Zero reflection validation method (generated)
func (c *BenchmarkConfig) Validate() goconfig.ValidationErrors {
	var errors goconfig.ValidationErrors

	// Direct field access - no reflection!
	if c.Host == "" {
		errors = append(errors, "field 'Host' is required")
	}
	if c.Port == 0 {
		errors = append(errors, "field 'Port' is required")
	}
	if c.Port < 1000 || c.Port > 65535 {
		errors = append(errors, "field 'Port' must be between 1000 and 65535")
	}
	if c.Protocol != "http" && c.Protocol != "https" {
		errors = append(errors, "field 'Protocol' must be one of [http|https]")
	}
	if c.DatabaseURL == "" {
		errors = append(errors, "field 'DatabaseURL' is required")
	}
	if c.MaxConns < 1 {
		errors = append(errors, "field 'MaxConns' must be at least 1")
	}
	if c.MaxConns > 100 {
		errors = append(errors, "field 'MaxConns' must be at most 100")
	}
	if c.JWTSecret == "" {
		errors = append(errors, "field 'JWTSecret' is required")
	}
	if len(c.JWTSecret) < 32 {
		errors = append(errors, "field 'JWTSecret' must be at least 32 characters long")
	}
	if c.AdminEmail == "" {
		errors = append(errors, "field 'AdminEmail' is required")
	}
	if c.LogLevel != "" && c.LogLevel != "debug" && c.LogLevel != "info" && c.LogLevel != "warn" && c.LogLevel != "error" {
		errors = append(errors, "field 'LogLevel' must be one of [debug|info|warn|error]")
	}
	if c.Version == "" {
		errors = append(errors, "field 'Version' is required")
	}
	if c.APIKey == "" {
		errors = append(errors, "field 'APIKey' is required")
	}
	if len(c.APIKey) != 40 {
		errors = append(errors, "field 'APIKey' must be exactly 40 characters long")
	}
	if c.Rate < 0.1 {
		errors = append(errors, "field 'Rate' must be at least 0.1")
	}
	if c.Rate > 100.0 {
		errors = append(errors, "field 'Rate' must be at most 100.0")
	}

	return errors
}

func getBenchmarkConfig() *BenchmarkConfig {
	return &BenchmarkConfig{
		Host:        "localhost",
		Port:        8080,
		Protocol:    "http",
		DatabaseURL: "postgres://user:pass@localhost:5432/mydb",
		MaxConns:    100,
		JWTSecret:   "this-is-a-very-long-jwt-secret-key-that-meets-32-character-minimum",
		AdminEmail:  "admin@example.com",
		LogLevel:    "debug",
		Version:     "v1.0.0",
		APIKey:      "abcd1234efgh5678ijkl9012mnop3456qrst7890",
		Rate:        5.5,
	}
}

// Benchmark the original reflection-based validation
func BenchmarkOriginalReflectionValidation(b *testing.B) {
	config := getBenchmarkConfig()
	validator := goconfig.NewStructValidator()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = validator.Validate(config)
	}
}

// Benchmark the fast struct validator (minimal reflection)
func BenchmarkFastStructValidation(b *testing.B) {
	config := getBenchmarkConfig()
	validator := goconfig.NewFastStructValidator()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = validator.Validate(config)
	}
}

// Benchmark the zero reflection validation (generated code)
func BenchmarkZeroReflectionValidation(b *testing.B) {
	config := getBenchmarkConfig()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = config.Validate()
	}
}

// Test that all validation approaches produce the same results
func TestValidationConsistency(t *testing.T) {
	config := getBenchmarkConfig()

	// Original reflection validator
	originalValidator := goconfig.NewStructValidator()
	originalErr := originalValidator.Validate(config)

	// Fast struct validator
	fastValidator := goconfig.NewFastStructValidator()
	fastErr := fastValidator.Validate(config)

	// Zero reflection validator
	zeroReflectionErrors := config.Validate()

	// All should succeed for valid config
	if originalErr != nil {
		t.Errorf("Original validator failed: %v", originalErr)
	}
	if fastErr != nil {
		t.Errorf("Fast validator failed: %v", fastErr)
	}
	if zeroReflectionErrors.HasErrors() {
		t.Errorf("Zero reflection validator failed: %v", zeroReflectionErrors)
	}

	// Test with invalid config
	invalidConfig := &BenchmarkConfig{
		Host: "",  // Required field missing
		Port: 999, // Below minimum
	}

	originalErr = originalValidator.Validate(invalidConfig)
	fastErr = fastValidator.Validate(invalidConfig)
	zeroReflectionErrors = invalidConfig.Validate()

	// All should fail for invalid config
	if originalErr == nil {
		t.Error("Original validator should have failed")
	}
	if fastErr == nil {
		t.Error("Fast validator should have failed")
	}
	if !zeroReflectionErrors.HasErrors() {
		t.Error("Zero reflection validator should have failed")
	}
}
