package goconfig

import (
	"testing"
)

// BenchmarkConfig for testing validation performance
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

func BenchmarkOriginalStructValidator(b *testing.B) {
	config := getBenchmarkConfig()
	validator := NewStructValidator()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = validator.Validate(config)
	}
}

func BenchmarkFastStructValidator(b *testing.B) {
	config := getBenchmarkConfig()
	validator := NewFastStructValidator()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = validator.Validate(config)
	}
}

func BenchmarkTypedValidation(b *testing.B) {
	config := getBenchmarkConfig()
	fastValidator := NewFastValidator()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Direct type-specific validation (no reflection for values)
		rules := map[string]string{"required": "", "regex": "^[a-zA-Z0-9.-]+$"}
		_ = fastValidator.ValidateString("Host", config.Host, rules)
		
		rules = map[string]string{"required": "", "range": "1000:65535"}
		_ = fastValidator.ValidateInt("Port", config.Port, rules)
		
		rules = map[string]string{"oneof": "http|https"}
		_ = fastValidator.ValidateString("Protocol", config.Protocol, rules)
		
		rules = map[string]string{"required": "", "url": ""}
		_ = fastValidator.ValidateString("DatabaseURL", config.DatabaseURL, rules)
		
		rules = map[string]string{"min": "1", "max": "100"}
		_ = fastValidator.ValidateInt("MaxConns", config.MaxConns, rules)
		
		rules = map[string]string{"required": "", "minlen": "32"}
		_ = fastValidator.ValidateString("JWTSecret", config.JWTSecret, rules)
		
		rules = map[string]string{"required": "", "email": ""}
		_ = fastValidator.ValidateString("AdminEmail", config.AdminEmail, rules)
		
		rules = map[string]string{"oneof": "debug|info|warn|error"}
		_ = fastValidator.ValidateString("LogLevel", config.LogLevel, rules)
		
		rules = map[string]string{"required": "", "regex": "^v[0-9]+\\.[0-9]+\\.[0-9]+$"}
		_ = fastValidator.ValidateString("Version", config.Version, rules)
		
		rules = map[string]string{"required": "", "len": "40", "alphanumeric": ""}
		_ = fastValidator.ValidateString("APIKey", config.APIKey, rules)
		
		rules = map[string]string{"min": "0.1", "max": "100.0"}
		_ = fastValidator.ValidateFloat("Rate", config.Rate, rules)
	}
}